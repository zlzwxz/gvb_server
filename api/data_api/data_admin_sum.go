package data_api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
)

// AdminDataSumResponse 后台运营统计。
type AdminDataSumResponse struct {
	PendingArticleReviewCount int `json:"pending_article_review_count"`
	PendingArticleReportCount int `json:"pending_article_report_count"`
	AnnouncementCount         int `json:"announcement_count"`
	EnabledBoardCount         int `json:"enabled_board_count"`
}

// AdminDataSumView 获取后台运营统计。
func (DataApi) AdminDataSumView(c *gin.Context) {
	var pendingArticleReportCount, announcementCount, enabledBoardCount int64

	global.DB.Model(models.ArticleReportModel{}).
		Where("status = ?", ctype.ArticleReportPending).
		Count(&pendingArticleReportCount)
	global.DB.Model(models.AnnouncementModel{}).
		Count(&announcementCount)
	global.DB.Model(models.BoardModel{}).
		Where("is_enabled = ?", true).
		Count(&enabledBoardCount)

	pendingArticleReviewCount := 0
	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		TrackTotalHits(true).
		Query(elastic.NewTermQuery("review_status", int(ctype.ArticleReviewPending))).
		Size(0).
		Do(context.Background())
	if err == nil && result != nil && result.Hits != nil && result.Hits.TotalHits != nil {
		pendingArticleReviewCount = int(result.Hits.TotalHits.Value)
	}

	res.OkWithData(AdminDataSumResponse{
		PendingArticleReviewCount: pendingArticleReviewCount,
		PendingArticleReportCount: int(pendingArticleReportCount),
		AnnouncementCount:         int(announcementCount),
		EnabledBoardCount:         int(enabledBoardCount),
	}, c)
}
