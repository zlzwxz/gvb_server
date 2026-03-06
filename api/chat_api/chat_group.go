package chat_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// ConnGroupMap 保存当前在线连接。
// key 使用 remote address，是因为同一个聊天室里每个 websocket 连接天然唯一。
var ConnGroupMap = map[string]ChatUser{}

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
			return true
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
	ConnGroupMap[addr] = chatUser
	logrus.Infof("%s %s 连接成功", addr, chatUser.NickName)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			SendGroupMsg(conn, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				MsgType:     OutRoomMsg,
				Content:     fmt.Sprintf("%s 离开聊天室", chatUser.NickName),
				Date:        time.Now(),
				OnlineCount: len(ConnGroupMap) - 1,
			})
			break
		}

		var request GroupRequest
		if err = json.Unmarshal(p, &request); err != nil {
			SendMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				MsgType:     SystemMsg,
				Content:     "参数绑定失败",
				OnlineCount: len(ConnGroupMap),
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
					OnlineCount: len(ConnGroupMap),
				})
				continue
			}
			SendGroupMsg(conn, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				Content:     request.Content,
				MsgType:     TextMsg,
				Date:        time.Now(),
				OnlineCount: len(ConnGroupMap),
			})
		case InRoomMsg:
			SendGroupMsg(conn, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				Content:     fmt.Sprintf("%s 进入聊天室", chatUser.NickName),
				Date:        time.Now(),
				OnlineCount: len(ConnGroupMap),
			})
		default:
			SendMsg(addr, GroupResponse{
				NickName:    chatUser.NickName,
				Avatar:      chatUser.Avatar,
				MsgType:     SystemMsg,
				Content:     "消息类型错误",
				OnlineCount: len(ConnGroupMap),
			})
		}
	}
	defer conn.Close()
	delete(ConnGroupMap, addr)
}

// SendGroupMsg 广播消息给聊天室里的全部在线用户，并把消息写入聊天记录表。
func SendGroupMsg(conn *websocket.Conn, response GroupResponse) {
	byteData, _ := json.Marshal(response)
	_addr := conn.RemoteAddr().String()
	ip, addr := getIPAndAddr(_addr)

	global.DB.Create(&models.ChatModel{
		NickName: response.NickName,
		Avatar:   response.Avatar,
		Content:  response.Content,
		IP:       ip,
		Addr:     addr,
		IsGroup:  true,
		MsgType:  response.MsgType,
	})
	for _, chatUser := range ConnGroupMap {
		chatUser.Conn.WriteMessage(websocket.TextMessage, byteData)
	}
}

// SendMsg 向单个在线用户推送系统提示类消息。
func SendMsg(_addr string, response GroupResponse) {
	byteData, _ := json.Marshal(response)
	chatUser := ConnGroupMap[_addr]
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

// getIPAndAddr 从 remote address 里拆出 IP，并解析归属地。
func getIPAndAddr(_addr string) (ip string, addr string) {
	addrList := strings.Split(_addr, ":")
	addr = utils.GetAddr(addrList[0])
	return addrList[0], addr
}
