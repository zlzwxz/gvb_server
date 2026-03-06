package menu_api

import (
	"errors"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MenuDetailView 获取菜单详情
// @Summary 获取菜单详情
// @Description 根据ID获取菜单详细信息及其关联的轮播图
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path string true "菜单ID"
// @Success 200 {object} res.Response{data=MenuResponse} "返回菜单详情"
// @Router /api/menus/{id} [get]
func (MenuApi) MenuDetailView(c *gin.Context) {
	id := c.Param("id")
	var menuModel models.MenuModel
	if err := global.DB.Take(&menuModel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.FailWithMessage("菜单不存在", c)
			return
		}
		global.Log.Error(err)
		res.FailWithMessage("获取菜单详情失败", c)
		return
	}

	var menuBanners []models.MenuBannerModel
	if err := global.DB.Preload("BannerModel").
		Order("sort asc").
		Find(&menuBanners, "menu_id = ?", id).Error; err != nil {
		global.Log.Error(err)
		res.FailWithMessage("获取菜单详情失败", c)
		return
	}

	banners := make([]Banner, 0, len(menuBanners))
	for _, banner := range menuBanners {
		banners = append(banners, Banner{
			ID:   banner.BannerID,
			Path: banner.BannerModel.Path,
		})
	}

	res.OkWithData(MenuResponse{
		MenuModel: menuModel,
		Banners:   banners,
	}, c)
}
