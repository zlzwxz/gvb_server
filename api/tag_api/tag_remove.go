package tag_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// TagRemoveView 批量删除标签
// @Summary 批量删除标签
// @Description 根据标签ID列表批量删除标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body models.RemoveRequest true "标签ID列表"
// @Success 200 {object} res.Response{msg=string} "删除成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "标签不存在"
// @Router /api/tags [delete]
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
