package menu_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// Banner 定义轮播图id
// @Description 轮播图路径信息
type Banner struct {
	ID   uint   `json:"id"`
	Path string `json:"path"`
}

// MenuResponse 定义菜单响应结构
// @Description 菜单响应数据
type MenuResponse struct {
	models.MenuModel
	Banners []Banner `json:"banners"`
}

// MenuListView 获取菜单列表
// @Summary 获取菜单列表
// @Description 获取所有菜单信息及关联的轮播图，按排序升序排列，便于前端直接渲染
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=[]MenuResponse} "返回菜单列表"
// @Router /api/menus [get]
func (MenuApi) MenuListView(c *gin.Context) {
	var menuList []models.MenuModel
	if err := global.DB.Order("sort asc").Find(&menuList).Error; err != nil {
		global.Log.Error(err)
		res.FailWithMessage("获取菜单列表失败", c)
		return
	}
	if len(menuList) == 0 {
		res.OkWithData([]MenuResponse{}, c)
		return
	}

	menuIDList := make([]uint, 0, len(menuList))
	for _, menu := range menuList {
		menuIDList = append(menuIDList, menu.ID)
	}

	var menuBanners []models.MenuBannerModel
	if err := global.DB.Preload("BannerModel").
		Order("menu_id asc").
		Order("sort asc").
		Find(&menuBanners, "menu_id in ?", menuIDList).Error; err != nil {
		global.Log.Error(err)
		res.FailWithMessage("获取菜单轮播图失败", c)
		return
	}

	bannerMap := make(map[uint][]Banner, len(menuList))
	for _, banner := range menuBanners {
		bannerMap[banner.MenuID] = append(bannerMap[banner.MenuID], Banner{
			ID:   banner.BannerID,
			Path: banner.BannerModel.Path,
		})
	}

	menus := make([]MenuResponse, 0, len(menuList))
	for _, model := range menuList {
		menus = append(menus, MenuResponse{
			MenuModel: model,
			Banners:   bannerMap[model.ID],
		})
	}
	res.OkWithData(menus, c)
}
