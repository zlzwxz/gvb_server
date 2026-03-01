package tag_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

func (TagApi) TagRemoveView(c *gin.Context) {
	var cr models.RemoveRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var tagtList []models.TagModel
	count := global.DB.Find(&tagtList, cr.IDList).RowsAffected
	if count == 0 {
		res.FailWithMessage("标签不存在", c)
		return
	}
	global.DB.Delete(&tagtList)
	res.OkWithMessage(fmt.Sprintf("共删除 %d 个标签告", count), c)

}
