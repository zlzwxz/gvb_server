package data_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"time"

	"github.com/gin-gonic/gin"
)

// DateCount 日期计数结构
type DateCount struct {
	Date  string `json:"date" swag:"description:日期"`
	Count int    `json:"count" swag:"description:数量"`
}

// DateCountResponse 日期计数响应结构
type DateCountResponse struct {
	DateList  []string `json:"date_list" swag:"description:日期列表"`
	LoginData []int    `json:"login_data" swag:"description:登录数据"`
	SignData  []int    `json:"sign_data" swag:"description:注册数据"`
}

// SevenLoginView 获取近七日登录注册数据
// @Summary 获取近七日登录注册数据
// @Description 获取近七天内每天的登录和注册人数统计
// @Tags 数据统计
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=DateCountResponse} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/data/seven_login [get]
func (DataApi) SevenLoginView(c *gin.Context) {
	var loginDateCount, signDateCount []DateCount

	global.DB.Model(models.LoginDataModel{}).
		Where("date_sub(curdate(), interval 7 day) <= created_at").
		Select("date_format(created_at, '%Y-%m-%d') as date", "count(id) as count").
		Group("date").
		Scan(&loginDateCount)
	global.DB.Model(models.UserModel{}).
		Where("date_sub(curdate(), interval 7 day) <= created_at").
		Select("date_format(created_at, '%Y-%m-%d') as date", "count(id) as count").
		Group("date").
		Scan(&signDateCount)
	var loginDateCountMap = map[string]int{}
	var signDateCountMap = map[string]int{}
	var loginCountList, signCountList []int
	now := time.Now()
	for _, i2 := range loginDateCount {
		loginDateCountMap[i2.Date] = i2.Count
	}
	for _, i2 := range signDateCount {
		signDateCountMap[i2.Date] = i2.Count
	}
	var dateList []string
	for i := -6; i <= 0; i++ {
		day := now.AddDate(0, 0, i).Format("2006-01-02")
		loginCount := loginDateCountMap[day]
		signCount := signDateCountMap[day]
		dateList = append(dateList, day)
		loginCountList = append(loginCountList, loginCount)
		signCountList = append(signCountList, signCount)
	}

	res.OkWithData(DateCountResponse{
		DateList:  dateList,
		LoginData: loginCountList,
		SignData:  signCountList,
	}, c)

}
