package announcement_api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
)

const announcementTimeLayout = "2006-01-02 15:04:05"

type announcementListQuery struct {
	models.PageInfo
	BoardID uint   `form:"board_id"`
	IsShow  *bool  `form:"is_show"`
	Scope   string `form:"scope"`
}

type announcementSaveRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Level    string `json:"level"`
	JumpLink string `json:"jump_link"`
	BoardID  uint   `json:"board_id"`
	Sort     int    `json:"sort"`
	IsShow   *bool  `json:"is_show"`
	StartsAt string `json:"starts_at"`
	EndsAt   string `json:"ends_at"`
}

type announcementItem struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Level     string `json:"level"`
	JumpLink  string `json:"jump_link"`
	BoardID   uint   `json:"board_id"`
	BoardName string `json:"board_name"`
	BoardSlug string `json:"board_slug"`
	Sort      int    `json:"sort"`
	IsShow    bool   `json:"is_show"`
	StartsAt  string `json:"starts_at"`
	EndsAt    string `json:"ends_at"`
	CreatedAt string `json:"created_at"`
}

func normalizeAnnouncementLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "success":
		return "success"
	case "warning":
		return "warning"
	case "danger":
		return "danger"
	default:
		return "info"
	}
}

func sanitizeAnnouncementJumpLink(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	lowerValue := strings.ToLower(value)
	if strings.HasPrefix(lowerValue, "javascript:") {
		return "", errors.New("公告跳转链接非法")
	}
	if strings.HasPrefix(lowerValue, "http://") || strings.HasPrefix(lowerValue, "https://") || strings.HasPrefix(value, "/") {
		return value, nil
	}
	return "", errors.New("公告跳转链接仅支持 http(s) 或站内路径")
}

func parseAnnouncementTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	layouts := []string{
		announcementTimeLayout,
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("时间格式错误，请使用 %s", announcementTimeLayout)
}

func formatAnnouncementTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Local().Format(announcementTimeLayout)
}

func validateAnnouncementBoard(boardID uint) error {
	if boardID == 0 {
		return nil
	}
	var board models.BoardModel
	if err := global.DB.Take(&board, boardID).Error; err != nil {
		return errors.New("所选板块不存在")
	}
	return nil
}

func buildAnnouncementItems(list []models.AnnouncementModel) ([]announcementItem, error) {
	boardIDSet := map[uint]struct{}{}
	for _, item := range list {
		if item.BoardID > 0 {
			boardIDSet[item.BoardID] = struct{}{}
		}
	}

	boardIDs := make([]uint, 0, len(boardIDSet))
	for boardID := range boardIDSet {
		boardIDs = append(boardIDs, boardID)
	}

	boardMap := map[uint]models.BoardModel{}
	if len(boardIDs) > 0 {
		var boardList []models.BoardModel
		if err := global.DB.Where("id IN ?", boardIDs).Find(&boardList).Error; err != nil {
			return nil, err
		}
		for _, board := range boardList {
			boardMap[board.ID] = board
		}
	}

	result := make([]announcementItem, 0, len(list))
	for _, item := range list {
		board := boardMap[item.BoardID]
		result = append(result, announcementItem{
			ID:        item.ID,
			Title:     item.Title,
			Content:   item.Content,
			Level:     normalizeAnnouncementLevel(item.Level),
			JumpLink:  item.JumpLink,
			BoardID:   item.BoardID,
			BoardName: board.Name,
			BoardSlug: board.Slug,
			Sort:      item.Sort,
			IsShow:    item.IsShow,
			StartsAt:  formatAnnouncementTime(item.StartsAt),
			EndsAt:    formatAnnouncementTime(item.EndsAt),
			CreatedAt: item.CreatedAt.Format(announcementTimeLayout),
		})
	}
	return result, nil
}
