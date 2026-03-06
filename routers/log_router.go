package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// LogRouter 注册日志管理路由。
// 日志属于后台能力，所以统一要求管理员权限。
func (router RouterGroup) LogRouter() {
	logApp := api.ApiGroupApp.LogApi
	router.GET("logs", middleware.JwtAdmin(), logApp.LogListView)
	router.DELETE("logs", middleware.JwtAdmin(), logApp.LogRemoveListView)
}
