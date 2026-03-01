package comment_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"
	"strconv"
)

// CommentListRequest 评论列表请求参数
type CommentListRequest struct {
	ArticleID string `form:"article_id" swag:"description:文章ID"`
}

// CommentListView 获取评论列表
// @Summary 获取评论列表
// @Description 获取指定文章的评论列表，包括嵌套的子评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param article_id query string true "文章ID"
// @Success 200 {object} res.Response{data=[]models.CommentModel} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/comments [get]
func (CommentApi) CommentListView(c *gin.Context) {
	var cr CommentListRequest
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	rootCommentList := FindArticleCommentList(cr.ArticleID)
	res.OkWithData(filter.Select("c", rootCommentList), c)
	return
}

func FindArticleCommentList(articleID string) (RootCommentList []*models.CommentModel) {
	// 先把文章下的根评论查出来
	global.DB.Preload("User").Find(&RootCommentList, "article_id = ? and parent_comment_id is null", articleID)

	// 遍历根评论，递归查根评论下的所有子评论
	for _, model := range RootCommentList {
		fmt.Println(model)
		var subCommentList []models.CommentModel
		FindSubComment(*model, &subCommentList)
		model.SubComments = subCommentList
		//评论点赞数从redis获取并且赋值给model
		model.DiggCount += redis_ser.NewCommentDigg().Get(strconv.Itoa(int(model.ID)))
	}
	return
}

// FindSubComment 递归查评论下的子评论
func FindSubComment(model models.CommentModel, subCommentList *[]models.CommentModel) {
	global.DB.Preload("SubComments.User").Take(&model)
	fmt.Println(model)
	for _, sub := range model.SubComments {
		//评论点赞数从redis获取并且赋值给model
		sub.DiggCount = redis_ser.NewCommentDigg().Get(strconv.Itoa(int(model.ID)))
		*subCommentList = append(*subCommentList, sub)
		FindSubComment(sub, subCommentList)
	}
	return
}

// FindSubCommentCount 查询评论里面的所有评论
func FindSubCommentCount(model models.CommentModel) (subCommentList []models.CommentModel) {
	findSubCommentList(model, &subCommentList)
	return subCommentList
}

// findSubCommentList 查询评论里面的所有评论
func findSubCommentList(model models.CommentModel, subCommentList *[]models.CommentModel) {
	global.DB.Preload("SubComments").Take(&model)
	for _, sub := range model.SubComments {
		*subCommentList = append(*subCommentList, sub)
		FindSubComment(sub, subCommentList)
	}
}
