package board_api

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/board_ser"
)

type boardSaveRequest struct {
	ID                     uint     `json:"id"`
	Name                   string   `json:"name" binding:"required"`
	Slug                   string   `json:"slug"`
	Description            string   `json:"description"`
	Notice                 string   `json:"notice"`
	Rules                  string   `json:"rules"`
	Sort                   int      `json:"sort"`
	IsEnabled              *bool    `json:"is_enabled"`
	PinnedArticleIDs       []string `json:"pinned_article_ids"`
	ModeratorUserIDs       []uint   `json:"moderator_user_ids"`
	DeputyModeratorUserIDs []uint   `json:"deputy_moderator_user_ids"`
}

var boardSlugRegex = regexp.MustCompile(`[^a-z0-9_-]+`)

func sanitizeBoardSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = boardSlugRegex.ReplaceAllString(value, "")
	value = strings.Trim(value, "-_")
	if value == "" {
		return "board"
	}
	return value
}

func sanitizeBoardArticleIDs(values []string) []string {
	list := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		list = append(list, text)
		if len(list) >= 20 {
			break
		}
	}
	return list
}

// BoardCreateView 新建板块。
func (BoardApi) BoardCreateView(c *gin.Context) {
	var cr boardSaveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	cr.Name = strings.TrimSpace(cr.Name)
	if cr.Name == "" {
		res.FailWithMessage("板块名称不能为空", c)
		return
	}
	slug := sanitizeBoardSlug(cr.Slug)
	if slug == "board" {
		slug = sanitizeBoardSlug(cr.Name)
	}
	enabled := true
	if cr.IsEnabled != nil {
		enabled = *cr.IsEnabled
	}
	model := models.BoardModel{
		Name:               cr.Name,
		Slug:               slug,
		Description:        strings.TrimSpace(cr.Description),
		Notice:             strings.TrimSpace(cr.Notice),
		Rules:              strings.TrimSpace(cr.Rules),
		PinnedArticleIDs:   ctype.Array(sanitizeBoardArticleIDs(cr.PinnedArticleIDs)),
		Sort:               cr.Sort,
		IsEnabled:          enabled,
		ModeratorIDs:       ctype.Array(board_ser.ToStringIDs(cr.ModeratorUserIDs)),
		DeputyModeratorIDs: ctype.Array(board_ser.ToStringIDs(cr.DeputyModeratorUserIDs)),
	}
	if err := global.DB.Create(&model).Error; err != nil {
		res.FailWithMessage("创建板块失败: "+err.Error(), c)
		return
	}
	res.OkWithData(model, c)
}

// BoardUpdateView 更新板块。
func (BoardApi) BoardUpdateView(c *gin.Context) {
	var cr boardSaveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	if cr.ID == 0 {
		res.FailWithMessage("板块ID不能为空", c)
		return
	}
	var model models.BoardModel
	if err := global.DB.Take(&model, cr.ID).Error; err != nil {
		res.FailWithMessage("板块不存在", c)
		return
	}

	updateMap := map[string]any{}
	if strings.TrimSpace(cr.Name) != "" {
		updateMap["name"] = strings.TrimSpace(cr.Name)
	}
	if strings.TrimSpace(cr.Slug) != "" {
		updateMap["slug"] = sanitizeBoardSlug(cr.Slug)
	}
	updateMap["description"] = strings.TrimSpace(cr.Description)
	updateMap["notice"] = strings.TrimSpace(cr.Notice)
	updateMap["rules"] = strings.TrimSpace(cr.Rules)
	updateMap["pinned_article_ids"] = ctype.Array(sanitizeBoardArticleIDs(cr.PinnedArticleIDs))
	updateMap["sort"] = cr.Sort
	if cr.IsEnabled != nil {
		updateMap["is_enabled"] = *cr.IsEnabled
	}
	if cr.ModeratorUserIDs != nil {
		updateMap["moderator_ids"] = ctype.Array(board_ser.ToStringIDs(cr.ModeratorUserIDs))
	}
	if cr.DeputyModeratorUserIDs != nil {
		updateMap["deputy_moderator_ids"] = ctype.Array(board_ser.ToStringIDs(cr.DeputyModeratorUserIDs))
	}
	if err := global.DB.Model(&model).Updates(updateMap).Error; err != nil {
		res.FailWithMessage("更新板块失败: "+err.Error(), c)
		return
	}
	res.OkWithMessage("更新板块成功", c)
}
