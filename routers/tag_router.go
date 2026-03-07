package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// TagRouter 注册标签相关路由。
func (router RouterGroup) TagRouter() {
	tagApp := api.ApiGroupApp.TagApi
	router.POST("tags", middleware.JwtAdmin(), tagApp.TagCreateView)
	router.GET("tags", tagApp.TagListView)
	router.PUT("tags/:id", middleware.JwtAdmin(), tagApp.TagUpdateView)
	router.DELETE("tags", middleware.JwtAdmin(), tagApp.TagRemoveView)
	router.GET("tags/names", tagApp.TagNameListView)
}
