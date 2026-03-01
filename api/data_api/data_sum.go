package data_api

import (
	"context"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

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

	var userCount, articleCount, messageCount, ChatGroupCount int
	var nowLoginCount, nowSignCount int

	result, _ := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(elastic.NewMatchAllQuery()).
		Do(context.Background())
	articleCount = int(result.Hits.TotalHits.Value) //搜索到结果总条数
	global.DB.Model(models.UserModel{}).Select("count(id)").Scan(&userCount)
	global.DB.Model(models.MessageModel{}).Select("count(id)").Scan(&messageCount)
	global.DB.Model(models.ChatModel{IsGroup: true}).Select("count(id)").Scan(&ChatGroupCount)
	global.DB.Model(models.LoginDataModel{}).Where("to_days(created_at)=to_days(now())").
		Select("count(id)").Scan(&nowLoginCount)
	global.DB.Model(models.UserModel{}).Where("to_days(created_at)=to_days(now())").
		Select("count(id)").Scan(&nowSignCount)

	res.OkWithData(DataSumResponse{
		UserCount:      userCount,
		ArticleCount:   articleCount,
		MessageCount:   messageCount,
		ChatGroupCount: ChatGroupCount,
		NowLoginCount:  nowLoginCount,
		NowSignCount:   nowSignCount,
	}, c)
}
