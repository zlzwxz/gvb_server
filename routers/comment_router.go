package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

func (router RouterGroup) CommentRouter() {
	commentApp := api.ApiGroupApp.CommentApi
	router.POST("comments", middleware.JwtAuth(), commentApp.CommentCreateView)
	router.GET("comments", commentApp.CommentListView)
	router.GET("comments/:id", commentApp.CommentDigg)
	router.DELETE("comments/:id", commentApp.CommentRemoveView)
}
