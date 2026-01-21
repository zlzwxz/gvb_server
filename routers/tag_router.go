package routers

import "gvb-server/api"

func (router RouterGroup) TagRouter() {
	tagApp := api.ApiGroupApp.TagApi
	router.POST("tags", tagApp.TagCreateView)
	router.GET("tags", tagApp.TagListView)
	router.PUT("tags/:id", tagApp.TagUpdateView)
	router.DELETE("tags", tagApp.TagRemoveView)
}
