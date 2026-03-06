package routers

import "gvb-server/api"

// MenuRouter 注册导航菜单相关路由。
func (router RouterGroup) MenuRouter() {
	menuApi := api.ApiGroupApp.MenuApi
	router.POST("/menus", menuApi.MenuCreateView)
	router.GET("/menus", menuApi.MenuListView)
	router.GET("/menu_names", menuApi.MenuNameList)
	router.GET("/menus/:id", menuApi.MenuDetailView)
	router.DELETE("/menus", menuApi.MenuRemoveView)
	router.PUT("/menus", menuApi.MenuUpdateView)
}
