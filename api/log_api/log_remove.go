package log_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"

	"github.com/gin-gonic/gin"
)

// LogRemoveListView 批量删除日志
// @Summary 批量删除日志
// @Description 根据日志ID列表批量删除日志
// @Tags 日志管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body models.RemoveRequest true "日志ID列表"
// @Success 200 {object} res.Response{msg=string} "删除成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "日志不存在"
// @Router /api/logs [delete]
func (LogApi) LogRemoveListView(c *gin.Context) {
	var cr models.RemoveRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var list []log_stash.LogStashModel
	count := global.DB.Find(&list, cr.IDList).RowsAffected
	if count == 0 {
		res.FailWithMessage("日志不存在", c)
		return
	}
	global.DB.Delete(&list)
	res.OkWithMessage(fmt.Sprintf("共删除 %d 个日志", count), c)

}
