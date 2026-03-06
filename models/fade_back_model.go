package models

// FadeBackModel 记录用户反馈及后台回复内容。
type FadeBackModel struct {
	MODEL
	Email        string `gorm:"size:64" json:"email"`          // 反馈人的邮箱
	Content      string `gorm:"size:128" json:"content"`       // 用户反馈内容
	ApplyContent string `gorm:"size:128" json:"apply_content"` // 后台回复内容
	IsApply      bool   `json:"is_apply"`                      // 是否已经回复
}
