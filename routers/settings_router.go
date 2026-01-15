package routers

import (
	"gvb-server/api"
)

/*func SettinsRouter(router *gin.Engine) {
	srttinsApi := api.ApiGroupApp.SettingsApi
	router.GET("/", srttinsApi.SettingsInfoView)
}*/

func (router RouterGroup) SettinsRouter() {
	srttinsApi := api.ApiGroupApp.SettingsApi
	router.GET("/settings/:name", srttinsApi.SettingsInfoView)
	router.PUT("/settings/:name", srttinsApi.SettingsInfoUpdateView)
}
