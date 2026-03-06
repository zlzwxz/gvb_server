package data_api

import (
	"context"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

// DataSumResponse 数据统计响应结构
type DataSumResponse struct {
	UserCount      int `json:"user_count" swag:"description:用户总数"`
	ArticleCount   int `json:"article_count" swag:"description:文章总数"`
	MessageCount   int `json:"message_count" swag:"description:消息总数"`
	ChatGroupCount int `json:"chat_group_count" swag:"description:聊天群组总数"`
	NowLoginCount  int `json:"now_login_count" swag:"description:今日登录人数"`
	NowSignCount   int `json:"now_sign_count" swag:"description:今日注册人数"`
	LookTotal      int `json:"look_total" swag:"description:阅读总量"`
	CommentTotal   int `json:"comment_total" swag:"description:评论总量"`
	DiggTotal      int `json:"digg_total" swag:"description:点赞总量"`
	CollectTotal   int `json:"collect_total" swag:"description:收藏总量"`
}

// DataSumView 获取数据统计
// @Summary 获取数据统计
// @Description 获取系统的各项数据统计，包括用户、文章、消息、聊天群组等数量
// @Tags 数据管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=DataSumResponse} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/data_sum [get]
func (DataApi) DataSumView(c *gin.Context) {
	var userCount, articleCount, messageCount, chatGroupCount int
	var nowLoginCount, nowSignCount int

	result, _ := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		TrackTotalHits(true).
		Query(elastic.NewMatchAllQuery()).
		Size(0).
		Aggregation("sum_look", elastic.NewSumAggregation().Field("look_count")).
		Aggregation("sum_comment", elastic.NewSumAggregation().Field("comment_count")).
		Aggregation("sum_digg", elastic.NewSumAggregation().Field("digg_count")).
		Aggregation("sum_collect", elastic.NewSumAggregation().Field("collects_count")).
		Do(context.Background())
	if result != nil && result.Hits != nil && result.Hits.TotalHits != nil {
		articleCount = int(result.Hits.TotalHits.Value)
	}

	lookTotal := int(aggregationValue(result, "sum_look")) + sumMapValue(redis_ser.NewArticleLook().GetInfo())
	commentTotal := int(aggregationValue(result, "sum_comment")) + sumMapValue(redis_ser.NewCommentCount().GetInfo())
	diggTotal := int(aggregationValue(result, "sum_digg")) + sumMapValue(redis_ser.NewDigg().GetInfo())
	collectTotal := int(aggregationValue(result, "sum_collect"))

	global.DB.Model(models.UserModel{}).Select("count(id)").Scan(&userCount)
	global.DB.Model(models.MessageModel{}).Select("count(id)").Scan(&messageCount)
	global.DB.Model(models.ChatModel{IsGroup: true}).Select("count(id)").Scan(&chatGroupCount)
	global.DB.Model(models.LoginDataModel{}).Where("to_days(created_at)=to_days(now())").
		Select("count(id)").Scan(&nowLoginCount)
	global.DB.Model(models.UserModel{}).Where("to_days(created_at)=to_days(now())").
		Select("count(id)").Scan(&nowSignCount)

	res.OkWithData(DataSumResponse{
		UserCount:      userCount,
		ArticleCount:   articleCount,
		MessageCount:   messageCount,
		ChatGroupCount: chatGroupCount,
		NowLoginCount:  nowLoginCount,
		NowSignCount:   nowSignCount,
		LookTotal:      lookTotal,
		CommentTotal:   commentTotal,
		DiggTotal:      diggTotal,
		CollectTotal:   collectTotal,
	}, c)
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

func sumMapValue(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}
