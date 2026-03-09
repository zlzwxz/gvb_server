package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
	"gvb-server/global"
	"gvb-server/models/ctype"
)

const (
	ImageVisibilityPublic  = "public"
	ImageVisibilityPrivate = "private"
)

// BannerModel 记录上传后的图片素材。
// 菜单轮播图、文章封面等场景都会复用这一张表。
type BannerModel struct {
	MODEL
	UserID        uint            `gorm:"index" json:"user_id"`
	Path          string          `json:"path"`                                       // 图片访问路径
	Hash          string          `json:"hash"`                                       // 图片 hash，用于判重
	Name          string          `gorm:"size:64" json:"name"`                        // 图片名称
	ImageType     ctype.ImageType `gorm:"default:1" json:"image_type"`                // 图片来源类型：本地或云存储
	ImageCategory string          `gorm:"size:32;index" json:"image_category"`
	Visibility    string          `gorm:"size:16;default:public;index" json:"visibility"`
	SourceURL     string          `gorm:"size:255" json:"source_url"`
}

func NormalizeImageVisibility(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ImageVisibilityPrivate:
		return ImageVisibilityPrivate
	default:
		return ImageVisibilityPublic
	}
}

func (b BannerModel) IsPrivate() bool {
	return NormalizeImageVisibility(b.Visibility) == ImageVisibilityPrivate
}

func (b BannerModel) MatchesOwner(userID uint, nickName string) bool {
	if userID == 0 {
		return false
	}
	if b.UserID > 0 && b.UserID == userID {
		return true
	}
	normalizedPath := filepath.ToSlash(strings.TrimSpace(b.Path))
	if normalizedPath == "" {
		return false
	}
	if strings.Contains(normalizedPath, fmt.Sprintf("/uploads/file/u_%d/", userID)) {
		return true
	}
	nickName = strings.TrimSpace(nickName)
	if nickName == "" {
		return false
	}
	legacyPrefix := "/uploads/file/" + filepath.ToSlash(strings.Trim(nickName, "/")) + "/"
	return strings.HasPrefix(normalizedPath, legacyPrefix)
}

// BeforeDelete 在删除图片记录前，顺手清理本地磁盘文件。
// 这样数据库和文件系统就不容易出现“数据库删了，但磁盘还留着垃圾文件”的不一致问题。
func (b *BannerModel) BeforeDelete(tx *gorm.DB) (err error) {
	if b.ImageType == ctype.Local {
		err = os.Remove(b.Path)
		if err != nil {
			global.Log.Error(err)
			return err
		}
	}
	return nil
}
