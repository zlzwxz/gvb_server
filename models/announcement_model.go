package models

import "time"

// AnnouncementModel 站内公告。
// 用于首页公告栏、板块运营公告和后台统一公告管理。
type AnnouncementModel struct {
	MODEL
	Title    string     `gorm:"size:120;index" json:"title"`
	Content  string     `gorm:"type:text" json:"content"`
	Level    string     `gorm:"size:16;default:info" json:"level"`
	JumpLink string     `gorm:"size:255" json:"jump_link"`
	BoardID  uint       `gorm:"default:0;index" json:"board_id"`
	Sort     int        `gorm:"default:0" json:"sort"`
	IsShow   bool       `gorm:"default:true;index" json:"is_show"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}
