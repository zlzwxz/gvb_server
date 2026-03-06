package routers

import "gvb-server/api"

// ChatRouter 注册聊天室相关路由。
func (router RouterGroup) ChatRouter() {
	chatApp := api.ApiGroupApp.ChatApi
	router.GET("chat_groups", chatApp.ChatListView)
	router.GET("chat_groups_records", chatApp.ChatGroupView)
}
