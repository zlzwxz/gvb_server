package comment_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"

	"github.com/gin-gonic/gin"
)

// CommentIDRequest 评论ID请求参数
type CommentIDRequest struct {
	ID uint `json:"id" form:"id" uri:"id" swag:"description:评论ID"`
}

// CommentDigg 评论点赞
// @Summary 评论点赞
// @Description 对指定评论进行点赞
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param id path uint true "评论ID"
// @Success 200 {object} res.Response{msg=string} "点赞成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 404 {object} res.Response "评论不存在"
// @Router /api/comments/digg/{id} [post]
func (CommentApi) CommentDigg(c *gin.Context) {
	var cr CommentIDRequest
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var commentModel models.CommentModel
	err = global.DB.Take(&commentModel, cr.ID).Error
	if err != nil {
		res.FailWithMessage("评论不存在", c)
		return
	}
	redis_ser.NewCommentDigg().Set(fmt.Sprintf("%d", cr.ID))
	res.OkWithMessage("评论点赞成功", c)
	return

}
