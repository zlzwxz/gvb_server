package cron_ser

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/service/redis_ser"

	"gorm.io/gorm"
)

// SyncCommentData redis同步评论数据到数据库
func SyncCommentData() {
	// 从redis中获取评论数据
	commentDiggInfo := redis_ser.NewCommentDigg().GetInfo()
	for key, value := range commentDiggInfo {
		var comment models.CommentModel
		err := global.DB.Take(&comment, key).Error
		if err != nil {
			global.Log.Error("查询评论数据失败: ", err)
			continue
		}
		// 更新评论数
		err = global.DB.Model(&comment).
			Update("digg_count", gorm.Expr("digg_count + ?", value)).Error
		if err != nil {
			global.Log.Error("更新评论数据失败: ", err)
			continue
		}
		// 更新评论成功，清空redis中的数据
		global.Log.Infof("更新评论数据成功: %v 评论数: %v", comment, value)
		redis_ser.NewCommentDigg().Clear()

	}
}
