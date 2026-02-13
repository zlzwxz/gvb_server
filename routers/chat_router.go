package routers

import (
	"gvb-server/api"
)

func (router RouterGroup) ChatRouter() {
	chatApp := api.ApiGroupApp.ChatApi
	router.POST("chat", chatApp.ChatListView)
	router.POST("chat/group", chatApp.ChatGroupView)
}
