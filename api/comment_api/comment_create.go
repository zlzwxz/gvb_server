package comment_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CommentRequest 评论创建请求参数
type CommentRequest struct {
	ArticleID       string `json:"article_id" binding:"required" msg:"请选择文章" swag:"description:文章ID"`
	Content         string `json:"content" binding:"required" msg:"请输入评论内容" swag:"description:评论内容"`
	ParentCommentID *uint  `json:"parent_comment_id" swag:"description:父评论ID，为空则创建根评论"` // 父评论id
}

// CommentCreateView 创建评论
// @Summary 创建评论
// @Description 创建文章评论，支持创建根评论和子评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body CommentRequest true "评论信息"
// @Success 200 {object} res.Response{msg=string} "创建成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "文章或父评论不存在"
// @Router /api/comments [post]
func (CommentApi) CommentCreateView(c *gin.Context) {
	var cr CommentRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	// 文章是否存在
	_, err = es_ser.CommDetail(cr.ArticleID)
	if err != nil {
		res.FailWithMessage("文章不存在", c)
		return
	}

	// 判断是否是子评论
	if cr.ParentCommentID != nil {
		// 子评论
		// 给父评论数 +1
		// 父评论id
		var parentComment models.CommentModel
		// 找父评论
		err = global.DB.Take(&parentComment, cr.ParentCommentID).Error
		if err != nil {
			res.FailWithMessage("父评论不存在", c)
			return
		}
		// 判断父评论的文章是否和当前文章一致
		if parentComment.ArticleID != cr.ArticleID {
			res.FailWithMessage("评论文章不一致", c)
			return
		}
		// 给父评论评论数+1
		global.DB.Model(&parentComment).Update("comment_count", gorm.Expr("comment_count + 1"))
	}
	// 添加评论
	err = global.DB.Create(&models.CommentModel{
		ParentCommentID: cr.ParentCommentID,
		Content:         cr.Content,
		ArticleID:       cr.ArticleID,
		UserID:          claims.UserID,
	}).Error
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	// 拿到文章数，新的文章评论数存缓存里
	//newCommentCount := article.CommentCount + 1
	// 给文章评论数 +1
	redis_ser.NewCommentCount().Set(cr.ArticleID)
	res.OkWithMessage("文章评论成功", c)
	return
}
