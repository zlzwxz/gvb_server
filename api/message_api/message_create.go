package message_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// MessageRequest 消息创建请求参数
type MessageRequest struct {
	SendUserID uint   `json:"send_user_id" binding:"required" swag:"description:发送人ID"` // 发送人id
	RevUserID  uint   `json:"rev_user_id" binding:"required" swag:"description:接收人ID"`  // 接收人id
	Content    string `json:"content" binding:"required" swag:"description:消息内容"`       // 消息内容
}

// MessageCreateView 发布消息
// @Summary 发布消息
// @Description 发送消息给指定用户
// @Tags 消息管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body MessageRequest true "消息信息"
// @Success 200 {object} res.Response{msg=string} "发送成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "发送人或接收人不存在"
// @Router /api/messages [post]
func (MessageApi) MessageCreateView(c *gin.Context) {
	// 当前用户发布消息
	// SendUserID 就是当前登录人的id
	var cr MessageRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	var senUser, recvUser models.UserModel

	err = global.DB.Take(&senUser, cr.SendUserID).Error
	if err != nil {
		res.FailWithMessage("发送人不存在", c)
		return
	}
	err = global.DB.Take(&recvUser, cr.RevUserID).Error
	if err != nil {
		res.FailWithMessage("接收人不存在", c)
		return
	}

	err = global.DB.Create(&models.MessageModel{
		SendUserID:       cr.SendUserID,
		SendUserNickName: senUser.NickName,
		SendUserAvatar:   senUser.Avatar,
		RevUserID:        cr.RevUserID,
		RevUserNickName:  recvUser.NickName,
		RevUserAvatar:    recvUser.Avatar,
		IsRead:           false,
		Content:          cr.Content,
	}).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("消息发送失败", c)
		return
	}
	res.OkWithMessage("消息发送成功", c)
	return
}
