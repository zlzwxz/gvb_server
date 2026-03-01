package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

func (router RouterGroup) ArticleRouter() {
	articleApp := api.ApiGroupApp.ArticleApi
	router.POST("articles", middleware.JwtAuth(), articleApp.ArticleCreateView)
	router.GET("articles", articleApp.ArticleListView)
	//http://127.0.0.1:8080/api/articles/_WoX6ZsBPMPoP5eKRYpO
	router.GET("articles/:id", articleApp.ArticleDetailView)
	//http://127.0.0.1:8080/api/articles/detail/?title=gogo111golan1111111222nd
	router.GET("articles/detail", articleApp.ArticleDetailByTitleView)
	//文章日历
	router.GET("articles/calendar", articleApp.ArticleCalendarView)
	//文章标签
	router.GET("articles/tags", articleApp.ArticleTagListView)
	//修改文章
	router.PUT("articles", articleApp.ArticleUpdateView)
	//删除文章
	router.DELETE("articles", articleApp.ArticleRemoveView)
	//文章收藏
	router.POST("articles/collects", middleware.JwtAuth(), articleApp.ArticleCollCreateView)
	//文章收藏列表
	router.GET("articles/collects", middleware.JwtAuth(), articleApp.ArticleCollListView)
	//文章收藏批量删除
	router.DELETE("articles/collects/batch", middleware.JwtAuth(), articleApp.ArticleCollBatchRemoveView)
	//全文文章搜索
	router.GET("article/text", articleApp.FullTextSearchView)
	//文章分类列表
	router.GET("articles/categorys", articleApp.ArticleCategoryListView)
	//根据文章id获取文章正文
	router.GET("articles/content/:id", articleApp.ArticleContentByIDView)
}
