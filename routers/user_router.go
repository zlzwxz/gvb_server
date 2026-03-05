package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

var store = cookie.NewStore([]byte("HyvCD89g3VDJ9646BFGEh37GFJ"))

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
	router.POST("/user_create", app.UserCreateView)
	//qq登录正式地址
	router.POST("/qq_login", app.QQLoginView)
	//用户信息
	router.GET("/user_info", middleware.JwtAuth(), app.UserInfoView)
	//修改用户昵称，签名，链接
	router.PUT("/user_update_nick_name", middleware.JwtAuth(), app.UserUpdateNickName)
	//获取qq登录的跳转链接
	router.GET("/qq_login_path", app.QQLoginLinkView)
}
