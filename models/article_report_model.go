package models

import (
	"time"

	"gvb-server/models/ctype"
)

// ArticleReportModel 记录用户对文章的举报和后台处理结果。
// 当运营确认需要复审时，会把对应文章重新打回待审核队列。
type ArticleReportModel struct {
	MODEL
	ArticleID           string                    `gorm:"size:64;index" json:"article_id"`
	ArticleTitle        string                    `gorm:"size:255" json:"article_title"`
	BoardID             uint                      `gorm:"index" json:"board_id"`
	BoardName           string                    `gorm:"size:64" json:"board_name"`
	ReporterUserID      uint                      `gorm:"index" json:"reporter_user_id"`
	ReporterNickName    string                    `gorm:"size:64" json:"reporter_nick_name"`
	Reason              string                    `gorm:"size:64" json:"reason"`
	Content             string                    `gorm:"size:500" json:"content"`
	Status              ctype.ArticleReportStatus `gorm:"default:1;index" json:"status"`
	HandleNote          string                    `gorm:"size:255" json:"handle_note"`
	HandlerUserID       uint                      `json:"handler_user_id"`
	HandlerUserNickName string                    `gorm:"size:64" json:"handler_user_nick_name"`
	HandledAt           *time.Time                `json:"handled_at"`
}
