package images_api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
)

type imageListQuery struct {
	models.PageInfo
	Visibility string `form:"visibility"`
	Category   string `form:"category"`
	Mine       bool   `form:"mine"`
}

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
	var cr imageListQuery
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

	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 12
	}
	if cr.Limit > 100 {
		cr.Limit = 100
	}

	query := global.DB.Model(&models.BannerModel{})
	query = applyImageAccessQuery(query, claims, cr.Visibility, cr.Mine)

	key := strings.TrimSpace(cr.Key)
	if key != "" {
		like := "%" + key + "%"
		query = query.Where("name LIKE ? OR image_category LIKE ? OR path LIKE ?", like, like, like)
	}

	category := normalizeImageCategory(cr.Category)
	if strings.TrimSpace(cr.Category) != "" {
		query = query.Where("image_category = ? OR (image_category = '' AND ? = '未分类')", category, category)
	}

	var count int64
	if err = query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取图片数量失败", c)
		return
	}

	sortValue := normalizeImageSort(cr.Sort)
	var list []models.BannerModel
	if err = query.Order(sortValue).Limit(cr.Limit).Offset((cr.Page - 1) * cr.Limit).Find(&list).Error; err != nil {
		res.FailWithMessage("获取图片列表失败", c)
		return
	}

	res.OkWithList(normalizeImageRows(list), count, c)

}
