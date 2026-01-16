package advert_api

import (
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

// AdvertUpdateView 修改广告
// @Tags 广告管理
// @Summary 更新广告
// @Accept json
// @Produce json
// @Param id path int true "广告ID"
// @Param data body AdvertRequest true "广告更新数据"
// @Router /api/adverts/{id} [put]
// @Success 200 {object} res.Response{message=string}
func (AdvertApi) AdvertUpdateView(c *gin.Context) {

	// 从上下文参数中获取ID值
	// 该方法用于提取URL路径中的id参数
	id := c.Param("id")
	var cr AdvertRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	var advert models.AdvertModel
	err = global.DB.Take(&advert, id).Error
	if err != nil {
		res.FailWithMessage("广告不存在", c)
		return
	}
	// 结构体转map的第三方包
	maps := structs.Map(&cr)
	err = global.DB.Model(&advert).Updates(maps).Error

	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("修改广告失败", c)
		return
	}

	res.OkWithMessage("修改广告成功", c)
}
