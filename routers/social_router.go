package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// SocialRouter 注册好友/群组/在线状态相关路由。
func (router RouterGroup) SocialRouter() {
	app := api.ApiGroupApp.SocialApi
	router.GET("social/manage/summary", middleware.JwtAdmin(), app.AdminSummaryView)
	router.GET("social/manage/follows", middleware.JwtAdmin(), app.AdminFollowListView)
	router.GET("social/manage/blocks", middleware.JwtAdmin(), app.AdminBlockListView)
	router.GET("social/manage/groups", middleware.JwtAdmin(), app.AdminGroupListView)
	router.GET("social/summary", middleware.JwtAuth(), app.SummaryView)
	router.GET("social/discovery", middleware.JwtAuth(), app.DiscoveryView)
	router.GET("social/friends", middleware.JwtAuth(), app.FriendListView)
	router.GET("social/relations/:id", middleware.JwtAuth(), app.RelationView)
	router.POST("social/follows/:id", middleware.JwtAuth(), app.FollowView)
	router.DELETE("social/follows/:id", middleware.JwtAuth(), app.UnfollowView)
	router.GET("social/blocks", middleware.JwtAuth(), app.BlockListView)
	router.POST("social/blocks/:id", middleware.JwtAuth(), app.BlockView)
	router.DELETE("social/blocks/:id", middleware.JwtAuth(), app.UnblockView)
	router.GET("social/conversations", middleware.JwtAuth(), app.ConversationListView)
	router.GET("social/direct/messages", middleware.JwtAuth(), app.DirectMessageListView)
	router.POST("social/direct/messages", middleware.JwtAuth(), app.DirectMessageCreateView)
	router.GET("social/messages/search", middleware.JwtAuth(), app.MessageSearchView)
	router.POST("social/messages/:id/recall", middleware.JwtAuth(), app.MessageRecallView)
	router.GET("social/calls", middleware.JwtAuth(), app.CallLogListView)
	router.GET("social/groups", middleware.JwtAuth(), app.GroupListView)
	router.POST("social/groups", middleware.JwtAuth(), app.GroupCreateView)
	router.POST("social/groups/join", middleware.JwtAuth(), app.GroupJoinView)
	router.GET("social/groups/:id", middleware.JwtAuth(), app.GroupDetailView)
	router.GET("social/groups/:id/messages", middleware.JwtAuth(), app.GroupMessageListView)
	router.POST("social/groups/:id/messages", middleware.JwtAuth(), app.GroupMessageCreateView)
	router.POST("social/groups/:id/members", middleware.JwtAuth(), app.GroupMemberAddView)
	router.PUT("social/groups/:id/members/:user_id/role", middleware.JwtAuth(), app.GroupMemberRoleUpdateView)
	router.DELETE("social/groups/:id/members/:user_id", middleware.JwtAuth(), app.GroupMemberRemoveView)
	router.PUT("social/groups/:id/transfer-owner", middleware.JwtAuth(), app.GroupTransferOwnerView)
	router.PUT("social/presence", middleware.JwtAuth(), app.PresenceUpdateView)
	router.POST("social/files", middleware.JwtAuth(), app.FileUploadView)
	router.GET("social/files/:id/download", middleware.JwtAuth(), app.FileDownloadView)
	router.GET("social/ws", app.SocketView)
}
