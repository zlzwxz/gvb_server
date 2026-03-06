package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// MessageRouter 注册私信相关路由。
func (router RouterGroup) MessageRouter() {
	app := api.ApiGroupApp.MessageApi
	router.POST("messages", middleware.JwtAuth(), app.MessageCreateView)
	router.GET("messages", middleware.JwtAuth(), app.MessageListView)
	router.GET("messages/record", middleware.JwtAuth(), app.MessageRecordView)
	router.GET("messages/all", middleware.JwtAdmin(), app.MessageListAllView)

	// 保留旧地址，避免前端或第三方调用方还没迁移时直接失效。
	router.GET("messages_record", middleware.JwtAuth(), app.MessageRecordView)
	router.GET("messages_all", middleware.JwtAdmin(), app.MessageListAllView)
}
