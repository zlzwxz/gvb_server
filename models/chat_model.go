package models

import "gvb-server/models/ctype"

// ChatModel 保存聊天室里的历史消息。
// 群聊广播和系统提示消息都会落到这张表，方便后续回放历史记录。
type ChatModel struct {
	MODEL
	NickName string        `gorm:"size:15" json:"nick_name"`       // 发送者昵称
	Avatar   string        `gorm:"size:128" json:"avatar"`         // 发送者头像
	Content  string        `gorm:"size:256" json:"content"`        // 消息内容
	IP       string        `gorm:"size:32" json:"ip,omit(list)"`   // 发送者 IP
	Addr     string        `gorm:"size:64" json:"addr,omit(list)"` // 发送者归属地
	IsGroup  bool          `json:"is_group"`                       // 是否为群发消息
	MsgType  ctype.MsgType `gorm:"size:4" json:"msg_type"`         // 消息类型：文本、图片、系统提示等
}
