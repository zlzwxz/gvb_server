package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// ImageRouter 注册图片管理路由。
func (router RouterGroup) ImageRouter() {
	imagesApi := api.ApiGroupApp.ImagesApi
	router.POST("images", middleware.JwtAuth(), imagesApi.ImageUploadView)
	router.GET("images", middleware.JwtAuth(), imagesApi.ImageListView)
	router.DELETE("images", middleware.JwtAuth(), imagesApi.ImageRemoveView)
	router.PUT("images", middleware.JwtAuth(), imagesApi.ImageUpdateView)
}
