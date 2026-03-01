package tag_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/common"

	"github.com/gin-gonic/gin"
)

// TagListView 获取标签列表
// @Summary 获取标签列表
// @Description 获取标签列表，支持分页
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认10"
// @Param sort query string false "排序方式"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.TagModel}} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/tags [get]
func (TagApi) TagListView(c *gin.Context) {
	var cr models.PageInfo
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	list, count, _ := common.ComList(models.TagModel{}, common.Option{
		PageInfo: cr,
	})
	res.OkWithList(list, count, c)
}
