package models

type UserCheckInModel struct {
	MODEL
	UserID      uint   `gorm:"index:idx_user_check_in,priority:1;not null" json:"user_id"`
	CheckInDate string `gorm:"size:10;index:idx_user_check_in,priority:2,unique;not null" json:"check_in_date"`
	Points      int    `gorm:"default:0" json:"points"`
	Experience  int    `gorm:"default:0" json:"experience"`
	Streak      int    `gorm:"default:1" json:"streak"`
}
