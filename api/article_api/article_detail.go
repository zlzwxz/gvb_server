package article_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/service/redis_ser"

	"github.com/gin-gonic/gin"
)

// ArticleDetailView 获取文章详情
// @Summary 获取文章详情
// @Description 根据文章ID获取文章详细信息
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param id path string true "文章ID"
// @Success 200 {object} res.Response{data=models.ArticleModel} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 404 {object} res.Response "文章不存在"
// @Router /articles/{id} [get]
func (ArticleApi) ArticleDetailView(c *gin.Context) {
	var cr models.ESIDRequest
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	model, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	if !canViewArticle(model, optionalClaims(c)) {
		res.FailWithMessage("文章审核中或已驳回", c)
		return
	}
	//用户浏览的时候加上浏览量
	redis_ser.NewArticleLook().Set(cr.ID)
	res.OkWithData(model, c)
}

// ArticleDetailRequest 文章详情请求参数
type ArticleDetailRequest struct {
	Title string `json:"title" form:"title" swag:"description:文章标题"`
}

// ArticleDetailByTitleView 根据标题获取文章详情
// @Summary 根据标题获取文章详情
// @Description 根据文章标题关键词获取文章详细信息
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param title query string true "文章标题"
// @Success 200 {object} res.Response{data=models.ArticleModel} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 404 {object} res.Response "文章不存在"
// @Router /api/articles/detail [get]
func (ArticleApi) ArticleDetailByTitleView(c *gin.Context) {
	var cr ArticleDetailRequest
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	model, err := es_ser.CommDetailByKeyword(cr.Title)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	if !canViewArticle(model, optionalClaims(c)) {
		res.FailWithMessage("文章审核中或已驳回", c)
		return
	}
	res.OkWithData(model, c)
}
