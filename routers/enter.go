package routers

import (
	"gvb-server/global"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger" //swagger包
)

type RouterGroup struct {
	*gin.RouterGroup
}

// 路由初始化
func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.Env)

	router := gin.Default()
	//启动swaggerweb网页路由
	router.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	//静态文件路由
	router.StaticFS("uploads", http.Dir("uploads"))
	//测试qq登录
	//router.GET("/login", user_api.UserApi{}.QQLoginView)
	apiRouterGroup := router.Group("/api")
	routerGroupApp := RouterGroup{
		apiRouterGroup,
	}
	routerGroupApp.SettinsRouter()
	routerGroupApp.ImageRouter()
	routerGroupApp.AdvertRouter()
	routerGroupApp.MenuRouter()
	routerGroupApp.UserRouter()
	routerGroupApp.TagRouter()
	routerGroupApp.MessageRouter()
	routerGroupApp.ArticleRouter()
	routerGroupApp.DiggRouter()
	routerGroupApp.CommentRouter()
	routerGroupApp.NewRouter()
	routerGroupApp.ChatRouter()
	routerGroupApp.LogRouter()
	routerGroupApp.DataRouter()
	return router
}
