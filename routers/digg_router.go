package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// DiggRouter 注册文章点赞路由。
func (router RouterGroup) DiggRouter() {
	diggApp := api.ApiGroupApp.DiggApi
	router.POST("article/digg", middleware.JwtAuth(), diggApp.DiggArticleView)
}
