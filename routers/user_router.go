package routers

import (
	"gvb-server/api"
	"gvb-server/global"
	"gvb-server/middleware"
	"strings"
	"sync"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

var (
	sessionStore     sessions.Store
	sessionStoreOnce sync.Once
)

func getSessionStore() sessions.Store {
	sessionStoreOnce.Do(func() {
		secret := strings.TrimSpace(global.Config.Jwt.Secret)
		if secret == "" {
			secret = "gvb_fallback_session_secret_change_me"
		}
		sessionStore = cookie.NewStore([]byte(secret))
	})
	return sessionStore
}

// UserRouter 注册用户认证与用户资料相关路由。
func (router RouterGroup) UserRouter() {
	app := api.ApiGroupApp.UserApi
	router.Use(sessions.Sessions("sessionid", getSessionStore()))
	router.POST("/email_login", app.EmailLoginView)
	router.GET("/users", middleware.JwtAuth(), app.UserListView)
	router.PUT("/user_role", middleware.JwtAdmin(), app.UserUpdateRoleView)
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
	router.GET("/users/:id/profile", app.UserPublicProfileView)
	router.GET("/users/:id/space/posts", app.UserSpacePostListView)
	router.GET("/users/:id/space/messages", app.UserSpaceMessageListView)
	router.POST("/space/posts", middleware.JwtAuth(), app.UserSpacePostCreateView)
	router.DELETE("/space/posts/:id", middleware.JwtAuth(), app.UserSpacePostRemoveView)
	router.POST("/space/messages", middleware.JwtAuth(), app.UserSpaceMessageCreateView)
	router.DELETE("/space/messages/:id", middleware.JwtAuth(), app.UserSpaceMessageRemoveView)
}
