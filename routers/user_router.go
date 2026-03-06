package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

// store 用于 QQ 登录等依赖 session 的场景。
var store = cookie.NewStore([]byte("HyvCD89g3VDJ9646BFGEh37GFJ"))

// UserRouter 注册用户认证与用户资料相关路由。
func (router RouterGroup) UserRouter() {
	app := api.ApiGroupApp.UserApi
	router.Use(sessions.Sessions("sessionid", store))
	router.POST("/email_login", app.EmailLoginView)
	router.GET("/users", middleware.JwtAuth(), app.UserListView)
	router.PUT("/user_role", middleware.JwtAuth(), app.UserUpdateRoleView)
	router.PUT("/user_password", middleware.JwtAdmin(), app.UserUpdatePassword)
	router.POST("/logout", middleware.JwtAuth(), app.LogoutView)
	router.DELETE("/users", middleware.JwtAdmin(), app.UserRemoveView)
	router.POST("/user_bind_email", middleware.JwtAuth(), app.UserBindEmailView)
	router.POST("/user_register_email_code", app.UserRegisterEmailCodeView)
	router.POST("/user_create", app.UserCreateView)
	router.POST("/qq_login", app.QQLoginView)
	router.GET("/user_info", middleware.JwtAuth(), app.UserInfoView)
	router.PUT("/user_update_nick_name", middleware.JwtAuth(), app.UserUpdateNickName)
	router.POST("/user_check_in", middleware.JwtAuth(), app.UserCheckInView)
	router.GET("/user_check_in_status", middleware.JwtAuth(), app.UserCheckInStatusView)
	router.GET("/qq_login_path", app.QQLoginLinkView)
	router.GET("/user_level_rank", app.UserLevelRankView)
}
