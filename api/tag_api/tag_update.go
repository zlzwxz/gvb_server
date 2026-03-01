package tag_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
)

// TagUpdateView 更新标签
// @Summary 更新标签
// @Description 更新指定ID的标签名称
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path string true "标签ID"
// @Param data body TagRequest true "标签信息"
// @Success 200 {object} res.Response{msg=string} "更新成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "标签不存在"
// @Router /api/tags/{id} [put]
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

	res.OkWithMessage("修改标签成功", c)
}
