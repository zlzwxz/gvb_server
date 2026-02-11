package article_api

import (
	"fmt"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/service/redis_ser"

	"github.com/gin-gonic/gin"
)

// http://127.0.0.1:8080/api/articles/_WoX6ZsBPMPoP5eKRYpO
func (ArticleApi) ArticleDetailView(c *gin.Context) {
	var cr models.ESIDRequest
	err := c.ShouldBindUri(&cr)
	fmt.Println(cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	model, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	//用户浏览的时候加上浏览量
	redis_ser.NewArticleLook().Set(cr.ID)
	res.OkWithData(model, c)
}

type ArticleDetailRequest struct {
	Title string `json:"title" form:"title"`
}

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
	res.OkWithData(model, c)
}
