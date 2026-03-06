package images_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"
	"gvb-server/utils/jwts"
)

// ImageListView 图片列表
// @Tags 图片管理
// @Summary 获取图片列表
// @Description 获取分页的图片列表数据
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param sort query string false "排序方式"
// @Success 200 {object} res.Response{data=res.ListResponse[models.BannerModel]}
// @Router /api/images [get]
func (ImagesApi) ImageListView(c *gin.Context) {
	var cr models.PageInfo
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	_claims, ok := c.Get("claims")
	if !ok {
		res.FailWithMessage("未登录", c)
		return
	}
	claims := _claims.(*jwts.CustomClaims)

	option := common.Option{
		PageInfo: cr,
		Debug:    false,
	}
	if !isImageAdmin(claims) {
		like := imageOwnerPathLike(claims.NickName)
		if like == "" {
			res.OkWithList([]models.BannerModel{}, 0, c)
			return
		}
		option.Where = "path LIKE ?"
		option.WhereArgs = []interface{}{like}
	}

	list, count, err := common.ComList(models.BannerModel{}, option)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	res.OkWithList(list, count, c)

	return

}
