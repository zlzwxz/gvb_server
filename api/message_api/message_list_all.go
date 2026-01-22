package message_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"
)

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
