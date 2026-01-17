package routers

import "gvb-server/api"

func (router RouterGroup) MenuRouter() {
	menuApi := api.ApiGroupApp.MenuApi
	//添加菜单
	router.POST("/menus", menuApi.MenuCreateView)
	//查询全部菜单
	router.GET("/menus", menuApi.MenuListView)
	//菜单名称列表
	router.GET("/menu_names", menuApi.MenuNameList)
	//菜单详情
	router.GET("/menus/:id", menuApi.MenuDetailView)
	//删除菜单
	router.DELETE("/menus", menuApi.MenuRemoveView)
	//更新菜单
	router.PUT("/menus", menuApi.MenuUpdateView)
}
