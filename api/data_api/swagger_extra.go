package data_api

import "gvb-server/models/res"

var _ = res.Response{}

// AdminDataSumViewDoc 获取后台运营统计。
// @Tags 数据统计
// @Summary 获取后台运营统计
// @Description 返回待审核文章数、待处理举报数、公告总数和启用板块数，适合作为后台首页卡片数据。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=AdminDataSumResponse} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/data_sum/admin [get]
func AdminDataSumViewDoc() {}
