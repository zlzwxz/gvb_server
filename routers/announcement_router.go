package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// AnnouncementRouter 注册公告模块路由。
func (router RouterGroup) AnnouncementRouter() {
	announcementApp := api.ApiGroupApp.AnnouncementApi
	router.GET("announcements", announcementApp.AnnouncementListView)
	router.GET("announcements/manage", middleware.JwtAdmin(), announcementApp.AnnouncementManageListView)
	router.POST("announcements", middleware.JwtAdmin(), announcementApp.AnnouncementCreateView)
	router.PUT("announcements/:id", middleware.JwtAdmin(), announcementApp.AnnouncementUpdateView)
	router.DELETE("announcements", middleware.JwtAdmin(), announcementApp.AnnouncementRemoveView)
}
