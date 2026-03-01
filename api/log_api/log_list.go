package log_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"
	"gvb-server/service/common"

	"github.com/gin-gonic/gin"
)

// LogRequest 日志请求参数
type LogRequest struct {
	models.PageInfo
	Level log_stash.Level `form:"level" swag:"description:日志级别"`
}

// LogListView 获取日志列表
// @Summary 获取日志列表
// @Description 获取系统日志列表，支持按级别筛选
// @Tags 日志管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Param level query int false "日志级别"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]log_stash.LogStashModel}} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/logs [get]
func (LogApi) LogListView(c *gin.Context) {
	var cr LogRequest
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	// 创建基础模型实例
	baseModel := log_stash.LogStashModel{}

	// 构建查询选项
	option := common.Option{
		PageInfo: cr.PageInfo,
		Debug:    true,
	}

	// 如果level不为空，则添加到option中用于筛选
	// 如果level不为空，则添加到option中用于筛选
	if cr.Level != 0 {
		option.Level = cr.Level
	}

	list, count, _ := common.ComList(baseModel, option)
	res.OkWithList(list, count, c)
	return
}
