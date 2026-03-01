package routers

import "gvb-server/api"

func (router RouterGroup) DataRouter() {
	dataApp := api.ApiGroupApp.DataApi
	router.GET("data_login", dataApp.SevenLoginView)
	router.GET("data_sum", dataApp.DataSumView)
}
