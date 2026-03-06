package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// CommentRouter 注册评论相关路由。
func (router RouterGroup) CommentRouter() {
	commentApp := api.ApiGroupApp.CommentApi
	router.POST("comments", middleware.JwtAuth(), commentApp.CommentCreateView)
	router.GET("comments", commentApp.CommentListView)
	router.GET("comments/:id", commentApp.CommentDigg)
	router.DELETE("comments/:id", commentApp.CommentRemoveView)
}
