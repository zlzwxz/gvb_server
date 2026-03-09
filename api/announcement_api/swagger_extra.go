package announcement_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
)

var (
	_ = res.Response{}
	_ = models.RemoveRequest{}
)

// AnnouncementListViewDoc 获取公告列表。
// @Tags 公告管理
// @Summary 获取公告列表
// @Description 前台公告列表。默认只返回全站公告；传 board_id 时会同时返回该板块公告和全站公告。
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 6，最大 20"
// @Param key query string false "按标题或内容模糊搜索"
// @Param board_id query int false "板块 ID"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]announcementItem}} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Router /api/announcements [get]
func AnnouncementListViewDoc() {}

// AnnouncementManageListViewDoc 获取后台公告列表。
// @Tags 公告管理
// @Summary 获取后台公告列表
// @Description 支持按关键词、板块、是否展示筛选。scope=global 时只返回全站公告。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 10，最大 100"
// @Param key query string false "按标题或内容搜索"
// @Param board_id query int false "板块 ID"
// @Param is_show query bool false "是否展示"
// @Param scope query string false "global 表示只看全站公告"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]announcementItem}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/announcements/manage [get]
func AnnouncementManageListViewDoc() {}

// AnnouncementCreateViewDoc 创建公告。
// @Tags 公告管理
// @Summary 创建公告
// @Description 支持创建全站公告或板块公告。开始结束时间留空表示长期有效。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body announcementSaveRequest true "公告信息"
// @Success 200 {object} res.Response "创建成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/announcements [post]
func AnnouncementCreateViewDoc() {}

// AnnouncementUpdateViewDoc 更新公告。
// @Tags 公告管理
// @Summary 更新公告
// @Description 根据公告 ID 更新公告内容、展示状态、排序和有效期。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "公告 ID"
// @Param data body announcementSaveRequest true "公告信息"
// @Success 200 {object} res.Response "更新成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/announcements/{id} [put]
func AnnouncementUpdateViewDoc() {}

// AnnouncementRemoveViewDoc 删除公告。
// @Tags 公告管理
// @Summary 删除公告
// @Description 支持批量删除公告。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body models.RemoveRequest true "公告 ID 列表"
// @Success 200 {object} res.Response "删除成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/announcements [delete]
func AnnouncementRemoveViewDoc() {}
