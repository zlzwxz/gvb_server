package message_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"time"
)

// Message 消息响应结构
type Message struct {
	SendUserID       uint      `json:"send_user_id" swag:"description:发送人ID"` // 发送人id
	SendUserNickName string    `json:"send_user_nick_name" swag:"description:发送人昵称"`
	SendUserAvatar   string    `json:"send_user_avatar" swag:"description:发送人头像"`
	RevUserID        uint      `json:"rev_user_id" swag:"description:接收人ID"` // 接收人id
	RevUserNickName  string    `json:"rev_user_nick_name" swag:"description:接收人昵称"`
	RevUserAvatar    string    `json:"rev_user_avatar" swag:"description:接收人头像"`
	Content          string    `json:"content" swag:"description:消息内容"`       // 消息内容
	CreatedAt        time.Time `json:"created_at" swag:"description:最新消息时间"`    // 最新的消息时间
	MessageCount     int       `json:"message_count" swag:"description:消息条数"` // 消息条数
}

type MessageGroup map[uint]*Message

// MessageListView 获取消息列表
// @Summary 获取消息列表
// @Description 获取当前用户的消息列表，按消息组分组显示
// @Tags 消息管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=[]Message} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages [get]
func (MessageApi) MessageListView(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	var messageGroup = MessageGroup{}
	var messageList []models.MessageModel
	var messages []Message

	global.DB.Order("created_at asc").
		Find(&messageList, "send_user_id = ? or rev_user_id = ?", claims.UserID, claims.UserID)
	for _, model := range messageList {
		// 判断是一个组的条件
		// send_user_id 和 rev_user_id 其中一个
		// 1 2  2 1
		// 1 3  3 1 是一组
		message := Message{
			SendUserID:       model.SendUserID,
			SendUserNickName: model.SendUserNickName,
			SendUserAvatar:   model.SendUserAvatar,
			RevUserID:        model.RevUserID,
			RevUserNickName:  model.RevUserNickName,
			RevUserAvatar:    model.RevUserAvatar,
			Content:          model.Content,
			CreatedAt:        model.CreatedAt,
			MessageCount:     1,
		}
		idNum := model.SendUserID + model.RevUserID
		val, ok := messageGroup[idNum]
		if !ok {
			// 不存在
			messageGroup[idNum] = &message
			continue
		}
		message.MessageCount = val.MessageCount + 1
		messageGroup[idNum] = &message
	}
	for _, message := range messageGroup {
		messages = append(messages, *message)
	}

	res.OkWithData(messages, c)
	return
}
