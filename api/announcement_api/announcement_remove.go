package announcement_api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

// AnnouncementRemoveView 批量删除公告。
func (AnnouncementApi) AnnouncementRemoveView(c *gin.Context) {
	var cr models.RemoveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var list []models.AnnouncementModel
	count := global.DB.Find(&list, cr.IDList).RowsAffected
	if count == 0 {
		res.FailWithMessage("公告不存在", c)
		return
	}
	if err := global.DB.Delete(&list).Error; err != nil {
		res.FailWithMessage("删除公告失败", c)
		return
	}
	res.OkWithMessage(fmt.Sprintf("共删除 %d 条公告", count), c)
}
