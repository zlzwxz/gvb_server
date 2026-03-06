package models

import (
	"gvb-server/models/ctype"
	"time"
)

// UserModel 用户表
type UserModel struct {
	MODEL
	NickName      string           `gorm:"size:36" json:"nick_name,select(c),select(info)"`            // 昵称
	UserName      string           `gorm:"size:36" json:"user_name,select(c),select(info)"`            // 用户名
	Password      string           `gorm:"size:128" json:"-"`                                          // 密码
	Avatar        string           `gorm:"size:256" json:"avatar,select(c),select(info)"`              // 头像id
	Email         string           `gorm:"size:128" json:"email,select(info)"`                         // 邮箱
	Tel           string           `gorm:"size:18" json:"tel,select(info)"`                            // 手机号
	Addr          string           `gorm:"size:64" json:"addr,select(c),select(info)"`                 // 地址
	Token         string           `gorm:"size:64" json:"token"`                                       // 其他平台的唯一id
	IP            string           `gorm:"size:20" json:"ip,select(c),select(info)"`                   // ip地址
	Role          ctype.Role       `gorm:"size:4;default:1" json:"role,select(info)"`                  // 权限  1 管理员  2 普通用户  3 游客
	SignStatus    ctype.SignStatus `gorm:"type=smallint(6)" json:"sign_status,select(info)"`           // 注册来源
	Sign          string           `gorm:"size:20" json:"sign,select(info),select(c)" structs:"sign"`  // 签名
	Link          string           `gorm:"size:128" json:"link,select(info),select(c)" structs:"link"` // 链接
	Points        int              `gorm:"default:0" json:"points,select(info),select(c)"`             // 积分
	Experience    int              `gorm:"default:0" json:"experience,select(info),select(c)"`         // 经验
	Level         int              `gorm:"default:1" json:"level,select(info),select(c)"`              // 等级
	CheckInStreak int              `gorm:"default:0" json:"check_in_streak,select(info),select(c)"`    // 连续签到天数
	LastCheckInAt *time.Time       `json:"last_check_in_at,select(info),select(c)"`                    // 上次签到时间
}
