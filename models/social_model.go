package models

import "time"

type SocialConversationType string

const (
	SocialConversationDirect SocialConversationType = "direct"
	SocialConversationGroup  SocialConversationType = "group"
)

type SocialMessageType string

const (
	SocialMessageText   SocialMessageType = "text"
	SocialMessageFile   SocialMessageType = "file"
	SocialMessageSystem SocialMessageType = "system"
	SocialMessageCall   SocialMessageType = "call"
)

type SocialGroupRole string

const (
	SocialGroupRoleOwner  SocialGroupRole = "owner"
	SocialGroupRoleAdmin  SocialGroupRole = "admin"
	SocialGroupRoleMember SocialGroupRole = "member"
)

type SocialCallStatus string

const (
	SocialCallStatusRinging   SocialCallStatus = "ringing"
	SocialCallStatusRejected  SocialCallStatus = "rejected"
	SocialCallStatusMissed    SocialCallStatus = "missed"
	SocialCallStatusCompleted SocialCallStatus = "completed"
	SocialCallStatusCanceled  SocialCallStatus = "canceled"
)

// UserFollowModel 记录关注关系；互相关注即视为好友。
type UserFollowModel struct {
	MODEL
	UserID       uint `gorm:"index:idx_social_follow,unique;not null" json:"user_id"`
	FollowUserID uint `gorm:"index:idx_social_follow,unique;not null" json:"follow_user_id"`
}

// UserBlockModel 记录拉黑关系。
type UserBlockModel struct {
	MODEL
	UserID      uint   `gorm:"index:idx_social_block,unique;not null" json:"user_id"`
	BlockUserID uint   `gorm:"index:idx_social_block,unique;not null" json:"block_user_id"`
	Reason      string `gorm:"size:120" json:"reason"`
}

// UserPresenceModel 保存在线状态设置与最后活跃时间。
type UserPresenceModel struct {
	MODEL
	UserID       uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	Mode         string     `gorm:"size:20;default:online" json:"mode"`
	StatusText   string     `gorm:"size:120" json:"status_text"`
	IsInvisible  bool       `gorm:"default:false" json:"is_invisible"`
	LastActiveAt *time.Time `json:"last_active_at"`
}

// SocialGroupModel 表示好友群组。
type SocialGroupModel struct {
	MODEL
	OwnerUserID uint   `gorm:"index;not null" json:"owner_user_id"`
	GroupNo     string `gorm:"size:24;uniqueIndex" json:"group_no"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Avatar      string `gorm:"size:255" json:"avatar"`
	Notice      string `gorm:"size:255" json:"notice"`
}

// SocialGroupMemberModel 表示群成员。
type SocialGroupMemberModel struct {
	MODEL
	GroupID  uint   `gorm:"index:idx_social_group_user,unique;not null" json:"group_id"`
	UserID   uint   `gorm:"index:idx_social_group_user,unique;not null" json:"user_id"`
	Role     string `gorm:"size:20;default:member" json:"role"`
	NickName string `gorm:"size:64" json:"nick_name"`
	Avatar   string `gorm:"size:255" json:"avatar"`
}

// SocialFileModel 保存好友/群聊文件。
type SocialFileModel struct {
	MODEL
	UserID uint   `gorm:"index;not null" json:"user_id"`
	Name   string `gorm:"size:255;not null" json:"name"`
	Path   string `gorm:"size:512;not null" json:"path"`
	Size   int64  `json:"size"`
	Ext    string `gorm:"size:32" json:"ext"`
	Mime   string `gorm:"size:128" json:"mime"`
	Hash   string `gorm:"size:64;index" json:"hash"`
}

// SocialMessageModel 统一承载单聊与群聊消息。
type SocialMessageModel struct {
	MODEL
	ConversationKey  string `gorm:"size:128;index;not null" json:"conversation_key"`
	ConversationType string `gorm:"size:20;index;not null" json:"conversation_type"`
	GroupID          uint   `gorm:"index" json:"group_id"`

	SendUserID       uint   `gorm:"index;not null" json:"send_user_id"`
	SendUserNickName string `gorm:"size:64" json:"send_user_nick_name"`
	SendUserAvatar   string `gorm:"size:255" json:"send_user_avatar"`

	ReceiveUserID       uint   `gorm:"index" json:"receive_user_id"`
	ReceiveUserNickName string `gorm:"size:64" json:"receive_user_nick_name"`
	ReceiveUserAvatar   string `gorm:"size:255" json:"receive_user_avatar"`

	MsgType       string     `gorm:"size:20;index;not null" json:"msg_type"`
	Content       string     `gorm:"type:text" json:"content"`
	FileID        uint       `gorm:"index" json:"file_id"`
	FileName      string     `gorm:"size:255" json:"file_name"`
	FileSize      int64      `json:"file_size"`
	FileMime      string     `gorm:"size:128" json:"file_mime"`
	FileURL       string     `gorm:"size:512" json:"file_url"`
	RelatedCallID string     `gorm:"size:64;index" json:"related_call_id"`
	IsRecalled    bool       `gorm:"default:false" json:"is_recalled"`
	RecalledBy    uint       `gorm:"index" json:"recalled_by"`
	RecalledAt    *time.Time `json:"recalled_at"`
	Extra         string     `gorm:"type:text" json:"extra"`
}

// SocialConversationReadModel 记录用户对会话的已读游标。
type SocialConversationReadModel struct {
	MODEL
	ConversationKey   string     `gorm:"size:128;index:idx_social_read,unique;not null" json:"conversation_key"`
	UserID            uint       `gorm:"index:idx_social_read,unique;not null" json:"user_id"`
	LastReadMessageID uint       `gorm:"default:0" json:"last_read_message_id"`
	LastReadAt        *time.Time `json:"last_read_at"`
}

// SocialCallLogModel 保存好友之间的语音通话记录。
type SocialCallLogModel struct {
	MODEL
	CallID             string     `gorm:"size:64;uniqueIndex;not null" json:"call_id"`
	ConversationKey    string     `gorm:"size:128;index;not null" json:"conversation_key"`
	ConversationType   string     `gorm:"size:20;index;not null" json:"conversation_type"`
	CallerUserID       uint       `gorm:"index;not null" json:"caller_user_id"`
	CallerNickName     string     `gorm:"size:64" json:"caller_nick_name"`
	CallerAvatar       string     `gorm:"size:255" json:"caller_avatar"`
	CalleeUserID       uint       `gorm:"index;not null" json:"callee_user_id"`
	CalleeNickName     string     `gorm:"size:64" json:"callee_nick_name"`
	CalleeAvatar       string     `gorm:"size:255" json:"callee_avatar"`
	Status             string     `gorm:"size:20;index;not null" json:"status"`
	StartedAt          *time.Time `json:"started_at"`
	AnsweredAt         *time.Time `json:"answered_at"`
	EndedAt            *time.Time `json:"ended_at"`
	DurationSec        int        `json:"duration_sec"`
	MissedByUserID     uint       `gorm:"index" json:"missed_by_user_id"`
	RelatedMessageID   uint       `gorm:"index" json:"related_message_id"`
	LastOperatorUserID uint       `gorm:"index" json:"last_operator_user_id"`
}
