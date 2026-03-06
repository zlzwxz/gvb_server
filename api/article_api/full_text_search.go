package article_api

import (
	"context"
	"encoding/json"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

// FullTextSearchView 全文文章搜索
// @Summary 全文文章搜索
// @Description 根据关键词搜索文章标题和正文内容
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param key query string false "搜索关键词"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.FullTextModel}} "搜索成功"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/articles/search [get]
func (ArticleApi) FullTextSearchView(c *gin.Context) {
	var cr models.PageInfo
	_ = c.ShouldBindQuery(&cr)

	boolQuery := elastic.NewBoolQuery()

	fmt.Println(cr.Key)
	if cr.Key != "" && cr.Id == "" {
		boolQuery.Must(elastic.NewMultiMatchQuery(cr.Key, "title", "body"))
	}
	//如果前端传了文章id值，那么应该是精确查询，应该根据id查询单条文章
	if cr.Id != "" && cr.Key == "" {
		boolQuery.Must(elastic.NewTermQuery("key", cr.Id))
	}

	result, err := global.ESClient.
		Search(models.FullTextModel{}.Index()).
		Query(boolQuery).
		Highlight(elastic.NewHighlight().Field("body")).
		Size(100).
		Do(context.Background())
	if err != nil {
		return
	}
	count := result.Hits.TotalHits.Value //搜索到结果总条数
	fullTextList := make([]models.FullTextModel, 0)
	for _, hit := range result.Hits.Hits {
		var model models.FullTextModel
		json.Unmarshal(hit.Source, &model)

		body, ok := hit.Highlight["body"]
		if ok {
			model.Body = body[0]
		}

		fullTextList = append(fullTextList, model)
	}

	res.OkWithList(fullTextList, count, c)
}
