package article_api

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/board_ser"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"
)

type articleReportCreateRequest struct {
	ArticleID string `json:"article_id" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
	Content   string `json:"content"`
}

type articleReportHandleRequest struct {
	ID         uint   `json:"id" binding:"required"`
	Status     int    `json:"status" binding:"required"`
	HandleNote string `json:"handle_note"`
}

type articleReportListQuery struct {
	models.PageInfo
	Status int `form:"status"`
}

func canHandleArticleReport(item models.ArticleReportModel, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	if isAdmin(claims) {
		return true
	}
	if item.BoardID == 0 {
		return false
	}
	board, err := board_ser.GetBoardByID(item.BoardID)
	if err != nil {
		return false
	}
	return board_ser.IsUserBoardManager(board, claims.UserID)
}

// ArticleReportCreateView 提交文章举报。
func (ArticleApi) ArticleReportCreateView(c *gin.Context) {
	var cr articleReportCreateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	article, err := es_ser.CommDetail(strings.TrimSpace(cr.ArticleID))
	if err != nil {
		res.FailWithMessage("文章不存在", c)
		return
	}
	if !canViewArticle(article, claims) {
		res.FailWithMessage("文章不可见，无法举报", c)
		return
	}
	if article.UserID == claims.UserID {
		res.FailWithMessage("不能举报自己的文章", c)
		return
	}

	reason := strings.TrimSpace(cr.Reason)
	content := strings.TrimSpace(cr.Content)
	if reason == "" {
		res.FailWithMessage("举报原因不能为空", c)
		return
	}
	if len([]rune(reason)) > 64 {
		reason = string([]rune(reason)[:64])
	}
	if len([]rune(content)) > 500 {
		res.FailWithMessage("举报补充说明不能超过 500 个字符", c)
		return
	}

	var count int64
	if err = global.DB.Model(&models.ArticleReportModel{}).
		Where("article_id = ? AND reporter_user_id = ? AND status = ?", article.ID, claims.UserID, ctype.ArticleReportPending).
		Count(&count).Error; err == nil && count > 0 {
		res.FailWithMessage("你已经举报过该文章，等待处理即可", c)
		return
	}

	reporterName := strings.TrimSpace(claims.NickName)
	if reporterName == "" {
		reporterName = strings.TrimSpace(claims.Username)
	}

	model := models.ArticleReportModel{
		ArticleID:        article.ID,
		ArticleTitle:     article.Title,
		BoardID:          article.BoardID,
		BoardName:        article.BoardName,
		ReporterUserID:   claims.UserID,
		ReporterNickName: reporterName,
		Reason:           reason,
		Content:          content,
		Status:           ctype.ArticleReportPending,
	}
	if err = global.DB.Create(&model).Error; err != nil {
		res.FailWithMessage("提交举报失败", c)
		return
	}
	res.OkWithMessage("举报已提交，等待版主或管理员处理", c)
}

// ArticleReportListView 获取文章举报列表。
func (ArticleApi) ArticleReportListView(c *gin.Context) {
	var cr articleReportListQuery
	_ = c.ShouldBindQuery(&cr)

	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	query := global.DB.Model(&models.ArticleReportModel{})
	if !isAdmin(claims) {
		boardIDs, err := board_ser.ListModeratedBoardIDs(claims.UserID)
		if err != nil {
			res.FailWithMessage("获取可处理板块失败", c)
			return
		}
		if len(boardIDs) == 0 {
			res.OkWithList([]models.ArticleReportModel{}, 0, c)
			return
		}
		query = query.Where("board_id IN ?", boardIDs)
	}

	if cr.Status != 0 {
		query = query.Where("status = ?", cr.Status)
	}
	if key := strings.TrimSpace(cr.Key); key != "" {
		like := "%" + key + "%"
		query = query.Where("article_title LIKE ? OR reporter_nick_name LIKE ? OR reason LIKE ? OR content LIKE ?", like, like, like, like)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取举报数量失败", c)
		return
	}

	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 10
	}
	if cr.Limit > 100 {
		cr.Limit = 100
	}

	var list []models.ArticleReportModel
	if err := query.Order("status asc").Order("created_at desc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&list).Error; err != nil {
		res.FailWithMessage("获取举报列表失败", c)
		return
	}
	res.OkWithList(list, count, c)
}

// ArticleReportHandleView 处理文章举报。
// status=2 表示进入复审；status=3 表示忽略。
func (ArticleApi) ArticleReportHandleView(c *gin.Context) {
	var cr articleReportHandleRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	if cr.Status != int(ctype.ArticleReportReReview) && cr.Status != int(ctype.ArticleReportDismissed) {
		res.FailWithMessage("处理状态非法", c)
		return
	}

	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	var model models.ArticleReportModel
	if err := global.DB.Take(&model, cr.ID).Error; err != nil {
		res.FailWithMessage("举报记录不存在", c)
		return
	}
	if !canHandleArticleReport(model, claims) {
		res.FailWithMessage("无权处理该举报", c)
		return
	}

	handleNote := strings.TrimSpace(cr.HandleNote)
	if len([]rune(handleNote)) > 255 {
		handleNote = string([]rune(handleNote)[:255])
	}
	handlerName := strings.TrimSpace(claims.NickName)
	if handlerName == "" {
		handlerName = strings.TrimSpace(claims.Username)
	}

	now := time.Now()
	if err := global.DB.Model(&model).Updates(map[string]any{
		"status":                 ctype.ArticleReportStatus(cr.Status),
		"handle_note":            handleNote,
		"handler_user_id":        claims.UserID,
		"handler_user_nick_name": handlerName,
		"handled_at":             &now,
	}).Error; err != nil {
		res.FailWithMessage("处理举报失败", c)
		return
	}

	if cr.Status == int(ctype.ArticleReportReReview) {
		if _, err := es_ser.CommDetail(model.ArticleID); err == nil {
			_ = es_ser.ArticleUpdate(model.ArticleID, map[string]any{
				"review_status":      ctype.ArticleReviewPending,
				"review_reason":      "文章被举报后进入复审",
				"reviewed_at":        "",
				"reviewer_id":        0,
				"reviewer_nick_name": "",
			})
			es_ser.DeleteFullTextByArticleID(model.ArticleID)
		}
		res.OkWithMessage("已转入复审队列", c)
		return
	}

	res.OkWithMessage("已忽略该举报", c)
}
