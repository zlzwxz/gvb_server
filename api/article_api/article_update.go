package article_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

// ArticleUpdateRequest 文章更新请求参数
type ArticleUpdateRequest struct {
	Title       string                     `json:"title" swag:"description:文章标题"`       // 文章标题
	Abstract    string                     `json:"abstract" swag:"description:文章简介"`    // 文章简介
	Content     string                     `json:"content" swag:"description:文章内容"`     // 文章内容
	Category    string                     `json:"category" swag:"description:文章分类"`    // 文章分类
	Source      string                     `json:"source" swag:"description:文章来源"`      // 文章来源
	Link        string                     `json:"link" swag:"description:原文链接"`        // 原文链接
	BannerID    uint                       `json:"banner_id" swag:"description:文章封面ID"` // 文章封面id
	Tags        []string                   `json:"tags" swag:"description:文章标签"`        // 文章标签
	Attachments []models.ArticleAttachment `json:"attachments" swag:"description:文章附件"`
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
	if cr.Title != "" {
		dataMap["title"] = cr.Title
		dataMap["keyword"] = cr.Title
	}
	if cr.Abstract != "" {
		dataMap["abstract"] = cr.Abstract
	}
	if cr.Content != "" {
		dataMap["content"] = cr.Content
	}
	if cr.Category != "" {
		dataMap["category"] = cr.Category
	}
	if cr.Source != "" {
		dataMap["source"] = cr.Source
	}
	if cr.Link != "" {
		dataMap["link"] = cr.Link
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

	if newArticle.ReviewStatus == ctype.ArticleReviewApproved || newArticle.ReviewStatus == ctype.ArticleReviewLegacy {
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
