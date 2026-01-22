package api

import (
	"gvb-server/api/advert_api"
	"gvb-server/api/images_api"
	"gvb-server/api/menu_api"
	"gvb-server/api/message_api"
	"gvb-server/api/settings_api"
	"gvb-server/api/tag_api"
	"gvb-server/api/user_api"
)

type ApiGroup struct {
	SettingsApi settings_api.SettingsApi
	ImagesApi   images_api.ImagesApi
	AdvertApi   advert_api.AdvertApi
	MenuApi     menu_api.MenuApi
	UserApi     user_api.UserApi
	TagApi      tag_api.TagApi
	MessageApi  message_api.MessageApi
}

// 实例化对象
var ApiGroupApp = new(ApiGroup)
