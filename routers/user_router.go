package routers

import (
	"gvb-server/api"
	"gvb-server/middleware"
)

func (router RouterGroup) UserRouter() {
	app := api.ApiGroupApp.UserApi
	router.POST("/email_login", app.EmailLoginView)
	router.GET("/users", middleware.JwtAuth(), app.UserListView)
	router.PUT("/user_role", middleware.JwtAuth(), app.UserUpdateRoleView)
	router.PUT("/user_password", middleware.JwtAdmin(), app.UserUpdatePassword)
	router.POST("/logout", app.LogoutView)
}
