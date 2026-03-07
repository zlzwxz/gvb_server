package board_api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

// BoardRemoveView 批量删除板块。
func (BoardApi) BoardRemoveView(c *gin.Context) {
	var cr models.RemoveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var list []models.BoardModel
	count := global.DB.Find(&list, cr.IDList).RowsAffected
	if count == 0 {
		res.FailWithMessage("板块不存在", c)
		return
	}
	if err := global.DB.Delete(&list).Error; err != nil {
		res.FailWithMessage("删除板块失败", c)
		return
	}
	res.OkWithMessage(fmt.Sprintf("共删除 %d 个板块", count), c)
}
