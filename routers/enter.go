package routers

import (
	"github.com/gin-gonic/gin"
	"gvb-server/global"
)

type RouterGroup struct {
	*gin.RouterGroup
}

// 路由初始化
func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.Env)
	router := gin.Default()
	apiRouterGroup := router.Group("/api")
	routerGroupApp := RouterGroup{
		apiRouterGroup,
	}

	routerGroupApp.SettinsRouter()
	routerGroupApp.ImageRouter()
	return router
}
