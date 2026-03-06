package routers

import "gvb-server/api"

// SettinsRouter 注册系统配置相关路由。
// 方法名沿用旧拼写，避免影响现有调用代码。
func (router RouterGroup) SettinsRouter() {
	srttinsApi := api.ApiGroupApp.SettingsApi
	router.GET("/settings/:name", srttinsApi.SettingsInfoView)
	router.PUT("/settings/:name", srttinsApi.SettingsInfoUpdateView)
}
