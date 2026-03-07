package api

import (
	"gvb-server/api/advert_api"
	"gvb-server/api/announcement_api"
	"gvb-server/api/article_api"
	"gvb-server/api/board_api"
	"gvb-server/api/chat_api"
	"gvb-server/api/comment_api"
	"gvb-server/api/data_api"
	"gvb-server/api/digg_api"
	"gvb-server/api/file_api"
	"gvb-server/api/images_api"
	"gvb-server/api/log_api"
	"gvb-server/api/menu_api"
	"gvb-server/api/message_api"
	"gvb-server/api/new_api"
	"gvb-server/api/settings_api"
	"gvb-server/api/social_api"
	"gvb-server/api/tag_api"
	"gvb-server/api/user_api"
)

// ApiGroup 统一收口所有 API handler 分组。
// 路由层只需要依赖这个聚合对象，就能按模块拿到对应的控制器。
//
// 这样做的好处是：
// 1. 路由层不需要一个个 import 具体 handler 文件；
// 2. 各模块入口统一放在一个总结构体里，项目结构更稳定；
// 3. 新增模块时，只需要在这里补一个字段，再去路由层注册即可。
type ApiGroup struct {
	AnnouncementApi announcement_api.AnnouncementApi
	SettingsApi     settings_api.SettingsApi
	ImagesApi       images_api.ImagesApi
	AdvertApi       advert_api.AdvertApi
	MenuApi         menu_api.MenuApi
	BoardApi        board_api.BoardApi
	UserApi         user_api.UserApi
	SocialApi       social_api.SocialApi
	TagApi          tag_api.TagApi
	MessageApi      message_api.MessageApi
	ArticleApi      article_api.ArticleApi
	DiggApi         digg_api.DiggApi
	CommentApi      comment_api.CommentApi
	NewApi          new_api.NewApi
	ChatApi         chat_api.ChatApi
	LogApi          log_api.LogApi
	DataApi         data_api.DataApi
	FileApi         file_api.FileApi
}

// ApiGroupApp 是全局单例，路由注册时直接复用这一份 handler 聚合对象。
var ApiGroupApp = new(ApiGroup)
