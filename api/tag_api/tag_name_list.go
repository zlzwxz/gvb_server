package tag_api

import (
	"context"
	"encoding/json"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

// TagResponse 标签响应结构
type TagResponse struct {
	Label string `json:"label" swag:"description:标签显示名"`
	Value string `json:"value" swag:"description:标签值"`
}

// TagNameListView 获取标签名称列表
// @Summary 获取标签名称列表
// @Description 获取所有文章标签的名称列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=[]TagResponse} "获取成功"
// @Failure 500 {object} res.Response "查询失败"
// @Router /api/tags/names [get]
func (TagApi) TagNameListView(c *gin.Context) {
	type T struct {
		DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int `json:"sum_other_doc_count"`
		Buckets                 []struct {
			Key      string `json:"key"`
			DocCount int    `json:"doc_count"`
		} `json:"buckets"`
	}
	query := elastic.NewBoolQuery()
	agg := elastic.NewTermsAggregation().Field("tags")
	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(query).
		Aggregation("tags", agg).
		Size(0).
		Do(context.Background())
	if err != nil {
		logrus.Error(err)
		return
	}
	byteData := result.Aggregations["tags"]
	var tagType T
	json.Unmarshal(byteData, &tagType)

	var tagList = make([]TagResponse, 0)
	for _, bucket := range tagType.Buckets {
		tagList = append(tagList, TagResponse{
			Label: bucket.Key,
			Value: bucket.Key,
		})
	}

	res.OkWithData(tagList, c)

}
