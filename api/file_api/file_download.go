package file_api

import (
	"context"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gvb-server/utils/jwts"
)

type fileDownloadURI struct {
	ID uint `uri:"id" binding:"required"`
}

// FileDownloadView 下载文章附件
func (FileApi) FileDownloadView(c *gin.Context) {
	claimsAny, ok := c.Get("claims")
	if !ok {
		res.FailWithMessage("未登录", c)
		return
	}
	claims := claimsAny.(*jwts.CustomClaims)

	var cr fileDownloadURI
	if err := c.ShouldBindUri(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var record models.ArticleFileModel
	if err := global.DB.Take(&record, cr.ID).Error; err != nil {
		res.FailWithMessage("附件不存在", c)
		return
	}
	if claims.Role != int(ctype.PermissionAdmin) && record.UserID != claims.UserID && !isPublicAttachment(record.ID) {
		res.FailWithMessage("无权下载该附件", c)
		return
	}

	localPath := strings.TrimPrefix(record.Path, "/")
	localPath = filepath.Clean(localPath)
	if strings.HasPrefix(localPath, "..") {
		res.FailWithMessage("非法附件路径", c)
		return
	}
	if _, err := os.Stat(localPath); err != nil {
		res.FailWithMessage("附件文件不存在", c)
		return
	}
	uploadRoot := filepath.Clean(global.Config.Upload.Path)
	if !strings.HasPrefix(filepath.ToSlash(localPath), filepath.ToSlash(uploadRoot)+"/") {
		res.FailWithMessage("附件路径非法", c)
		return
	}

	c.FileAttachment(localPath, record.Name)
}

func isPublicAttachment(fileID uint) bool {
	if fileID == 0 {
		return false
	}

	attachmentQuery := elastic.NewNestedQuery(
		"attachments",
		elastic.NewTermQuery("attachments.file_id", int(fileID)),
	)
	visibleQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery("review_status", int(ctype.ArticleReviewApproved)),
			elastic.NewTermQuery("review_status", int(ctype.ArticleReviewLegacy)),
			elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("review_status")),
		).
		MinimumNumberShouldMatch(1)
	privateQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery("is_private", false),
			elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("is_private")),
		).
		MinimumNumberShouldMatch(1)

	query := elastic.NewBoolQuery().Must(attachmentQuery, visibleQuery, privateQuery)
	result, err := global.ESClient.Search("article_index").Query(query).Size(1).Do(context.Background())
	if err != nil {
		return false
	}
	return result.Hits != nil && result.Hits.TotalHits != nil && result.Hits.TotalHits.Value > 0
}
