package models

import "gvb-server/models/ctype"

// BoardModel 论坛板块表。
// 通过版主/副版主列表控制板块内的管理权限。
type BoardModel struct {
	MODEL
	Name               string      `gorm:"size:32;uniqueIndex" json:"name"`
	Slug               string      `gorm:"size:32;uniqueIndex" json:"slug"`
	Description        string      `gorm:"size:255" json:"description"`
	Notice             string      `gorm:"size:600" json:"notice"`
	Rules              string      `gorm:"size:1200" json:"rules"`
	PinnedArticleIDs   ctype.Array `gorm:"type:text" json:"pinned_article_ids"`
	ModeratorIDs       ctype.Array `gorm:"type:text" json:"moderator_ids"`
	DeputyModeratorIDs ctype.Array `gorm:"type:text" json:"deputy_moderator_ids"`
	Sort               int         `gorm:"default:0" json:"sort"`
	IsEnabled          bool        `gorm:"default:true" json:"is_enabled"`
}
