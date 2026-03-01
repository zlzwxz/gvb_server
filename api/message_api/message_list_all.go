package message_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"

	"github.com/gin-gonic/gin"
)

// MessageListAllView 获取所有消息列表
// @Summary 获取所有消息列表
// @Description 获取所有消息的列表数据（管理员权限）
// @Tags 消息管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.MessageModel}} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages/all [get]
func (MessageApi) MessageListAllView(c *gin.Context) {
	var cr models.PageInfo
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
	}
	list, counnt, _ := common.ComList(models.MessageModel{}, common.Option{
		PageInfo: cr,
	})
	res.OkWithList(list, counnt, c)
	return
}
