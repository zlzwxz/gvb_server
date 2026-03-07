package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// BoardRouter 注册板块管理路由。
func (router RouterGroup) BoardRouter() {
	boardApp := api.ApiGroupApp.BoardApi
	router.GET("boards", boardApp.BoardListView)
	router.POST("boards", middleware.JwtAdmin(), boardApp.BoardCreateView)
	router.PUT("boards", middleware.JwtAdmin(), boardApp.BoardUpdateView)
	router.DELETE("boards", middleware.JwtAdmin(), boardApp.BoardRemoveView)
}
