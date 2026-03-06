package article_api

import (
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleReviewRequest struct {
	ID           string `json:"id" binding:"required" msg:"文章ID不能为空"`
	ReviewStatus int    `json:"review_status" binding:"required" msg:"审核状态不能为空"`
	ReviewReason string `json:"review_reason"`
}

// ArticleReviewView 管理员审核文章
func (ArticleApi) ArticleReviewView(c *gin.Context) {
	var cr ArticleReviewRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	if cr.ReviewStatus != int(ctype.ArticleReviewApproved) && cr.ReviewStatus != int(ctype.ArticleReviewRejected) {
		res.FailWithMessage("审核状态无效", c)
		return
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	article, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage("文章不存在", c)
		return
	}

	reason := strings.TrimSpace(cr.ReviewReason)
	if cr.ReviewStatus == int(ctype.ArticleReviewApproved) {
		reason = ""
	}

	err = es_ser.ArticleUpdate(cr.ID, map[string]any{
		"review_status":      cr.ReviewStatus,
		"review_reason":      reason,
		"reviewed_at":        time.Now().Format("2006-01-02 15:04:05"),
		"reviewer_id":        claims.UserID,
		"reviewer_nick_name": claims.NickName,
	})
	if err != nil {
		res.FailWithMessage("审核失败", c)
		return
	}

	if cr.ReviewStatus == int(ctype.ArticleReviewApproved) {
		es_ser.DeleteFullTextByArticleID(cr.ID)
		es_ser.AsyncArticleByFullText(es_ser.SearchData{
			Key:   cr.ID,
			Body:  article.Content,
			Slug:  es_ser.GetSlug(article.Title),
			Title: article.Title,
		})
		res.OkWithMessage("审核通过", c)
		return
	}

	es_ser.DeleteFullTextByArticleID(cr.ID)
	res.OkWithMessage("已驳回该文章", c)
}
