package models

type ArticleFileModel struct {
	MODEL
	UserID       uint   `gorm:"index;not null" json:"user_id"`
	UserNickName string `gorm:"size:64" json:"user_nick_name"`
	Name         string `gorm:"size:255;not null" json:"name"`
	Hash         string `gorm:"size:64;index" json:"hash"`
	Path         string `gorm:"size:512;not null" json:"path"`
	Size         int64  `json:"size"`
	Ext          string `gorm:"size:20" json:"ext"`
}
