package message_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
)

// MessageRecordRequest 消息记录请求参数
type MessageRecordRequest struct {
	UserID uint `json:"user_id" binding:"required" msg:"请输入查询的用户id" swag:"description:用户ID"`
}

// MessageRecordView 获取用户消息记录
// @Summary 获取用户消息记录
// @Description 获取当前用户与指定用户的消息记录
// @Tags 消息管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body MessageRecordRequest true "用户ID"
// @Success 200 {object} res.Response{data=[]models.MessageModel} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages/record [post]
func (MessageApi) MessageRecordView(c *gin.Context) {
	var cr MessageRecordRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	var _messageList []models.MessageModel
	var messageList = make([]models.MessageModel, 0)
	global.DB.Order("created_at asc").
		Find(&_messageList, "send_user_id = ? or rev_user_id = ?", claims.UserID, claims.UserID)
	for _, model := range _messageList {
		// 判断是一个组的条件
		// send_user_id 和 rev_user_id 其中一个
		// 1 2  2 1
		// 1 3  3 1 是一组
		if model.RevUserID == cr.UserID || model.SendUserID == cr.UserID {
			messageList = append(messageList, model)
		}
	}

	// 点开消息，里面的每一条消息，都从未读变成已读
	fmt.Println(messageList)
	res.OkWithData(messageList, c)
}
