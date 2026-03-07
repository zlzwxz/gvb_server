package log_stash

import "time"

type LogStashModel struct {
	ID           uint      `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt    time.Time `json:"created_at"`           // 创建时间
	IP           string    `gorm:"size:32" json:"ip"`
	Addr         string    `gorm:"size:64" json:"addr"`
	Level        Level     `gorm:"size:4" json:"level"`     // 日志的等级
	Content      string    `gorm:"size:255" json:"content"` // 日志消息内容
	UserID       uint      `json:"user_id"`                 // 登录用户的用户id，需要自己在查询的时候做关联查询
	Method       string    `gorm:"size:16" json:"method"`   // 请求方法
	Path         string    `gorm:"size:180" json:"path"`    // 接口地址
	StatusCode   int       `json:"status_code"`             // HTTP 状态码
	RespCode     int       `json:"resp_code"`               // 业务状态码
	RequestBody  string    `gorm:"type:text" json:"request_body"`
	ResponseBody string    `gorm:"type:text" json:"response_body"`
}
