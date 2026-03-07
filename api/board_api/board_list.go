package board_api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/board_ser"
)

type boardQuery struct {
	models.PageInfo
}

type boardItem struct {
	models.BoardModel
	ModeratorUserIDs       []uint `json:"moderator_user_ids"`
	DeputyModeratorUserIDs []uint `json:"deputy_moderator_user_ids"`
}

// BoardListView 板块列表。
func (BoardApi) BoardListView(c *gin.Context) {
	if err := board_ser.EnsureDefaultBoards(); err != nil {
		res.FailWithMessage("初始化默认板块失败: "+err.Error(), c)
		return
	}

	var cr boardQuery
	_ = c.ShouldBindQuery(&cr)

	query := global.DB.Model(&models.BoardModel{})
	if strings.ToLower(strings.TrimSpace(c.Query("scope"))) != "all" {
		query = query.Where("is_enabled = ?", true)
	}
	if strings.TrimSpace(cr.Key) != "" {
		like := "%" + strings.TrimSpace(cr.Key) + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", like, like)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取板块数量失败", c)
		return
	}

	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 50
	}
	if cr.Limit > 200 {
		cr.Limit = 200
	}
	var list []models.BoardModel
	err := query.Order("sort asc").Order("id asc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&list).Error
	if err != nil {
		res.FailWithMessage("获取板块列表失败", c)
		return
	}

	result := make([]boardItem, 0, len(list))
	for _, item := range list {
		result = append(result, boardItem{
			BoardModel:             item,
			ModeratorUserIDs:       board_ser.ParseUintIDs(item.ModeratorIDs),
			DeputyModeratorUserIDs: board_ser.ParseUintIDs(item.DeputyModeratorIDs),
		})
	}
	res.OkWithList(result, count, c)
}
