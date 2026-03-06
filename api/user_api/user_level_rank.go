package user_api

import (
	"fmt"
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// UserLevelRankItem 等级排行榜条目。
type UserLevelRankItem struct {
	ID            uint   `json:"id"`
	NickName      string `json:"nick_name"`
	UserName      string `json:"user_name"`
	Avatar        string `json:"avatar"`
	Level         int    `json:"level"`
	Experience    int    `json:"experience"`
	Points        int    `json:"points"`
	CheckInStreak int    `json:"check_in_streak"`
}

// UserLevelRankQuery 等级排行榜查询参数。
type UserLevelRankQuery struct {
	Limit int `form:"limit" swag:"description:榜单数量，默认10，最大50"`
}

// UserLevelRankView 获取用户等级排行榜。
// @Summary 获取用户等级排行榜
// @Description 按等级、经验、积分排序返回用户等级榜（游客账号默认不参与）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param limit query int false "榜单数量，默认10，最大50"
// @Success 200 {object} res.Response{data=[]UserLevelRankItem} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/user_level_rank [get]
func (UserApi) UserLevelRankView(c *gin.Context) {
	var cr UserLevelRankQuery
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	limit := cr.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var users []models.UserModel
	err := global.DB.
		Model(&models.UserModel{}).
		Select("id", "nick_name", "user_name", "avatar", "level", "experience", "points", "check_in_streak", "role").
		Where("role <> ?", ctype.PermissionVisitor).
		Order("level DESC").
		Order("experience DESC").
		Order("points DESC").
		Order("id ASC").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("获取用户等级排行榜失败", c)
		return
	}

	list := make([]UserLevelRankItem, 0, len(users))
	for _, user := range users {
		displayName := strings.TrimSpace(user.NickName)
		if displayName == "" {
			displayName = strings.TrimSpace(user.UserName)
		}
		if displayName == "" {
			displayName = fmt.Sprintf("用户%d", user.ID)
		}
		list = append(list, UserLevelRankItem{
			ID:            user.ID,
			NickName:      displayName,
			UserName:      user.UserName,
			Avatar:        user.Avatar,
			Level:         user.Level,
			Experience:    user.Experience,
			Points:        user.Points,
			CheckInStreak: user.CheckInStreak,
		})
	}

	res.OkWithData(list, c)
}
