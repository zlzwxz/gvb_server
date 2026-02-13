package routers

import (
	"gvb-server/api"
)

func (router RouterGroup) NewRouter() {
	newApp := api.ApiGroupApp.NewApi
	router.GET("/news", newApp.NewListView)
}
