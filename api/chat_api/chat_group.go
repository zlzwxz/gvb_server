package chat_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/utils"

	"github.com/DanPlayer/randomname"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ChatUser 表示当前在线聊天室用户。
type ChatUser struct {
	Conn     *websocket.Conn
	NickName string `json:"nick_name" swag:"description:聊天室昵称"`
	Avatar   string `json:"avatar" swag:"description:聊天室头像"`
}

var (
	connGroupMap = map[string]ChatUser{}
	connGroupMu  sync.RWMutex
)

// 消息类型常量定义。
const (
	TextMsg    ctype.MsgType = 1 // 文本消息
	ImageMsg   ctype.MsgType = 2 // 图片消息
	SystemMsg  ctype.MsgType = 3 // 系统提示
	InRoomMsg  ctype.MsgType = 0 // 进入聊天室事件
	OutRoomMsg ctype.MsgType = 5 // 离开聊天室事件
)

// GroupRequest 表示前端通过 websocket 发来的聊天消息。
type GroupRequest struct {
	Content string        `json:"content" swag:"description:聊天内容"`
	MsgType ctype.MsgType `json:"msg_type" swag:"description:消息类型"`
}

// GroupResponse 表示服务端广播给聊天室客户端的消息体。
type GroupResponse struct {
	NickName    string        `json:"nick_name" swag:"description:前端展示昵称"`
	Avatar      string        `json:"avatar" swag:"description:头像地址"`
	MsgType     ctype.MsgType `json:"msg_type" swag:"description:消息类型"`
	Content     string        `json:"content" swag:"description:消息内容"`
	Date        time.Time     `json:"date" swag:"description:消息时间"`
	OnlineCount int           `json:"online_count" swag:"description:当前在线人数"`
}

// ChatGroupView 建立聊天室 websocket 连接。
// @Summary 进入聊天室
// @Description 建立 websocket 连接并加入公共聊天室，后续消息通过 websocket 双向通信。
// @Tags 聊天管理
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} res.Response "请求参数错误"
// @Router /api/chat_groups_records [get]
func (ChatApi) ChatGroupView(c *gin.Context) {
	var upGrader = websocket.Upgrader{
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
			return isLocalOrigin(originURL.Hostname()) && isLocalOrigin(strings.Split(r.Host, ":")[0])
		},
	}
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	addr := conn.RemoteAddr().String()
	nickName := randomname.GenerateName()
	nickNameFirst := string([]rune(nickName)[0])
	avatar := fmt.Sprintf("uploads/chat_avatar/%s.png", nickNameFirst)

	chatUser := ChatUser{
		Conn:     conn,
		NickName: nickName,
		Avatar:   avatar,
	}
	setChatUser(addr, chatUser)
	defer func() {
		deleteChatUser(addr)
		_ = conn.Close()
	}()
	logrus.Infof("%s %s 连接成功", addr, chatUser.NickName)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			deleteChatUser(addr)
			SendGroupMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				MsgType:     OutRoomMsg,
				Content:     fmt.Sprintf("%s 离开聊天室", chatUser.NickName),
				Date:        time.Now(),
				OnlineCount: onlineCount(),
			})
			return
		}

		var request GroupRequest
		if err = json.Unmarshal(p, &request); err != nil {
			SendMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				MsgType:     SystemMsg,
				Content:     "参数绑定失败",
				OnlineCount: onlineCount(),
			})
			continue
		}

		switch request.MsgType {
		case TextMsg:
			if strings.TrimSpace(request.Content) == "" {
				SendMsg(addr, GroupResponse{
					NickName:    chatUser.NickName,
					Avatar:      chatUser.Avatar,
					MsgType:     SystemMsg,
					Content:     "消息不能为空",
					OnlineCount: onlineCount(),
				})
				continue
			}
			SendGroupMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				Content:     request.Content,
				MsgType:     TextMsg,
				Date:        time.Now(),
				OnlineCount: onlineCount(),
			})
		case InRoomMsg:
			SendGroupMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				Content:     fmt.Sprintf("%s 进入聊天室", chatUser.NickName),
				Date:        time.Now(),
				OnlineCount: onlineCount(),
			})
		default:
			SendMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				MsgType:     SystemMsg,
				Content:     "消息类型错误",
				OnlineCount: onlineCount(),
			})
		}
	}
}

// SendGroupMsg 广播消息给聊天室里的全部在线用户，并把消息写入聊天记录表。
func SendGroupMsg(senderAddr string, response GroupResponse) {
	byteData, _ := json.Marshal(response)
	ip, addr := getIPAndAddr(senderAddr)

	global.DB.Create(&models.ChatModel{
		NickName: response.NickName,
		Avatar:   response.Avatar,
		Content:  response.Content,
		IP:       ip,
		Addr:     addr,
		IsGroup:  true,
		MsgType:  response.MsgType,
	})
	for _, chatUser := range snapshotChatUsers() {
		chatUser.Conn.WriteMessage(websocket.TextMessage, byteData)
	}
}

// SendMsg 向单个在线用户推送系统提示类消息。
func SendMsg(_addr string, response GroupResponse) {
	byteData, _ := json.Marshal(response)
	chatUser, ok := getChatUser(_addr)
	if !ok {
		return
	}
	ip, addr := getIPAndAddr(_addr)
	global.DB.Create(&models.ChatModel{
		NickName: response.NickName,
		Avatar:   response.Avatar,
		Content:  response.Content,
		IP:       ip,
		Addr:     addr,
		IsGroup:  false,
		MsgType:  response.MsgType,
	})
	chatUser.Conn.WriteMessage(websocket.TextMessage, byteData)
}

func setChatUser(addr string, user ChatUser) {
	connGroupMu.Lock()
	defer connGroupMu.Unlock()
	connGroupMap[addr] = user
}

func deleteChatUser(addr string) {
	connGroupMu.Lock()
	defer connGroupMu.Unlock()
	delete(connGroupMap, addr)
}

func getChatUser(addr string) (ChatUser, bool) {
	connGroupMu.RLock()
	defer connGroupMu.RUnlock()
	user, ok := connGroupMap[addr]
	return user, ok
}

func snapshotChatUsers() []ChatUser {
	connGroupMu.RLock()
	defer connGroupMu.RUnlock()
	result := make([]ChatUser, 0, len(connGroupMap))
	for _, user := range connGroupMap {
		result = append(result, user)
	}
	return result
}

func onlineCount() int {
	connGroupMu.RLock()
	defer connGroupMu.RUnlock()
	return len(connGroupMap)
}

func isLocalOrigin(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// getIPAndAddr 从 remote address 里拆出 IP，并解析归属地。
func getIPAndAddr(_addr string) (ip string, addr string) {
	addrList := strings.Split(_addr, ":")
	addr = utils.GetAddr(addrList[0])
	return addrList[0], addr
}
