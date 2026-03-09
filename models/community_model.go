package models

import (
	"strings"
	"time"

	"gvb-server/models/ctype"
)

const (
	CommunityScenePlaza  = "plaza"
	CommunitySceneBounty = "bounty"
)

const (
	CommunityStatusPublished  = "published"
	CommunityStatusOpen       = "open"
	CommunityStatusInProgress = "in_progress"
	CommunityStatusResolved   = "resolved"
	CommunityStatusClosed     = "closed"
)

type CommunityPostModel struct {
	MODEL
	Scene              string              `gorm:"size:16;index" json:"scene"`
	UserID             uint                `gorm:"index" json:"user_id"`
	UserNickName       string              `gorm:"size:36" json:"user_nick_name"`
	UserAvatar         string              `gorm:"size:256" json:"user_avatar"`
	Title              string              `gorm:"size:120;index" json:"title"`
	Summary            string              `gorm:"size:255" json:"summary"`
	Content            string              `gorm:"type:longtext" json:"content"`
	Category           string              `gorm:"size:32;index" json:"category"`
	Tags               ctype.Array         `json:"tags"`
	CoverImage         string              `gorm:"size:255" json:"cover_image"`
	Attachments        SpaceAttachmentList `gorm:"type:json" json:"attachments"`
	Status             string              `gorm:"size:20;index" json:"status"`
	Budget             int                 `gorm:"default:0" json:"budget"`
	RewardUnit         string              `gorm:"size:16;default:points" json:"reward_unit"`
	Deadline           *time.Time          `json:"deadline"`
	AcceptedUserID     uint                `gorm:"index" json:"accepted_user_id"`
	AcceptedUserNick   string              `gorm:"size:36" json:"accepted_user_nick"`
	ReplyCount         int                 `gorm:"default:0" json:"reply_count"`
	ViewCount          int                 `gorm:"default:0" json:"view_count"`
	LastReplyAt        *time.Time          `json:"last_reply_at"`
	LastReplyNick      string              `gorm:"size:36" json:"last_reply_nick"`
	LastReplyPreview   string              `gorm:"size:255" json:"last_reply_preview"`
	IsPinned           bool                `gorm:"default:false;index" json:"is_pinned"`
}

type CommunityReplyModel struct {
	MODEL
	PostID        uint   `gorm:"index" json:"post_id"`
	ParentID      uint   `gorm:"index" json:"parent_id"`
	UserID        uint   `gorm:"index" json:"user_id"`
	UserNickName  string `gorm:"size:36" json:"user_nick_name"`
	UserAvatar    string `gorm:"size:256" json:"user_avatar"`
	Content       string `gorm:"type:text" json:"content"`
	IsOfficial    bool   `gorm:"default:false" json:"is_official"`
	QuotedUserID  uint   `gorm:"index" json:"quoted_user_id"`
	QuotedUserNick string `gorm:"size:36" json:"quoted_user_nick"`
}

func NormalizeCommunityScene(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case CommunitySceneBounty:
		return CommunitySceneBounty
	default:
		return CommunityScenePlaza
	}
}

func NormalizeCommunityStatus(scene string, value string) string {
	scene = NormalizeCommunityScene(scene)
	switch strings.ToLower(strings.TrimSpace(value)) {
	case CommunityStatusInProgress:
		if scene == CommunitySceneBounty {
			return CommunityStatusInProgress
		}
	case CommunityStatusResolved:
		return CommunityStatusResolved
	case CommunityStatusClosed:
		return CommunityStatusClosed
	case CommunityStatusPublished:
		return CommunityStatusPublished
	case CommunityStatusOpen:
		if scene == CommunitySceneBounty {
			return CommunityStatusOpen
		}
	}
	if scene == CommunitySceneBounty {
		return CommunityStatusOpen
	}
	return CommunityStatusPublished
}

func CommunityRewardUnitLabel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "cny", "yuan", "rmb", "cash", "money", "元":
		return "元"
	default:
		return "积分"
	}
}
