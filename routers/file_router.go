package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

func (router RouterGroup) FileRouter() {
	fileApi := api.ApiGroupApp.FileApi
	router.POST("files", middleware.JwtAuth(), fileApi.FileUploadView)
	router.GET("files/:id/download", middleware.JwtAuth(), fileApi.FileDownloadView)
}
