package article_api

import (
	"context"
	"encoding/json"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"

	"github.com/gin-gonic/gin"
)

// ArticleContentByIDView 获取文章正文
// @Tags 文章管理
// @Summary 获取文章正文
// @Description 根据文章ID获取文章正文内容
// @Param id path string true "文章ID"
// @Router /api/articles/content/{id} [get]
// @Produce json
// @Success 200 {object} res.Response{data=string}
func (ArticleApi) ArticleContentByIDView(c *gin.Context) {
	fmt.Println("1232")
	//根据文章id获取文章正文
	var cr models.ESIDRequest
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	fmt.Println(cr.ID, "1232")
	redis_ser.NewArticleLook().Set(cr.ID)
	result, err := global.ESClient.Get().
		Index(models.ArticleModel{}.Index()).
		Id(cr.ID).
		Do(context.Background())
	if err != nil {
		res.FailWithMessage("查询失败", c)
		return
	}
	var model models.ArticleModel
	err = json.Unmarshal(result.Source, &model)
	if err != nil {
		return
	}
	res.OkWithData(model.Content, c)
}
