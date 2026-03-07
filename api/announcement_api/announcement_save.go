package announcement_api

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
)

func buildAnnouncementModelFromRequest(cr announcementSaveRequest) (models.AnnouncementModel, error) {
	title := strings.TrimSpace(cr.Title)
	content := strings.TrimSpace(cr.Content)
	if title == "" {
		return models.AnnouncementModel{}, errors.New("公告标题不能为空")
	}
	if content == "" {
		return models.AnnouncementModel{}, errors.New("公告内容不能为空")
	}
	if len([]rune(title)) > 120 {
		return models.AnnouncementModel{}, errors.New("公告标题不能超过 120 个字符")
	}
	if len([]rune(content)) > 2000 {
		return models.AnnouncementModel{}, errors.New("公告内容不能超过 2000 个字符")
	}
	if err := validateAnnouncementBoard(cr.BoardID); err != nil {
		return models.AnnouncementModel{}, err
	}

	jumpLink, err := sanitizeAnnouncementJumpLink(cr.JumpLink)
	if err != nil {
		return models.AnnouncementModel{}, err
	}
	startsAt, err := parseAnnouncementTime(cr.StartsAt)
	if err != nil {
		return models.AnnouncementModel{}, err
	}
	endsAt, err := parseAnnouncementTime(cr.EndsAt)
	if err != nil {
		return models.AnnouncementModel{}, err
	}
	if startsAt != nil && endsAt != nil && endsAt.Before(*startsAt) {
		return models.AnnouncementModel{}, errors.New("结束时间不能早于开始时间")
	}

	isShow := true
	if cr.IsShow != nil {
		isShow = *cr.IsShow
	}

	return models.AnnouncementModel{
		Title:    title,
		Content:  content,
		Level:    normalizeAnnouncementLevel(cr.Level),
		JumpLink: jumpLink,
		BoardID:  cr.BoardID,
		Sort:     cr.Sort,
		IsShow:   isShow,
		StartsAt: startsAt,
		EndsAt:   endsAt,
	}, nil
}

// AnnouncementCreateView 创建公告。
func (AnnouncementApi) AnnouncementCreateView(c *gin.Context) {
	var cr announcementSaveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	model, err := buildAnnouncementModelFromRequest(cr)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	if err = global.DB.Create(&model).Error; err != nil {
		res.FailWithMessage("创建公告失败", c)
		return
	}
	res.OkWithMessage("创建公告成功", c)
}

// AnnouncementUpdateView 更新公告。
func (AnnouncementApi) AnnouncementUpdateView(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		res.FailWithMessage("公告ID不能为空", c)
		return
	}

	var model models.AnnouncementModel
	if err := global.DB.Take(&model, id).Error; err != nil {
		res.FailWithMessage("公告不存在", c)
		return
	}

	var cr announcementSaveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	updateModel, err := buildAnnouncementModelFromRequest(cr)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	if err = global.DB.Model(&model).Updates(map[string]any{
		"title":     updateModel.Title,
		"content":   updateModel.Content,
		"level":     updateModel.Level,
		"jump_link": updateModel.JumpLink,
		"board_id":  updateModel.BoardID,
		"sort":      updateModel.Sort,
		"is_show":   updateModel.IsShow,
		"starts_at": updateModel.StartsAt,
		"ends_at":   updateModel.EndsAt,
	}).Error; err != nil {
		res.FailWithMessage("更新公告失败", c)
		return
	}
	res.OkWithMessage("更新公告成功", c)
}
