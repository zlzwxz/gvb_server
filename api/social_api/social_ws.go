package social_api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type socialSocketRequest struct {
	Action       string          `json:"action"`
	TargetUserID uint            `json:"target_user_id"`
	CallID       string          `json:"call_id"`
	Mode         string          `json:"mode"`
	StatusText   string          `json:"status_text"`
	IsInvisible  bool            `json:"is_invisible"`
	Payload      json.RawMessage `json:"payload"`
}

type socialSocketEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

var (
	socialConnMu  sync.RWMutex
	socialConnMap = map[uint]map[string]*websocket.Conn{}
)

func registerSocialConn(userID uint, conn *websocket.Conn) string {
	socialConnMu.Lock()
	defer socialConnMu.Unlock()
	connID := time.Now().Format("20060102150405.000000000")
	if _, ok := socialConnMap[userID]; !ok {
		socialConnMap[userID] = map[string]*websocket.Conn{}
	}
	socialConnMap[userID][connID] = conn
	return connID
}

func unregisterSocialConn(userID uint, connID string) bool {
	socialConnMu.Lock()
	defer socialConnMu.Unlock()
	userConnMap, ok := socialConnMap[userID]
	if !ok {
		return false
	}
	delete(userConnMap, connID)
	if len(userConnMap) == 0 {
		delete(socialConnMap, userID)
		return true
	}
	return false
}

func snapshotOnlinePresence() map[uint]bool {
	socialConnMu.RLock()
	defer socialConnMu.RUnlock()
	result := map[uint]bool{}
	for userID, connMap := range socialConnMap {
		result[userID] = len(connMap) > 0
	}
	return result
}

func sendSocketEvent(userID uint, event string, data any) {
	byteData, _ := json.Marshal(socialSocketEvent{
		Event: event,
		Data:  data,
	})

	socialConnMu.RLock()
	connMap := socialConnMap[userID]
	connections := make([]*websocket.Conn, 0, len(connMap))
	for _, conn := range connMap {
		connections = append(connections, conn)
	}
	socialConnMu.RUnlock()

	for _, conn := range connections {
		if conn == nil {
			continue
		}
		_ = conn.WriteMessage(websocket.TextMessage, byteData)
	}
}

func broadcastSocketEvent(userIDs []uint, event string, data any) {
	for _, userID := range dedupeUintSlice(userIDs) {
		sendSocketEvent(userID, event, data)
	}
}

func notifyPresenceChange(userID uint) {
	friendIDs := fetchFriendIDs(userID)
	presence := loadPresenceMap([]uint{userID}, 0)[userID]
	broadcastSocketEvent(friendIDs, "presence_change", map[string]any{
		"user_id":        userID,
		"is_online":      presence.IsOnline,
		"presence_mode":  presence.Mode,
		"presence_text":  presence.StatusText,
		"is_invisible":   presence.IsInvisible,
		"last_active_at": presence.LastActiveAt,
	})
	sendSocketEvent(userID, "self_presence", presence)
}

func forwardCallEvent(senderID uint, req socialSocketRequest) {
	if req.TargetUserID == 0 || req.TargetUserID == senderID {
		sendSocketEvent(senderID, "socket_error", map[string]any{"message": "目标用户无效"})
		return
	}
	relation := buildRelation(senderID, req.TargetUserID)
	if !relation.CanCall {
		sendSocketEvent(senderID, "socket_error", map[string]any{"message": "仅好友之间支持语音通话"})
		return
	}
	sendSocketEvent(req.TargetUserID, req.Action, map[string]any{
		"from_user_id": senderID,
		"call_id":      strings.TrimSpace(req.CallID),
		"payload":      req.Payload,
	})
}

// SocketView 建立好友系统 websocket 连接。
func (SocialApi) SocketView(c *gin.Context) {
	claims, err := authenticateSocket(c)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				return true
			}
			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}
			if strings.EqualFold(originURL.Host, r.Host) {
				return true
			}
			host := strings.Split(r.Host, ":")[0]
			return originURL.Hostname() == "localhost" || originURL.Hostname() == "127.0.0.1" || originURL.Hostname() == host
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		res.FailWithMessage("websocket 连接失败", c)
		return
	}

	connID := registerSocialConn(claims.UserID, conn)
	touchPresence(claims.UserID)
	notifyPresenceChange(claims.UserID)
	sendSocketEvent(claims.UserID, "socket_ready", map[string]any{"user_id": claims.UserID})

	defer func() {
		offline := unregisterSocialConn(claims.UserID, connID)
		if offline {
			now := time.Now()
			global.DB.Model(&models.UserPresenceModel{}).Where("user_id = ?", claims.UserID).Updates(map[string]any{
				"last_active_at": &now,
			})
			notifyPresenceChange(claims.UserID)
		}
		_ = conn.Close()
	}()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var request socialSocketRequest
		if err = json.Unmarshal(payload, &request); err != nil {
			sendSocketEvent(claims.UserID, "socket_error", map[string]any{"message": "消息格式错误"})
			continue
		}
		switch strings.TrimSpace(request.Action) {
		case "ping":
			sendSocketEvent(claims.UserID, "pong", map[string]any{"time": time.Now()})
		case "update_presence":
			presence := savePresence(claims.UserID, socialPresenceRequest{
				Mode:        request.Mode,
				StatusText:  request.StatusText,
				IsInvisible: request.IsInvisible,
			})
			sendSocketEvent(claims.UserID, "self_presence", presence)
			notifyPresenceChange(claims.UserID)
		case "call_invite":
			handleCallInvite(claims.UserID, request)
		case "call_accept":
			handleCallAccept(claims.UserID, request)
		case "call_reject":
			handleCallReject(claims.UserID, request)
		case "call_end":
			handleCallEnd(claims.UserID, request)
		case "webrtc_offer", "webrtc_answer", "webrtc_candidate":
			forwardCallEvent(claims.UserID, request)
		default:
			sendSocketEvent(claims.UserID, "socket_error", map[string]any{"message": "未知动作"})
		}
	}
}
