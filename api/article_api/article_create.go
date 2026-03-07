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
	"math/rand"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

type ArticleRequest struct {
	Title       string                     `json:"title" binding:"required" msg:"文章标题必填"`   // 文章标题
	Abstract    string                     `json:"abstract"`                                // 文章简介
	Content     string                     `json:"content" binding:"required" msg:"文章内容必填"` // 文章内容
	BoardID     uint                       `json:"board_id" binding:"required" msg:"请选择板块"` // 文章板块
	Category    string                     `json:"category"`                                // 文章分类
	Source      string                     `json:"source"`                                  // 文章来源
	Link        string                     `json:"link"`                                    // 原文链接
	BannerID    uint                       `json:"banner_id"`                               // 文章封面id
	Tags        ctype.Array                `json:"tags"`                                    // 文章标签
	Attachments []models.ArticleAttachment `json:"attachments"`                             // 文章附件
	IsPrivate   bool                       `json:"is_private"`                              // 私密文章开关（仅作者/管理员可见）
}

// ArticleCreateView 创建新文章
// @Summary 创建新文章
// @Description 用户创建新文章，需要登录认证，支持富文本内容
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body ArticleRequest true "文章信息"
// @Success 200 {object} res.Response{msg=string} "创建成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles [post]
func (ArticleApi) ArticleCreateView(c *gin.Context) {
	var cr ArticleRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	cr.Title = strings.TrimSpace(cr.Title)
	cr.Abstract = strings.TrimSpace(cr.Abstract)
	cr.Category = strings.TrimSpace(cr.Category)
	cr.Source = strings.TrimSpace(cr.Source)
	cr.Content = sanitize.CleanMarkdownInput(cr.Content)
	if cr.Content == "" {
		res.FailWithMessage("文章内容不能为空", c)
		return
	}
	if strings.TrimSpace(cr.Link) != "" {
		cr.Link = sanitize.CleanURL(cr.Link, true)
		if cr.Link == "" {
			res.FailWithMessage("文章链接仅支持 http/https 或站内相对路径", c)
			return
		}
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	userID := claims.UserID
	userNickName := claims.NickName
	board, err := board_ser.GetEnabledBoardByID(cr.BoardID)
	if err != nil {
		res.FailWithMessage("板块不存在或已停用", c)
		return
	}
	reviewStatus := ctype.ArticleReviewPending
	reviewedAt := ""
	reviewerID := uint(0)
	reviewerNickName := ""
	if claims.Role == int(ctype.PermissionAdmin) {
		reviewStatus = ctype.ArticleReviewApproved
		reviewedAt = time.Now().Format("2006-01-02 15:04:05")
		reviewerID = claims.UserID
		reviewerNickName = claims.NickName
	}
	// 处理content
	unsafe := blackfriday.MarkdownCommon([]byte(cr.Content))
	// 是不是有script标签
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(unsafe)))
	//fmt.Println(doc.Text())
	nodes := doc.Find("script").Nodes
	if len(nodes) > 0 {
		// 有script标签
		doc.Find("script").Remove()
		converter := md.NewConverter("", true, nil)
		html, _ := doc.Html()
		markdown, _ := converter.ConvertString(html)
		cr.Content = markdown
	}
	if cr.Abstract == "" {
		// 汉字的截取不一样
		abs := []rune(doc.Text())
		// 将content转为html，并且过滤xss，以及获取中文内容
		if len(abs) > 100 {
			cr.Abstract = string(abs[:100])
		} else {
			cr.Abstract = string(abs)
		}
	}

	// 不传banner_id,后台就随机去选择一张
	if cr.BannerID == 0 {
		var bannerIDList []uint
		global.DB.Model(models.BannerModel{}).Select("id").Scan(&bannerIDList)
		if len(bannerIDList) == 0 {
			res.FailWithMessage("没有banner数据", c)
			return
		}
		rand.Seed(time.Now().UnixNano())
		cr.BannerID = bannerIDList[rand.Intn(len(bannerIDList))]
	}

	// 查banner_id下的banner_url
	var bannerUrl string
	err = global.DB.Model(models.BannerModel{}).Where("id = ?", cr.BannerID).Select("path").Scan(&bannerUrl).Error
	if err != nil {
		res.FailWithMessage("banner不存在", c)
		return
	}

	// 查用户头像
	var avatar string
	err = global.DB.Model(models.UserModel{}).Where("id = ?", userID).Select("avatar").Scan(&avatar).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}

	//文章title是否存在
	isExist := models.ArticleModel{Title: cr.Title}.ISExistData()
	if isExist {
		res.FailWithMessage("文章标题已存在", c)
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	article := models.ArticleModel{
		CreatedAt:        now,
		UpdatedAt:        now,
		Title:            cr.Title,
		Keyword:          cr.Title,
		Abstract:         cr.Abstract,
		Content:          cr.Content,
		UserID:           userID,
		UserNickName:     userNickName,
		UserAvatar:       avatar,
		BoardID:          board.ID,
		BoardName:        board.Name,
		Category:         board.Name,
		Source:           cr.Source,
		Link:             cr.Link,
		BannerID:         cr.BannerID,
		BannerUrl:        bannerUrl,
		Tags:             cr.Tags,
		Attachments:      cr.Attachments,
		IsPrivate:        cr.IsPrivate,
		ReviewStatus:     reviewStatus,
		ReviewedAt:       reviewedAt,
		ReviewerID:       reviewerID,
		ReviewerNickName: reviewerNickName,
	}

	err = article.Create()
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage(err.Error(), c)
		return
	}

	if rate, duplicateID, duplicateTitle, duplicateErr := crawl_ser.CalculateArticleDuplicateRate(article.ID, article.Title, article.Content); duplicateErr == nil {
		article.DuplicateRate = rate
		article.DuplicateTargetID = duplicateID
		article.DuplicateTargetTitle = duplicateTitle
		_ = es_ser.ArticleUpdate(article.ID, map[string]any{
			"duplicate_rate":         rate,
			"duplicate_target_id":    duplicateID,
			"duplicate_target_title": duplicateTitle,
		})
	}
	// 审核通过 + 非私密的文章才进入全文索引，避免私密内容被公开搜索到。
	if reviewStatus == ctype.ArticleReviewApproved && !article.IsPrivate {
		es_ser.AsyncArticleByFullText(es_ser.SearchData{
			Key:   article.ID,
			Body:  article.Content,
			Slug:  es_ser.GetSlug(article.Title),
			Title: article.Title,
		})
		res.OkWithMessage("文章发布成功", c)
		return
	}

	res.OkWithMessage("文章发布成功，等待管理员审核", c)

}
