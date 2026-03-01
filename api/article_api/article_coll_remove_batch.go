package article_api

import (
	"context"
	"encoding/json"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

// ArticleCollBatchRemoveView 批量取消收藏文章
// @Tags 文章管理
// @Summary 批量取消收藏文章
// @Description 批量取消当前登录用户收藏的文章
// @Param token header string true "token"
// @Param data body models.ESIDListRequest true "文章ID列表"
// @Router /api/articles/collects/batch [delete]
// @Produce json
// @Success 200 {object} res.Response{data=string}
func (ArticleApi) ArticleCollBatchRemoveView(c *gin.Context) {
	var cr models.ESIDListRequest

	err := c.ShouldBindJSON(&cr)
	fmt.Println(cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	var collects []models.UserCollectModel
	var articleIDList []string
	global.DB.Model(&models.UserCollectModel{}).
		Where("user_id = ? and article_id in ?", claims.UserID, cr.IDList).
		Select("article_id").
		Scan(&articleIDList)
	if len(articleIDList) == 0 {
		res.FailWithMessage("请求非法", c)
		return
	}
	var idList []interface{}
	for _, s := range articleIDList {
		idList = append(idList, s)
	}
	// 更新文章数
	boolSearch := elastic.NewTermsQuery("_id", idList...)
	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(boolSearch).
		Size(1000).
		Do(context.Background())
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	for _, hit := range result.Hits.Hits {
		var article models.ArticleModel
		err = json.Unmarshal(hit.Source, &article)
		if err != nil {
			global.Log.Error(err)
			continue
		}
		count := article.CollectsCount - 1
		err = es_ser.ArticleUpdate(hit.Id, map[string]any{
			"collects_count": count,
		})
		if err != nil {
			global.Log.Error(err)
			continue
		}
	}
	// 删除用户收藏
	global.DB.Delete(&collects, "user_id = ? and article_id in ?", claims.UserID, cr.IDList)
	res.OkWithMessage(fmt.Sprintf("成功取消收藏 %d 篇文章", len(articleIDList)), c)

}
