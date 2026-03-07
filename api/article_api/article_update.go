package article_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/board_ser"
	"gvb-server/service/crawl_ser"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"
	"gvb-server/utils/sanitize"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ArticleUpdateRequest 文章更新请求参数
type ArticleUpdateRequest struct {
	Title       string                     `json:"title" swag:"description:文章标题"`    // 文章标题
	Abstract    string                     `json:"abstract" swag:"description:文章简介"` // 文章简介
	Content     string                     `json:"content" swag:"description:文章内容"`  // 文章内容
	BoardID     *uint                      `json:"board_id" swag:"description:板块ID"`
	Category    string                     `json:"category" swag:"description:文章分类"`    // 文章分类
	Source      string                     `json:"source" swag:"description:文章来源"`      // 文章来源
	Link        string                     `json:"link" swag:"description:原文链接"`        // 原文链接
	BannerID    uint                       `json:"banner_id" swag:"description:文章封面ID"` // 文章封面id
	Tags        []string                   `json:"tags" swag:"description:文章标签"`        // 文章标签
	Attachments []models.ArticleAttachment `json:"attachments" swag:"description:文章附件"`
	IsPrivate   *bool                      `json:"is_private" swag:"description:是否私密文章"`
	ID          string                     `json:"id" swag:"description:文章ID"`
}

// ArticleUpdateView 更新文章
// @Summary 更新文章
// @Description 更新指定ID的文章信息
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body ArticleUpdateRequest true "文章更新信息"
// @Success 200 {object} res.Response{msg=string} "更新成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "文章不存在"
// @Router /api/articles [put]
func (ArticleApi) ArticleUpdateView(c *gin.Context) {
	var cr ArticleUpdateRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		global.Log.Error(err)
		res.FailWithError(err, &cr, c)
		return
	}
	if cr.ID == "" {
		res.FailWithMessage("文章ID不能为空", c)
		return
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	currentArticle, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage("文章不存在", c)
		return
	}
	if !canManageArticle(currentArticle, claims) {
		res.FailWithMessage("无权编辑该文章", c)
		return
	}

	var bannerUrl string
	if cr.BannerID != 0 {
		err = global.DB.Model(models.BannerModel{}).Where("id = ?", cr.BannerID).Select("path").Scan(&bannerUrl).Error
		if err != nil {
			res.FailWithMessage("banner不存在", c)
			return
		}
	}

	dataMap := map[string]any{
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	if strings.TrimSpace(cr.Title) != "" {
		title := strings.TrimSpace(cr.Title)
		dataMap["title"] = title
		dataMap["keyword"] = title
	}
	if strings.TrimSpace(cr.Abstract) != "" {
		dataMap["abstract"] = strings.TrimSpace(cr.Abstract)
	}
	if strings.TrimSpace(cr.Content) != "" {
		cleanContent := sanitize.CleanMarkdownInput(cr.Content)
		if cleanContent == "" {
			res.FailWithMessage("文章内容不能为空", c)
			return
		}
		dataMap["content"] = cleanContent
	}
	if cr.BoardID != nil && *cr.BoardID > 0 {
		board, boardErr := board_ser.GetEnabledBoardByID(*cr.BoardID)
		if boardErr != nil {
			res.FailWithMessage("板块不存在或已停用", c)
			return
		}
		dataMap["board_id"] = board.ID
		dataMap["board_name"] = board.Name
		dataMap["category"] = board.Name
	}
	if strings.TrimSpace(cr.Category) != "" {
		dataMap["category"] = strings.TrimSpace(cr.Category)
	}
	if strings.TrimSpace(cr.Source) != "" {
		dataMap["source"] = strings.TrimSpace(cr.Source)
	}
	if strings.TrimSpace(cr.Link) != "" {
		cleanLink := sanitize.CleanURL(cr.Link, true)
		if cleanLink == "" {
			res.FailWithMessage("文章链接仅支持 http/https 或站内相对路径", c)
			return
		}
		dataMap["link"] = cleanLink
	}
	if cr.BannerID != 0 {
		dataMap["banner_id"] = cr.BannerID
		dataMap["banner_url"] = bannerUrl
	}
	if len(cr.Tags) > 0 {
		dataMap["tags"] = ctype.Array(cr.Tags)
	}
	if cr.Attachments != nil {
		dataMap["attachments"] = cr.Attachments
	}
	if cr.IsPrivate != nil {
		dataMap["is_private"] = *cr.IsPrivate
	}

	if !isAdmin(claims) {
		// 普通用户编辑后回到待审核状态
		dataMap["review_status"] = ctype.ArticleReviewPending
		dataMap["review_reason"] = ""
		dataMap["reviewed_at"] = ""
		dataMap["reviewer_id"] = 0
		dataMap["reviewer_nick_name"] = ""
	}

	err = es_ser.ArticleUpdate(cr.ID, dataMap)
	if err != nil {
		res.FailWithMessage("更新失败", c)
		return
	}

	newArticle, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage("更新成功，但读取文章失败", c)
		return
	}

	if rate, duplicateID, duplicateTitle, duplicateErr := crawl_ser.CalculateArticleDuplicateRate(cr.ID, newArticle.Title, newArticle.Content); duplicateErr == nil {
		newArticle.DuplicateRate = rate
		newArticle.DuplicateTargetID = duplicateID
		newArticle.DuplicateTargetTitle = duplicateTitle
		_ = es_ser.ArticleUpdate(cr.ID, map[string]any{
			"duplicate_rate":         rate,
			"duplicate_target_id":    duplicateID,
			"duplicate_target_title": duplicateTitle,
		})
	}

	if (newArticle.ReviewStatus == ctype.ArticleReviewApproved || newArticle.ReviewStatus == ctype.ArticleReviewLegacy) && !newArticle.IsPrivate {
		es_ser.DeleteFullTextByArticleID(cr.ID)
		es_ser.AsyncArticleByFullText(es_ser.SearchData{
			Key:   cr.ID,
			Body:  newArticle.Content,
			Slug:  es_ser.GetSlug(newArticle.Title),
			Title: newArticle.Title,
		})
	} else {
		es_ser.DeleteFullTextByArticleID(cr.ID)
	}

	if !isAdmin(claims) {
		res.OkWithMessage("更新成功，等待管理员重新审核", c)
		return
	}
	res.OkWithMessage("更新成功", c)
}
