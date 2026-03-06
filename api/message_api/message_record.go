package message_api

import (
	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
)

// MessageRecordRequest 消息记录请求参数
type MessageRecordRequest struct {
	UserID uint `form:"user_id" json:"user_id" binding:"required" msg:"请输入查询的用户id" swag:"description:用户ID"`
}

// MessageRecordView 获取用户消息记录
// @Summary 获取用户消息记录
// @Description 获取当前用户与指定用户的消息记录
// @Tags 消息管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param user_id query int true "用户ID"
// @Success 200 {object} res.Response{data=[]models.MessageModel} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages/record [get]
func (MessageApi) MessageRecordView(c *gin.Context) {
	var cr MessageRecordRequest
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	var messageList []models.MessageModel
	err := global.DB.Order("created_at asc").
		Find(&messageList, "(send_user_id = ? and rev_user_id = ?) or (send_user_id = ? and rev_user_id = ?)", claims.UserID, cr.UserID, cr.UserID, claims.UserID).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("获取消息记录失败", c)
		return
	}

	if err = global.DB.Model(&models.MessageModel{}).
		Where("send_user_id = ? and rev_user_id = ? and is_read = ?", cr.UserID, claims.UserID, false).
		Update("is_read", true).Error; err != nil {
		global.Log.Error(err)
	}

	res.OkWithData(messageList, c)
}
