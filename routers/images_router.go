package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

func (router RouterGroup) ImageRouter() {
	imagesApi := api.ApiGroupApp.ImagesApi
	router.POST("images", middleware.JwtAuth(), imagesApi.ImageUploadView)
	router.GET("images", imagesApi.ImageListView)
	router.DELETE("images", imagesApi.ImageRemoveView)
	router.PUT("images", imagesApi.ImageUpdateView)
}
