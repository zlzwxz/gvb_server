package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// DataRouter 注册统计看板相关路由。
func (router RouterGroup) DataRouter() {
	dataApp := api.ApiGroupApp.DataApi
	router.GET("data_login", dataApp.SevenLoginView)
	router.GET("data_sum", dataApp.DataSumView)
	router.GET("data_sum/admin", middleware.JwtAdmin(), dataApp.AdminDataSumView)
}
