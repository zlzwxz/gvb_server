package message_api

import (
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
)

// Message 用于前端会话列表展示。
// 它不是数据库表，而是把多条 MessageModel 聚合之后的“会话摘要”。
type Message struct {
	UserID             uint      `json:"user_id" swag:"description:对话用户ID"`
	NickName           string    `json:"nick_name" swag:"description:对话用户昵称"`
	Avatar             string    `json:"avatar" swag:"description:对话用户头像"`
	SendUserID         uint      `json:"send_user_id" swag:"description:发送人ID"`
	SendUserNickName   string    `json:"send_user_nick_name" swag:"description:发送人昵称"`
	SendUserAvatar     string    `json:"send_user_avatar" swag:"description:发送人头像"`
	RevUserID          uint      `json:"rev_user_id" swag:"description:接收人ID"`
	RevUserNickName    string    `json:"rev_user_nick_name" swag:"description:接收人昵称"`
	RevUserAvatar      string    `json:"rev_user_avatar" swag:"description:接收人头像"`
	Content            string    `json:"content" swag:"description:最新消息内容"`
	CreatedAt          time.Time `json:"created_at" swag:"description:最新消息时间"`
	MessageCount       int       `json:"message_count" swag:"description:消息条数"`
	UnreadCount        int       `json:"unread_count" swag:"description:未读消息数"`
	LatestSenderUserID uint      `json:"latest_sender_user_id" swag:"description:最后一条消息发送人ID"`
}

// conversationKey 把两个人的用户 ID 归一化成同一个会话键。
// 例如 1 和 5，无论消息方向是 1->5 还是 5->1，最终都会落成 1:5。
func conversationKey(a, b uint) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("%d:%d", a, b)
}

// MessageListView 获取消息列表
// @Summary 获取消息列表
// @Description 获取当前用户的消息列表，按对话维度分组展示最近一条消息、总消息数和未读数。
// @Tags 消息管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=[]Message} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages [get]
func (MessageApi) MessageListView(c *gin.Context) {
	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	var messageList []models.MessageModel
	if err := global.DB.Order("created_at desc").
		Find(&messageList, "send_user_id = ? or rev_user_id = ?", claims.UserID, claims.UserID).Error; err != nil {
		global.Log.Error(err)
		res.FailWithMessage("获取消息列表失败", c)
		return
	}

	messageGroup := map[string]*Message{}
	for _, model := range messageList {
		key := conversationKey(model.SendUserID, model.RevUserID)
		current, exists := messageGroup[key]
		if !exists {
			item := Message{
				SendUserID:         model.SendUserID,
				SendUserNickName:   model.SendUserNickName,
				SendUserAvatar:     model.SendUserAvatar,
				RevUserID:          model.RevUserID,
				RevUserNickName:    model.RevUserNickName,
				RevUserAvatar:      model.RevUserAvatar,
				Content:            model.Content,
				CreatedAt:          model.CreatedAt,
				MessageCount:       0,
				UnreadCount:        0,
				LatestSenderUserID: model.SendUserID,
			}
			if model.SendUserID == claims.UserID {
				item.UserID = model.RevUserID
				item.NickName = model.RevUserNickName
				item.Avatar = model.RevUserAvatar
			} else {
				item.UserID = model.SendUserID
				item.NickName = model.SendUserNickName
				item.Avatar = model.SendUserAvatar
			}
			current = &item
			messageGroup[key] = current
		}
		current.MessageCount++
		if model.RevUserID == claims.UserID && !model.IsRead {
			current.UnreadCount++
		}
	}

	messages := make([]Message, 0, len(messageGroup))
	for _, item := range messageGroup {
		messages = append(messages, *item)
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})

	res.OkWithData(messages, c)
}
