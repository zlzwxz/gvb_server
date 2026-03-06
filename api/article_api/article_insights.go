package article_api

import (
	"context"
	"encoding/json"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type ArticleInsightsQuery struct {
	HotSize    int `form:"hot_size"`
	LatestSize int `form:"latest_size"`
}

type ArticleInsightsStats struct {
	ArticleCount int `json:"article_count"`
	LookTotal    int `json:"look_total"`
	CommentTotal int `json:"comment_total"`
	DiggTotal    int `json:"digg_total"`
	CollectTotal int `json:"collect_total"`
}

type TagTrend struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type ArticleInsightsResponse struct {
	Stats          ArticleInsightsStats  `json:"stats"`
	HotArticles    []models.ArticleModel `json:"hot_articles"`
	LatestArticles []models.ArticleModel `json:"latest_articles"`
	HotTags        []TagTrend            `json:"hot_tags"`
}

type scoredArticle struct {
	Article models.ArticleModel
	Score   float64
}

// ArticleInsightsView 首页企业级洞察数据
// @Summary 获取首页洞察数据
// @Description 返回统计数据、热门文章、最新文章、热门标签
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param hot_size query int false "热门文章数量，默认6"
// @Param latest_size query int false "最新文章数量，默认8"
// @Success 200 {object} res.Response{data=ArticleInsightsResponse}
// @Router /api/articles/insights [get]
func (ArticleApi) ArticleInsightsView(c *gin.Context) {
	var cr ArticleInsightsQuery
	_ = c.ShouldBindQuery(&cr)

	if cr.HotSize <= 0 {
		cr.HotSize = 6
	}
	if cr.HotSize > 20 {
		cr.HotSize = 20
	}
	if cr.LatestSize <= 0 {
		cr.LatestSize = 8
	}
	if cr.LatestSize > 20 {
		cr.LatestSize = 20
	}

	indexName := models.ArticleModel{}.Index()

	diggInfo := redis_ser.NewDigg().GetInfo()
	lookInfo := redis_ser.NewArticleLook().GetInfo()
	commentInfo := redis_ser.NewCommentCount().GetInfo()

	hotCandidates, err := global.ESClient.
		Search(indexName).
		Query(publicVisibleQuery()).
		Sort("created_at", false).
		Size(200).
		Do(context.Background())
	if err != nil {
		res.FailWithMessage("获取洞察数据失败", c)
		return
	}

	latestResult, err := global.ESClient.
		Search(indexName).
		Query(publicVisibleQuery()).
		Sort("created_at", false).
		Size(cr.LatestSize).
		Do(context.Background())
	if err != nil {
		res.FailWithMessage("获取洞察数据失败", c)
		return
	}

	totalResult, err := global.ESClient.
		Search(indexName).
		TrackTotalHits(true).
		Query(publicVisibleQuery()).
		Size(0).
		Aggregation("sum_look", elastic.NewSumAggregation().Field("look_count")).
		Aggregation("sum_comment", elastic.NewSumAggregation().Field("comment_count")).
		Aggregation("sum_digg", elastic.NewSumAggregation().Field("digg_count")).
		Aggregation("sum_collect", elastic.NewSumAggregation().Field("collects_count")).
		Do(context.Background())
	if err != nil {
		res.FailWithMessage("获取洞察统计失败", c)
		return
	}

	hotArticles := pickHotArticles(hotCandidates.Hits.Hits, diggInfo, lookInfo, commentInfo, cr.HotSize)
	latestArticles := parseArticles(latestResult.Hits.Hits, diggInfo, lookInfo, commentInfo)
	hotTags := extractHotTags(hotArticles, 12)

	lookBase := aggregationValue(totalResult, "sum_look")
	commentBase := aggregationValue(totalResult, "sum_comment")
	diggBase := aggregationValue(totalResult, "sum_digg")
	collectBase := aggregationValue(totalResult, "sum_collect")

	stats := ArticleInsightsStats{
		ArticleCount: int(totalResult.Hits.TotalHits.Value),
		LookTotal:    int(lookBase) + sumMapValue(lookInfo),
		CommentTotal: int(commentBase) + sumMapValue(commentInfo),
		DiggTotal:    int(diggBase) + sumMapValue(diggInfo),
		CollectTotal: int(collectBase),
	}

	res.OkWithData(ArticleInsightsResponse{
		Stats:          stats,
		HotArticles:    hotArticles,
		LatestArticles: latestArticles,
		HotTags:        hotTags,
	}, c)
}

func parseArticles(hits []*elastic.SearchHit, diggInfo map[string]int, lookInfo map[string]int, commentInfo map[string]int) []models.ArticleModel {
	list := make([]models.ArticleModel, 0, len(hits))
	for _, hit := range hits {
		var article models.ArticleModel
		if err := json.Unmarshal(hit.Source, &article); err != nil {
			continue
		}
		article.ID = hit.Id
		article.DiggCount += diggInfo[hit.Id]
		article.LookCount += lookInfo[hit.Id]
		article.CommentCount += commentInfo[hit.Id]
		list = append(list, article)
	}
	return list
}

func pickHotArticles(hits []*elastic.SearchHit, diggInfo map[string]int, lookInfo map[string]int, commentInfo map[string]int, limit int) []models.ArticleModel {
	scored := make([]scoredArticle, 0, len(hits))
	for _, hit := range hits {
		var article models.ArticleModel
		if err := json.Unmarshal(hit.Source, &article); err != nil {
			continue
		}
		article.ID = hit.Id
		article.DiggCount += diggInfo[hit.Id]
		article.LookCount += lookInfo[hit.Id]
		article.CommentCount += commentInfo[hit.Id]

		score := float64(article.LookCount) + float64(article.DiggCount)*3 + float64(article.CommentCount)*2 + float64(article.CollectsCount)*2
		scored = append(scored, scoredArticle{
			Article: article,
			Score:   score,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].Article.CreatedAt > scored[j].Article.CreatedAt
		}
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	list := make([]models.ArticleModel, 0, len(scored))
	for _, item := range scored {
		list = append(list, item.Article)
	}
	return list
}

func extractHotTags(articles []models.ArticleModel, limit int) []TagTrend {
	tagCounter := map[string]int{}
	for _, article := range articles {
		for _, tag := range article.Tags {
			if tag == "" {
				continue
			}
			tagCounter[tag]++
		}
	}

	list := make([]TagTrend, 0, len(tagCounter))
	for tag, count := range tagCounter {
		list = append(list, TagTrend{Tag: tag, Count: count})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].Tag < list[j].Tag
		}
		return list[i].Count > list[j].Count
	})

	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

func sumMapValue(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}

func aggregationValue(result *elastic.SearchResult, name string) float64 {
	if result == nil || result.Aggregations == nil {
		return 0
	}
	agg, ok := result.Aggregations.Sum(name)
	if !ok || agg == nil || agg.Value == nil {
		return 0
	}
	return *agg.Value
}
