package digg_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"

	"github.com/gin-gonic/gin"
)

// DiggArticleView 文章点赞
// @Summary 文章点赞
// @Description 对指定文章进行点赞
// @Tags 点赞管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body models.ESIDRequest true "文章ID"
// @Success 200 {object} res.Response{msg=string} "点赞成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/article/digg [post]
func (DiggApi) DiggArticleView(c *gin.Context) {
	var cr models.ESIDRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	// 对长度校验
	// 查es
	redis_ser.NewDigg().Set(cr.ID)
	res.OkWithMessage("文章点赞成功", c)
}
