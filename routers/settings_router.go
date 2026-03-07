package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// SettinsRouter 注册系统配置相关路由。
// 方法名沿用旧拼写，避免影响现有调用代码。
func (router RouterGroup) SettinsRouter() {
	srttinsApi := api.ApiGroupApp.SettingsApi
	router.GET("/settings/public/site_info", srttinsApi.PublicSiteInfoView)
	router.GET("/settings/site_info/sync_fengfeng_preview", middleware.JwtAdmin(), srttinsApi.PreviewFengfengArticlesView)
	router.POST("/settings/site_info/sync_fengfeng", middleware.JwtAdmin(), srttinsApi.SyncFengfengArticlesView)
	router.GET("/settings/site_info/sync_fengfeng_images_preview", middleware.JwtAdmin(), srttinsApi.PreviewFengfengImagesView)
	router.POST("/settings/site_info/sync_fengfeng_images", middleware.JwtAdmin(), srttinsApi.SyncFengfengImagesView)
	router.GET("/settings/:name", middleware.JwtAdmin(), srttinsApi.SettingsInfoView)
	router.PUT("/settings/:name", middleware.JwtAdmin(), srttinsApi.SettingsInfoUpdateView)
}
