package tag_api

import (
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

func (TagApi) TagUpdateView(c *gin.Context) {

	// 从上下文参数中获取ID值
	// 该方法用于提取URL路径中的id参数
	id := c.Param("id")
	var cr TagRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	var tag models.TagModel
	err = global.DB.Take(&tag, id).Error
	if err != nil {
		res.FailWithMessage("标签不存在", c)
		return
	}
	// 结构体转map的第三方包
	maps := structs.Map(&cr)
	err = global.DB.Model(&tag).Updates(maps).Error

	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("修改标签失败", c)
		return
	}

	res.OkWithMessage("修改广告成功", c)
}
