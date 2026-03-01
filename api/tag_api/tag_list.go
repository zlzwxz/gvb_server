package tag_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"

	"github.com/gin-gonic/gin"
)

func (TagApi) TagListView(c *gin.Context) {
	var cr models.PageInfo
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	list, count, _ := common.ComList(models.TagModel{}, common.Option{
		PageInfo: cr,
	})
	res.OkWithList(list, count, c)
}
