package article_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
)

// ArticleCollCreateView 用户收藏文章，或取消收藏
// @Tags 文章管理
// @Summary 用户收藏文章，或取消收藏
// @Description 用户收藏指定ID的文章，或取消已有的收藏
// @Param token header string true "token"
// @Param data body models.ESIDRequest true "文章ID"
// @Router /api/articles/collects [post]
// @Produce json
// @Success 200 {object} res.Response{data=string}
func (ArticleApi) ArticleCollCreateView(c *gin.Context) {
	var cr models.ESIDRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	model, err := es_ser.CommDetail(cr.ID)
	if err != nil {
		res.FailWithMessage("文章不存在", c)
		return
	}
	if !canViewArticle(model, claims) {
		res.FailWithMessage("该文章当前不可收藏", c)
		return
	}

	var coll models.UserCollectModel
	err = global.DB.Take(&coll, "user_id = ? and article_id = ?", claims.UserID, cr.ID).Error
	var num = -1
	if err != nil {
		// 没有找到 收藏文章
		global.DB.Create(&models.UserCollectModel{
			UserID:    claims.UserID,
			ArticleID: cr.ID,
		})
		// 给文章的收藏数 +1
		num = 1
	} else {
		// 取消收藏
		global.DB.Delete(&coll)
	}

	// 更新文章收藏数
	err = es_ser.ArticleUpdate(cr.ID, map[string]any{
		"collects_count": model.CollectsCount + num,
	})
	if num == 1 {
		res.OkWithMessage("收藏文章成功", c)
	} else {
		res.OkWithMessage("取消收藏成功", c)
	}
}
