package announcement_api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

// AnnouncementListView 公共公告列表。
// 首页默认只返回全站公告；传 board_id 时返回全站公告和指定板块公告。
func (AnnouncementApi) AnnouncementListView(c *gin.Context) {
	var cr announcementListQuery
	_ = c.ShouldBindQuery(&cr)

	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 6
	}
	if cr.Limit > 20 {
		cr.Limit = 20
	}

	now := time.Now()
	query := global.DB.Model(&models.AnnouncementModel{}).
		Where("is_show = ?", true).
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at >= ?)", now)

	if key := strings.TrimSpace(cr.Key); key != "" {
		like := "%" + key + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", like, like)
	}

	if cr.BoardID > 0 {
		query = query.Where("(board_id = 0 OR board_id = ?)", cr.BoardID)
		query = query.Order(fmt.Sprintf("CASE WHEN board_id = %d THEN 0 ELSE 1 END", cr.BoardID))
	} else {
		query = query.Where("board_id = 0")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取公告数量失败", c)
		return
	}

	var list []models.AnnouncementModel
	if err := query.Order("sort asc").Order("created_at desc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&list).Error; err != nil {
		res.FailWithMessage("获取公告列表失败", c)
		return
	}

	items, err := buildAnnouncementItems(list)
	if err != nil {
		res.FailWithMessage("组装公告列表失败", c)
		return
	}
	res.OkWithList(items, count, c)
}

// AnnouncementManageListView 后台公告列表。
func (AnnouncementApi) AnnouncementManageListView(c *gin.Context) {
	var cr announcementListQuery
	_ = c.ShouldBindQuery(&cr)

	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 10
	}
	if cr.Limit > 100 {
		cr.Limit = 100
	}

	query := global.DB.Model(&models.AnnouncementModel{})
	if key := strings.TrimSpace(cr.Key); key != "" {
		like := "%" + key + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", like, like)
	}
	if cr.BoardID > 0 {
		query = query.Where("board_id = ?", cr.BoardID)
	} else if strings.EqualFold(strings.TrimSpace(cr.Scope), "global") {
		query = query.Where("board_id = 0")
	}
	if cr.IsShow != nil {
		query = query.Where("is_show = ?", *cr.IsShow)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取公告数量失败", c)
		return
	}

	var list []models.AnnouncementModel
	if err := query.Order("sort asc").Order("created_at desc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&list).Error; err != nil {
		res.FailWithMessage("获取公告列表失败", c)
		return
	}

	items, err := buildAnnouncementItems(list)
	if err != nil {
		res.FailWithMessage("组装公告列表失败", c)
		return
	}
	res.OkWithList(items, count, c)
}
