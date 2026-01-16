package api

import (
	"gvb-server/api/advert_api"
	"gvb-server/api/images_api"
	"gvb-server/api/settings_api"
)

type ApiGroup struct {
	SettingsApi settings_api.SettingsApi
	ImagesApi   images_api.ImagesApi
	AdvertApi   advert_api.AdvertApi
}

// 实例化对象
var ApiGroupApp = new(ApiGroup)
