package routers

import "gvb-server/api"

func (router RouterGroup) AdvertRouter() {
	advertApp := api.ApiGroupApp.AdvertApi
	router.POST("adverts", advertApp.AdvertCreateView)
	router.GET("adverts", advertApp.AdvertListView)
	router.PUT("adverts/:id", advertApp.AdvertUpdateView)
	router.DELETE("adverts", advertApp.AdvertRemoveView)
}
