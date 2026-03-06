package menu_api

import "gvb-server/models"

// normalizeMenuRequest 统一给菜单请求补默认值。
// 这样前端即便没有传切换秒数，也不会把轮播和简介切成 0 秒。
func normalizeMenuRequest(cr *MenuRequest) {
	if cr.AbstractTime <= 0 {
		cr.AbstractTime = 7
	}
	if cr.BannerTime <= 0 {
		cr.BannerTime = 7
	}
}

// buildMenuBannerModels 根据前端的图片顺序，生成连接表记录。
// 如果前端没有显式传 sort，就回退到当前循环顺序，保证轮播顺序稳定。
func buildMenuBannerModels(menuID uint, imageSortList []ImageSort) []models.MenuBannerModel {
	menuBannerList := make([]models.MenuBannerModel, 0, len(imageSortList))
	for index, sortItem := range imageSortList {
		sortValue := sortItem.Sort
		if sortValue <= 0 {
			sortValue = index + 1
		}
		menuBannerList = append(menuBannerList, models.MenuBannerModel{
			MenuID:   menuID,
			BannerID: sortItem.ImageID,
			Sort:     sortValue,
		})
	}
	return menuBannerList
}

// buildMenuUpdateMap 把允许更新的字段收敛成一个 map。
// 这样可以显式控制哪些字段会被写回数据库，避免把 image_sort_list 之类的临时字段误更新进去。
func buildMenuUpdateMap(cr MenuRequest) map[string]any {
	return map[string]any{
		"title":         cr.Title,
		"path":          cr.Path,
		"slogan":        cr.Slogan,
		"abstract":      cr.Abstract,
		"abstract_time": cr.AbstractTime,
		"banner_time":   cr.BannerTime,
		"sort":          cr.Sort,
	}
}
