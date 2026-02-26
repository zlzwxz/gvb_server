package routers

import (
	"gvb-server/api"
)

func (router RouterGroup) ChatRouter() {
	chatApp := api.ApiGroupApp.ChatApi
	//聊天记录
	router.GET("chat_groups", chatApp.ChatListView)
	// 加入聊天室
	router.GET("chat_groups_records", chatApp.ChatGroupView)
}
