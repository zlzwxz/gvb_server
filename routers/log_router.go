package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

func (router RouterGroup) LogRouter() {
	logApp := api.ApiGroupApp.LogApi
	// 日志列表
	router.GET("logs", middleware.JwtAdmin(), logApp.LogListView)
	// 删除日志
	router.DELETE("logs", middleware.JwtAdmin(), logApp.LogRemoveListView)
}
