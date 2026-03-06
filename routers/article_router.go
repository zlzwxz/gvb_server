package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

// ArticleRouter 注册文章模块路由。
// 需要登录的写操作统一挂 JwtAuth，中后台扩展接口继续沿用同一套路由分组。
func (router RouterGroup) ArticleRouter() {
	articleApp := api.ApiGroupApp.ArticleApi
	router.POST("articles", middleware.JwtAuth(), articleApp.ArticleCreateView)
	router.GET("articles", articleApp.ArticleListView)
	router.PUT("articles/review", middleware.JwtAdmin(), articleApp.ArticleReviewView)
	router.GET("articles/detail", articleApp.ArticleDetailByTitleView)
	router.GET("articles/calendar", articleApp.ArticleCalendarView)
	router.GET("articles/tags", articleApp.ArticleTagListView)
	router.GET("articles/:id", articleApp.ArticleDetailView)
	router.PUT("articles", middleware.JwtAuth(), articleApp.ArticleUpdateView)
	router.DELETE("articles", middleware.JwtAuth(), articleApp.ArticleRemoveView)
	router.POST("articles/collects", middleware.JwtAuth(), articleApp.ArticleCollCreateView)
	router.GET("articles/collects", middleware.JwtAuth(), articleApp.ArticleCollListView)
	router.DELETE("articles/collects/batch", middleware.JwtAuth(), articleApp.ArticleCollBatchRemoveView)
	router.GET("articles/collects/manage", middleware.JwtAuth(), articleApp.ArticleCollManageListView)
	router.DELETE("articles/collects/manage", middleware.JwtAuth(), articleApp.ArticleCollManageBatchRemoveView)
	router.GET("articles/insights", articleApp.ArticleInsightsView)
	router.GET("article/text", articleApp.FullTextSearchView)
	router.GET("articles/categorys", articleApp.ArticleCategoryListView)
	router.GET("articles/content/:id", articleApp.ArticleContentByIDView)
}
