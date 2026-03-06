package models

// MessageModel 记录点对点私信。
// 发送人和接收人都冗余保存昵称、头像，是为了在列表查询时少做联表也能直接展示。
type MessageModel struct {
	MODEL
	SendUserID       uint      `gorm:"primaryKey" json:"send_user_id"` // 发送人 ID
	SendUserModel    UserModel `gorm:"foreignKey:SendUserID" json:"-"`
	SendUserNickName string    `gorm:"size:42" json:"send_user_nick_name"`
	SendUserAvatar   string    `json:"send_user_avatar"`

	RevUserID       uint      `gorm:"primaryKey" json:"rev_user_id"` // 接收人 ID
	RevUserModel    UserModel `gorm:"foreignKey:RevUserID" json:"-"`
	RevUserNickName string    `gorm:"size:42" json:"rev_user_nick_name"`
	RevUserAvatar   string    `json:"rev_user_avatar"`
	IsRead          bool      `gorm:"default:false" json:"is_read"` // 接收方是否已读
	Content         string    `json:"content"`                      // 消息正文
}
