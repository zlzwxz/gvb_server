package chat_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"

	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
)

// ChatListView 获取聊天列表
// @Summary 获取聊天列表
// @Description 获取聊天消息列表
// @Tags 聊天管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.ChatModel}} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/chats [get]
func (ChatApi) ChatListView(c *gin.Context) {
	var cr models.PageInfo
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
	}
	cr.Sort = "created_at desc"
	list, count, _ := common.ComList(models.ChatModel{}, common.Option{
		PageInfo: cr,
	})
	data := filter.Omit("list", list)
	_list, _ := data.(filter.Filter)
	if string(_list.MustMarshalJSON()) == "{}" {
		list = make([]models.ChatModel, 0)
		res.OkWithList(list, count, c)
		return
	}
	res.OkWithList(data, count, c)
}
