package comment_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"
	"gvb-server/utils"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CommentRemoveView 删除评论
// @Summary 删除评论
// @Description 删除指定评论及其子评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path uint true "评论ID"
// @Success 200 {object} res.Response{msg=string} "删除成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "评论不存在"
// @Router /api/comments/{id} [delete]
func (CommentApi) CommentRemoveView(c *gin.Context) {
	claimsAny, ok := c.Get("claims")
	if !ok {
		res.FailWithMessage("未登录", c)
		return
	}
	claims := claimsAny.(*jwts.CustomClaims)

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
	if claims.Role != int(ctype.PermissionAdmin) && commentModel.UserID != claims.UserID {
		res.FailWithMessage("无权删除该评论", c)
		return
	}
	// 统计评论下的子评论数 再把自己算上去
	subCommentList := FindSubCommentCount(commentModel)
	count := len(subCommentList) + 1
	//减去评论的评论数量
	redis_ser.NewCommentCount().SetCount(commentModel.ArticleID, -count)
	// 判断是否是子评论
	if commentModel.ParentCommentID != nil {
		// 子评论
		// 找父评论，减掉对应的评论数
		global.DB.Model(&models.CommentModel{}).
			Where("id = ?", *commentModel.ParentCommentID).
			Update("comment_count", gorm.Expr("comment_count - ?", count))
	}

	// 删除子评论以及当前评论
	var deleteCommentIDList []uint
	for _, model := range subCommentList {
		deleteCommentIDList = append(deleteCommentIDList, model.ID)
	}
	// 反转，然后一个一个删
	utils.Reverse(deleteCommentIDList)
	deleteCommentIDList = append(deleteCommentIDList, commentModel.ID)
	for _, id := range deleteCommentIDList {
		global.DB.Model(models.CommentModel{}).Delete("id = ?", id)
	}
	res.OkWithMessage(fmt.Sprintf("共删除 %d 条评论", len(deleteCommentIDList)), c)
	return
}
