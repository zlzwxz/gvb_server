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
// @Description 获取所有菜单信息及关联的轮播图，按排序降序排列
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=[]MenuResponse} "返回菜单列表"
// @Router /api/menus [get]
func (MenuApi) MenuListView(c *gin.Context) {
	// 先查菜单
	var menuList []models.MenuModel
	var menuIDList []uint
	global.DB.Order("sort desc").Find(&menuList).Select("id").Scan(&menuIDList)
	// 查连接表
	var menuBanners []models.MenuBannerModel
	global.DB.Preload("BannerModel").Order("sort desc").Find(&menuBanners, "menu_id in ?", menuIDList)
	var menus []MenuResponse
	for _, model := range menuList {
		// model就是一个菜单
		var banners []Banner
		for _, banner := range menuBanners {
			if model.ID != banner.MenuID {
				continue
			}
			banners = append(banners, Banner{
				ID:   banner.BannerID,
				Path: banner.BannerModel.Path,
			})
		}
		menus = append(menus, MenuResponse{
			MenuModel: model,
			Banners:   banners,
		})
	}
	res.OkWithData(menus, c)
	return
}
