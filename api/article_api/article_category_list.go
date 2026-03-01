package article_api

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

type CategoryResponse struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ArticleCategoryListView 获取文章分类列表
// @Tags 文章管理
// @Summary 获取文章分类列表
// @Description 获取所有文章分类的列表数据
// @Router /api/articles/categorys [get]
// @Produce json
// @Success 200 {object} res.Response{data=[]CategoryResponse}
func (ArticleApi) ArticleCategoryListView(c *gin.Context) {
	//文章分类列表
	type T struct {
		DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int `json:"sum_other_doc_count"`
		Buckets                 []struct {
			Key      string `json:"key"`
			DocCount int    `json:"doc_count"`
		} `json:"buckets"`
	}

	agg := elastic.NewTermsAggregation().Field("category")
	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(elastic.NewBoolQuery()).
		Aggregation("categorys", agg).
		Size(0).
		Do(context.Background())
	if err != nil {
		logrus.Error(err)
		return
	}
	byteData := result.Aggregations["categorys"]
	var categoryType T
	_ = json.Unmarshal(byteData, &categoryType)
	var categoryList = make([]CategoryResponse, 0)
	for _, i2 := range categoryType.Buckets {
		categoryList = append(categoryList, CategoryResponse{
			Label: i2.Key,
			Value: i2.Key,
		})
	}
	res.OkWithData(categoryList, c)

}
