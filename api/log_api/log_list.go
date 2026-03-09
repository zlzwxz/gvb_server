package log_api

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"
)

// LogRequest 日志请求参数（支持多条件筛选）。
type LogRequest struct {
	models.PageInfo
	Level      log_stash.Level `form:"level"`
	Method     string          `form:"method"`
	Path       string          `form:"path"`
	IP         string          `form:"ip"`
	UserID     uint            `form:"user_id"`
	StatusCode int             `form:"status_code"`
	RespCode   *int            `form:"resp_code"`
	DateFrom   string          `form:"date_from"`
	DateTo     string          `form:"date_to"`
}

// LogListView 获取日志列表。
// @Summary 获取日志列表
// @Description 按级别、请求方法、路径、用户、响应码和时间范围筛选后台操作日志
// @Tags 日志管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 20，最大 200"
// @Param key query string false "按内容、请求体、响应体搜索"
// @Param level query int false "日志级别"
// @Param method query string false "请求方法"
// @Param path query string false "请求路径"
// @Param ip query string false "IP"
// @Param user_id query int false "用户 ID"
// @Param status_code query int false "HTTP 状态码"
// @Param resp_code query int false "业务响应码"
// @Param date_from query string false "开始时间，支持 2006-01-02 或 2006-01-02 15:04:05"
// @Param date_to query string false "结束时间，支持 2006-01-02 或 2006-01-02 15:04:05"
// @Param sort query string false "排序，如 created_at desc"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]log_stash.LogStashModel}} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/logs [get]
func (LogApi) LogListView(c *gin.Context) {
	var cr LogRequest
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 20
	}
	if cr.Limit > 200 {
		cr.Limit = 200
	}

	query := global.DB.Model(&log_stash.LogStashModel{})
	if cr.Level != 0 {
		query = query.Where("level = ?", cr.Level)
	}
	if strings.TrimSpace(cr.Method) != "" {
		query = query.Where("method = ?", strings.ToUpper(strings.TrimSpace(cr.Method)))
	}
	if strings.TrimSpace(cr.Path) != "" {
		query = query.Where("path LIKE ?", "%"+strings.TrimSpace(cr.Path)+"%")
	}
	if strings.TrimSpace(cr.IP) != "" {
		query = query.Where("ip LIKE ?", "%"+strings.TrimSpace(cr.IP)+"%")
	}
	if cr.UserID > 0 {
		query = query.Where("user_id = ?", cr.UserID)
	}
	if cr.StatusCode > 0 {
		query = query.Where("status_code = ?", cr.StatusCode)
	}
	if cr.RespCode != nil {
		query = query.Where("resp_code = ?", *cr.RespCode)
	}
	if strings.TrimSpace(cr.Key) != "" {
		key := "%" + strings.TrimSpace(cr.Key) + "%"
		query = query.Where("content LIKE ? OR request_body LIKE ? OR response_body LIKE ?", key, key, key)
	}
	if from, ok := parseLogTime(strings.TrimSpace(cr.DateFrom), false); ok {
		query = query.Where("created_at >= ?", from)
	}
	if to, ok := parseLogTime(strings.TrimSpace(cr.DateTo), true); ok {
		query = query.Where("created_at <= ?", to)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取日志数量失败", c)
		return
	}

	sortClause := sanitizeLogSort(cr.Sort)
	var list []log_stash.LogStashModel
	err := query.
		Order(sortClause).
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&list).Error
	if err != nil {
		res.FailWithMessage("获取日志列表失败", c)
		return
	}
	res.OkWithList(list, count, c)
}

func parseLogTime(raw string, endOfDay bool) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			if endOfDay && layout == "2006-01-02" {
				t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			}
			return t, true
		}
	}
	return time.Time{}, false
}

func sanitizeLogSort(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "created_at desc"
	}
	parts := strings.Fields(raw)
	field := parts[0]
	direction := "desc"
	if len(parts) > 1 && (parts[1] == "asc" || parts[1] == "desc") {
		direction = parts[1]
	}
	allow := map[string]struct{}{
		"created_at":  {},
		"status_code": {},
		"resp_code":   {},
		"level":       {},
		"user_id":     {},
	}
	if _, ok := allow[field]; !ok {
		return "created_at desc"
	}
	return field + " " + direction
}
