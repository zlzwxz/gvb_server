package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// AdvertRouter 注册广告模块路由。
func (router RouterGroup) AdvertRouter() {
	advertApp := api.ApiGroupApp.AdvertApi
	router.POST("adverts", middleware.JwtAdmin(), advertApp.AdvertCreateView)
	router.GET("adverts", advertApp.AdvertListView)
	router.PUT("adverts/:id", middleware.JwtAdmin(), advertApp.AdvertUpdateView)
	router.DELETE("adverts", middleware.JwtAdmin(), advertApp.AdvertRemoveView)
}
