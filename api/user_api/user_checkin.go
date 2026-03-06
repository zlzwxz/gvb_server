package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type UserCheckInStatusResponse struct {
	CheckedToday bool   `json:"checked_today"`
	CheckInDate  string `json:"check_in_date"`
	Points       int    `json:"points"`
	Experience   int    `json:"experience"`
	Level        int    `json:"level"`
	Streak       int    `json:"streak"`
}

type UserCheckInResponse struct {
	UserCheckInStatusResponse
	RewardPoints     int `json:"reward_points"`
	RewardExperience int `json:"reward_experience"`
}

// UserCheckInStatusView 获取用户签到状态
func (UserApi) UserCheckInStatusView(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}

	today := time.Now().Format("2006-01-02")
	checkedToday := user.LastCheckInAt != nil && user.LastCheckInAt.Format("2006-01-02") == today

	res.OkWithData(UserCheckInStatusResponse{
		CheckedToday: checkedToday,
		CheckInDate:  today,
		Points:       user.Points,
		Experience:   user.Experience,
		Level:        user.Level,
		Streak:       user.CheckInStreak,
	}, c)
}

// UserCheckInView 用户签到
func (UserApi) UserCheckInView(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	if claims.Role == int(ctype.PermissionVisitor) {
		res.FailWithMessage("游客用户不可签到", c)
		return
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	tx := global.DB.Begin()
	var user models.UserModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&user, claims.UserID).Error; err != nil {
		tx.Rollback()
		res.FailWithMessage("用户不存在", c)
		return
	}

	var checkIn models.UserCheckInModel
	if err := tx.Take(&checkIn, "user_id = ? and check_in_date = ?", user.ID, today).Error; err == nil {
		tx.Rollback()
		res.FailWithMessage("今日已签到", c)
		return
	}

	streak := 1
	if user.LastCheckInAt != nil {
		lastDate := user.LastCheckInAt.Format("2006-01-02")
		if lastDate == today {
			tx.Rollback()
			res.FailWithMessage("今日已签到", c)
			return
		}
		if lastDate == yesterday {
			streak = user.CheckInStreak + 1
		}
	}

	bonusFactor := streak - 1
	if bonusFactor < 0 {
		bonusFactor = 0
	}
	if bonusFactor > 5 {
		bonusFactor = 5
	}

	rewardPoints := 5 + bonusFactor
	rewardExperience := 10 + bonusFactor*2
	newExperience := user.Experience + rewardExperience
	newLevel := calcUserLevel(newExperience)

	if err := tx.Create(&models.UserCheckInModel{
		UserID:      user.ID,
		CheckInDate: today,
		Points:      rewardPoints,
		Experience:  rewardExperience,
		Streak:      streak,
	}).Error; err != nil {
		tx.Rollback()
		res.FailWithMessage("签到失败", c)
		return
	}

	if err := tx.Model(&models.UserModel{}).Where("id = ?", user.ID).Updates(map[string]any{
		"points":           user.Points + rewardPoints,
		"experience":       newExperience,
		"level":            newLevel,
		"check_in_streak":  streak,
		"last_check_in_at": now,
	}).Error; err != nil {
		tx.Rollback()
		res.FailWithMessage("签到失败", c)
		return
	}

	if err := tx.Commit().Error; err != nil {
		res.FailWithMessage("签到失败", c)
		return
	}

	res.OkWithData(UserCheckInResponse{
		UserCheckInStatusResponse: UserCheckInStatusResponse{
			CheckedToday: true,
			CheckInDate:  today,
			Points:       user.Points + rewardPoints,
			Experience:   newExperience,
			Level:        newLevel,
			Streak:       streak,
		},
		RewardPoints:     rewardPoints,
		RewardExperience: rewardExperience,
	}, c)
}

func calcUserLevel(experience int) int {
	if experience <= 0 {
		return 1
	}
	level := experience/100 + 1
	if level > 100 {
		level = 100
	}
	return level
}
