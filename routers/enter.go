package routers

import (
	"net/http"

	"gvb-server/global"
	"gvb-server/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

// RouterGroup 把 *gin.RouterGroup 包一层，方便在方法上按业务模块拆分路由注册逻辑。
type RouterGroup struct {
	*gin.RouterGroup
}

// InitRouter 初始化整个 Gin 引擎。
// 这里集中做三件事：设置运行模式、挂公共中间能力、注册所有业务路由。
func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.Env)

	router := gin.Default()
	router.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	router.StaticFS("uploads", http.Dir("uploads"))

	apiRouterGroup := router.Group("/api")
	apiRouterGroup.Use(middleware.OperationAudit())
	routerGroupApp := RouterGroup{RouterGroup: apiRouterGroup}
	routerGroupApp.SettinsRouter()
	routerGroupApp.ImageRouter()
	routerGroupApp.FileRouter()
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
