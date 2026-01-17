package routers

import (
	"gvb-server/global"

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
	apiRouterGroup := router.Group("/api")
	routerGroupApp := RouterGroup{
		apiRouterGroup,
	}
	routerGroupApp.SettinsRouter()
	routerGroupApp.ImageRouter()
	routerGroupApp.AdvertRouter()
	routerGroupApp.MenuRouter()
	return router
}
