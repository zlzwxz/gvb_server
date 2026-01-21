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
	router.POST("/email_login", app.EmailLoginView)
	router.GET("/users", middleware.JwtAuth(), app.UserListView)
	router.PUT("/user_role", middleware.JwtAuth(), app.UserUpdateRoleView)
	router.PUT("/user_password", middleware.JwtAdmin(), app.UserUpdatePassword)
	router.POST("/logout", middleware.JwtAuth(), app.LogoutView)
	router.DELETE("/users", middleware.JwtAdmin(), app.UserRemoveView)
	router.Use(sessions.Sessions("sessionid", store))
	router.POST("user_bind_email", middleware.JwtAuth(), app.UserBindEmailView)
}
