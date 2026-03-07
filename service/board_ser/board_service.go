package board_ser

import (
	"fmt"
	"strconv"
	"strings"

	"gvb-server/global"
	"gvb-server/models"
)

var defaultBoards = []models.BoardModel{
	{
		Name:        "泳池板块",
		Slug:        "pool",
		Description: "泳池话题、健身训练、户外活动交流区",
		Sort:        1,
		IsEnabled:   true,
	},
	{
		Name:        "AI板块",
		Slug:        "ai",
		Description: "AI 工具、模型实践、提示词与落地经验分享",
		Sort:        2,
		IsEnabled:   true,
	},
	{
		Name:        "GitHub板块",
		Slug:        "github",
		Description: "开源项目、仓库推荐、协作流程与工程实践",
		Sort:        3,
		IsEnabled:   true,
	},
}

func EnsureDefaultBoards() error {
	if global.DB == nil {
		return nil
	}
	var count int64
	if err := global.DB.Model(&models.BoardModel{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return global.DB.Create(&defaultBoards).Error
}

func GetBoardByID(boardID uint) (models.BoardModel, error) {
	var board models.BoardModel
	if boardID == 0 {
		return board, fmt.Errorf("板块ID不能为空")
	}
	if err := global.DB.Take(&board, boardID).Error; err != nil {
		return board, err
	}
	return board, nil
}

func GetEnabledBoardByID(boardID uint) (models.BoardModel, error) {
	var board models.BoardModel
	if boardID == 0 {
		return board, fmt.Errorf("板块ID不能为空")
	}
	if err := global.DB.Where("id = ? and is_enabled = ?", boardID, true).Take(&board).Error; err != nil {
		return board, err
	}
	return board, nil
}

func IsUserBoardManager(board models.BoardModel, userID uint) bool {
	if userID == 0 {
		return false
	}
	if containsID(board.ModeratorIDs, userID) {
		return true
	}
	if containsID(board.DeputyModeratorIDs, userID) {
		return true
	}
	return false
}

func ParseUintIDs(values []string) []uint {
	list := make([]uint, 0, len(values))
	seen := map[uint]struct{}{}
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		id64, err := strconv.ParseUint(text, 10, 64)
		if err != nil || id64 == 0 {
			continue
		}
		id := uint(id64)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		list = append(list, id)
	}
	return list
}

func ToStringIDs(values []uint) []string {
	list := make([]string, 0, len(values))
	seen := map[uint]struct{}{}
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		list = append(list, strconv.FormatUint(uint64(value), 10))
	}
	return list
}

func containsID(values []string, target uint) bool {
	targetText := strconv.FormatUint(uint64(target), 10)
	for _, value := range values {
		if strings.TrimSpace(value) == targetText {
			return true
		}
	}
	return false
}

// ListModeratedBoardIDs 返回用户拥有审核权限的板块 ID 列表。
// 主版主和副版主都视为可审核该板块内容。
func ListModeratedBoardIDs(userID uint) ([]uint, error) {
	if userID == 0 {
		return []uint{}, nil
	}
	var boards []models.BoardModel
	if err := global.DB.Select("id", "moderator_ids").Find(&boards).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(boards))
	for _, board := range boards {
		if IsUserBoardManager(board, userID) {
			ids = append(ids, board.ID)
		}
	}
	return ids, nil
}
