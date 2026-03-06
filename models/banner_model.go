package models

import (
	"os"

	"gorm.io/gorm"
	"gvb-server/global"
	"gvb-server/models/ctype"
)

// BannerModel 记录上传后的图片素材。
// 菜单轮播图、文章封面等场景都会复用这一张表。
type BannerModel struct {
	MODEL
	Path      string          `json:"path"`                        // 图片访问路径
	Hash      string          `json:"hash"`                        // 图片 hash，用于判重
	Name      string          `gorm:"size:38" json:"name"`         // 图片名称
	ImageType ctype.ImageType `gorm:"default:1" json:"image_type"` // 图片来源类型：本地或云存储
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
