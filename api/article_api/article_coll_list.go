package article_api

import (
	"context"
	"encoding/json"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type CollResponse struct {
	models.ArticleModel
	CreatedAt string `json:"created_at"`
}

// ArticleCollListView 获取用户收藏的文章列表
// @Tags 文章管理
// @Summary 获取用户收藏的文章列表
// @Description 获取当前登录用户收藏的文章列表，支持分页
// @Param token header string true "token"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认10"
// @Router /api/articles/collects [get]
// @Produce json
// @Success 200 {object} res.Response{data=res.ListResponse{list=[]CollResponse}}
func (ArticleApi) ArticleCollListView(c *gin.Context) {

	var cr models.PageInfo

	c.ShouldBindQuery(&cr)

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	var articleIDList []interface{}

	list, count, err := common.ComList(models.UserCollectModel{UserID: claims.UserID}, common.Option{
		PageInfo: cr,
	})

	var collMap = map[string]string{}

	for _, model := range list {
		articleIDList = append(articleIDList, model.ArticleID)
		collMap[model.ArticleID] = model.CreatedAt.Format("2006-01-02 15:04:05")
	}

	boolSearch := elastic.NewTermsQuery("_id", articleIDList...)

	var collList = make([]CollResponse, 0)

	// 传id列表，查es
	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(boolSearch).
		Size(1000).
		Do(context.Background())
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	fmt.Println(result.Hits.TotalHits.Value, articleIDList)

	for _, hit := range result.Hits.Hits {
		var article models.ArticleModel
		err = json.Unmarshal(hit.Source, &article)
		if err != nil {
			global.Log.Error(err)
			continue
		}
		article.ID = hit.Id
		collList = append(collList, CollResponse{
			ArticleModel: article,
			CreatedAt:    collMap[hit.Id],
		})
	}
	res.OkWithList(collList, count, c)
}
