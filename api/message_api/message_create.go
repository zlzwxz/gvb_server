package message_api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
)

// MessageRequest 消息创建请求参数
type MessageRequest struct {
	RevUserID uint   `json:"rev_user_id" binding:"required" swag:"description:接收人ID"`
	Content   string `json:"content" binding:"required" swag:"description:消息内容"`
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
	var cr MessageRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)
	content := strings.TrimSpace(cr.Content)
	if content == "" {
		res.FailWithMessage("消息内容不能为空", c)
		return
	}
	if claims.UserID == cr.RevUserID {
		res.FailWithMessage("不能给自己发送私信", c)
		return
	}

	var sendUser models.UserModel
	if err := global.DB.Take(&sendUser, claims.UserID).Error; err != nil {
		res.FailWithMessage("发送人不存在", c)
		return
	}

	var recvUser models.UserModel
	if err := global.DB.Take(&recvUser, cr.RevUserID).Error; err != nil {
		res.FailWithMessage("接收人不存在", c)
		return
	}

	err := global.DB.Create(&models.MessageModel{
		SendUserID:       claims.UserID,
		SendUserNickName: sendUser.NickName,
		SendUserAvatar:   sendUser.Avatar,
		RevUserID:        cr.RevUserID,
		RevUserNickName:  recvUser.NickName,
		RevUserAvatar:    recvUser.Avatar,
		IsRead:           false,
		Content:          content,
	}).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("消息发送失败", c)
		return
	}

	res.OkWithMessage("消息发送成功", c)
}
