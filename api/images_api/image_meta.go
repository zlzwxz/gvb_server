package images_api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"gorm.io/gorm"
)

type imageMetaCategoryItem struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type imageMetaResponse struct {
	Categories       []imageMetaCategoryItem `json:"categories"`
	PublicCount      int64                   `json:"public_count"`
	PrivateCount     int64                   `json:"private_count"`
	MineCount        int64                   `json:"mine_count"`
	AccessibleCount  int64                   `json:"accessible_count"`
}

func (ImagesApi) ImageMetaView(c *gin.Context) {
	claimsAny, ok := c.Get("claims")
	if !ok {
		res.FailWithMessage("未登录", c)
		return
	}
	claims := claimsAny.(*jwts.CustomClaims)

	query := applyImageAccessQuery(global.DB.Model(&models.BannerModel{}), claims, "", false)

	var accessibleCount int64
	_ = query.Count(&accessibleCount).Error

	var publicCount int64
	_ = query.Session(&gorm.Session{}).Where(publicImageWhere()).Count(&publicCount).Error

	var privateCount int64
	_ = query.Session(&gorm.Session{}).Where("visibility = ?", models.ImageVisibilityPrivate).Count(&privateCount).Error

	mineQuery := applyImageAccessQuery(global.DB.Model(&models.BannerModel{}), claims, "", true)
	var mineCount int64
	_ = mineQuery.Count(&mineCount).Error

	var rawItems []imageMetaCategoryItem
	_ = query.Select("COALESCE(NULLIF(image_category, ''), '未分类') AS category, COUNT(1) AS count").
		Group("COALESCE(NULLIF(image_category, ''), '未分类')").
		Order("count desc, category asc").
		Scan(&rawItems).Error

	items := make([]imageMetaCategoryItem, 0, len(rawItems))
	for _, item := range rawItems {
		items = append(items, imageMetaCategoryItem{
			Category: normalizeImageCategory(strings.TrimSpace(item.Category)),
			Count:    item.Count,
		})
	}

	res.OkWithData(imageMetaResponse{
		Categories:      items,
		PublicCount:     publicCount,
		PrivateCount:    privateCount,
		MineCount:       mineCount,
		AccessibleCount: accessibleCount,
	}, c)
}
