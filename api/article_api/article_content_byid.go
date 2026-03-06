package article_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
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
	//根据文章id获取文章正文
	var cr models.ESIDRequest
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	article, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage("查询失败", c)
		return
	}
	if !canViewArticle(article, optionalClaims(c)) {
		res.FailWithMessage("文章审核中或已驳回", c)
		return
	}
	redis_ser.NewArticleLook().Set(cr.ID)
	res.OkWithData(article.Content, c)
}
