package chat_api

import (
	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"
)

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
