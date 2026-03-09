package board_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
)

var (
	_ = res.Response{}
	_ = models.RemoveRequest{}
)

// BoardListViewDoc 获取板块列表。
// @Tags 板块管理
// @Summary 获取板块列表
// @Description 默认只返回启用中的板块；scope=all 时返回全部板块。结果附带版主和副版主用户 ID 列表。
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 50，最大 200"
// @Param key query string false "按板块名称或描述搜索"
// @Param scope query string false "all 表示包含已停用板块"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]boardItem}} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Router /api/boards [get]
func BoardListViewDoc() {}

// BoardCreateViewDoc 创建板块。
// @Tags 板块管理
// @Summary 创建板块
// @Description 创建论坛板块，并可同时设置版主、副版主、置顶文章和展示状态。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body boardSaveRequest true "板块信息"
// @Success 200 {object} res.Response{data=models.BoardModel} "创建成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/boards [post]
func BoardCreateViewDoc() {}

// BoardUpdateViewDoc 更新板块。
// @Tags 板块管理
// @Summary 更新板块
// @Description 根据板块 ID 更新板块信息、版主管理关系和置顶文章设置。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body boardSaveRequest true "板块信息"
// @Success 200 {object} res.Response "更新成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/boards [put]
func BoardUpdateViewDoc() {}

// BoardRemoveViewDoc 删除板块。
// @Tags 板块管理
// @Summary 删除板块
// @Description 批量删除板块。建议先确保前台和文章没有继续依赖这些板块。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body models.RemoveRequest true "板块 ID 列表"
// @Success 200 {object} res.Response "删除成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/boards [delete]
func BoardRemoveViewDoc() {}
