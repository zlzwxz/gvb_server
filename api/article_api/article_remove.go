package article_api

import (
	"context"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

// IDListRequest 文章ID列表请求参数
type IDListRequest struct {
	IDList []string `json:"id_list" swag:"description:文章ID列表"`
}

// ArticleRemoveView 批量删除文章
// @Summary 批量删除文章
// @Description 根据文章ID列表批量删除文章
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body IDListRequest true "文章ID列表"
// @Success 200 {object} res.Response{msg=string} "删除成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles [delete]
func (ArticleApi) ArticleRemoveView(c *gin.Context) {
	var cr IDListRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		global.Log.Error(err)
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	bulkService := global.ESClient.Bulk().Index(models.ArticleModel{}.Index()).Refresh("true")
	for _, id := range cr.IDList {
		req := elastic.NewBulkDeleteRequest().Id(id)
		//同步删除全文搜索索引数据
		go es_ser.DeleteFullTextByArticleID(id)
		bulkService.Add(req)
	}
	result, err := bulkService.Do(context.Background())
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("删除失败", c)
		return
	}
	res.OkWithMessage(fmt.Sprintf("成功删除 %d 篇文章", len(result.Succeeded())), c)
	return

}
