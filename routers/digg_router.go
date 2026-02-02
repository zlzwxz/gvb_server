package routers

import (
	"gvb-server/api"
)

func (router RouterGroup) DiggRouter() {
	diggApp := api.ApiGroupApp.DiggApi
	router.POST("article/digg", diggApp.DiggArticleView)
}
