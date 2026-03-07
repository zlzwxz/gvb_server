package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// MenuRouter 注册导航菜单相关路由。
func (router RouterGroup) MenuRouter() {
	menuApi := api.ApiGroupApp.MenuApi
	router.POST("/menus", middleware.JwtAdmin(), menuApi.MenuCreateView)
	router.GET("/menus", menuApi.MenuListView)
	router.GET("/menu_names", menuApi.MenuNameList)
	router.GET("/menus/:id", menuApi.MenuDetailView)
	router.DELETE("/menus", middleware.JwtAdmin(), menuApi.MenuRemoveView)
	router.PUT("/menus", middleware.JwtAdmin(), menuApi.MenuUpdateView)
}
