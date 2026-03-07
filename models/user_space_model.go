package models

import (
	"database/sql/driver"
	"encoding/json"
)

type SpaceAttachment struct {
	FileID uint   `json:"file_id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Size   int64  `json:"size"`
}

type SpaceAttachmentList []SpaceAttachment

func (list *SpaceAttachmentList) Scan(value interface{}) error {
	if value == nil {
		*list = SpaceAttachmentList{}
		return nil
	}

	switch data := value.(type) {
	case []byte:
		if len(data) == 0 {
			*list = SpaceAttachmentList{}
			return nil
		}
		return json.Unmarshal(data, list)
	case string:
		if data == "" {
			*list = SpaceAttachmentList{}
			return nil
		}
		return json.Unmarshal([]byte(data), list)
	default:
		*list = SpaceAttachmentList{}
		return nil
	}
}

func (list SpaceAttachmentList) Value() (driver.Value, error) {
	if len(list) == 0 {
		return "[]", nil
	}
	byteData, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return string(byteData), nil
}

type UserSpacePostModel struct {
	MODEL
	UserID       uint                `gorm:"index" json:"user_id"`
	UserNickName string              `gorm:"size:36" json:"user_nick_name"`
	UserAvatar   string              `gorm:"size:256" json:"user_avatar"`
	Content      string              `gorm:"type:text" json:"content"`
	Attachments  SpaceAttachmentList `gorm:"type:json" json:"attachments"`
	IsPrivate    bool                `gorm:"default:false" json:"is_private"`
}

type UserSpaceMessageModel struct {
	MODEL
	SpaceUserID  uint   `gorm:"index" json:"space_user_id"`
	UserID       uint   `gorm:"index" json:"user_id"`
	UserNickName string `gorm:"size:36" json:"user_nick_name"`
	UserAvatar   string `gorm:"size:256" json:"user_avatar"`
	Content      string `gorm:"type:text" json:"content"`
	IsPrivate    bool   `gorm:"default:false" json:"is_private"`
}
