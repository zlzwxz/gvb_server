package images_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"
)

// ImageListView 图片列表
func (ImagesApi) ImageListView(c *gin.Context) {
	var cr models.PageInfo
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	fmt.Println(cr)
	list, count, err := common.ComList(models.BannerModel{}, common.Option{
		PageInfo: cr,
		Debug:    false,
	})

	res.OkWithList(list, count, c)

	return

}
