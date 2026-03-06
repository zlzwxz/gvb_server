package article_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
	"github.com/olivere/elastic/v7"
)

// ArticleSearchRequest 文章搜索请求参数
type ArticleSearchRequest struct {
	models.PageInfo
	Tag          string `json:"tag" form:"tag" swag:"description:标签"`
	IsUser       bool   `json:"is_user" form:"is_user" swag:"description:是否只显示当前用户的文章"` // 根据这个参数判断是否显示我收藏的文章列表
	ReviewStatus int    `json:"review_status" form:"review_status" swag:"description:审核状态筛选"`
	ReviewScope  string `json:"review_scope" form:"review_scope" swag:"description:审核视角，all代表管理员查看全量"`
	Category     string `json:"category" form:"category" swag:"description:分类筛选"`
}

// ArticleListView 获取文章列表
// @Summary 获取文章列表
// @Description 获取文章列表，支持分页、标签筛选和用户文章筛选
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认10"
// @Param sort query string false "排序方式"
// @Param tag query string false "标签"
// @Param is_user query bool false "是否只显示当前用户的文章"
// @Param token header string false "token，当is_user为true时必填"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.ArticleModel}} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles [get]
func (ArticleApi) ArticleListView(c *gin.Context) {
	var cr ArticleSearchRequest
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	boolSearch := elastic.NewBoolQuery()
	claims := optionalClaims(c)
	isAdminUser := isAdmin(claims)
	if cr.IsUser {
		if claims == nil {
			res.FailWithMessage("请先登录", c)
			return
		}
		boolSearch.Must(elastic.NewTermQuery("user_id", claims.UserID))
	} else {
		// 非“我的文章”列表默认只展示已通过文章；管理员可显式查看全量
		if !(isAdminUser && strings.EqualFold(cr.ReviewScope, "all")) {
			boolSearch.Must(publicVisibleQuery())
		}
	}

	if cr.ReviewStatus != 0 {
		if isAdminUser || cr.IsUser {
			boolSearch.Must(elastic.NewTermQuery("review_status", cr.ReviewStatus))
		}
	}
	if cr.Category != "" {
		boolSearch.Must(elastic.NewTermQuery("category", cr.Category))
	}

	list, count, err := es_ser.CommList(es_ser.Option{
		PageInfo: cr.PageInfo,
		Fields:   []string{"title", "content", "category"},
		Tag:      cr.Tag,
		Query:    boolSearch,
	})
	if err != nil {
		global.Log.Error(err)
		res.OkWithMessage("查询失败", c)
		return
	}

	// json-filter空值问题
	data := filter.Omit("list", list)
	_list, _ := data.(filter.Filter)
	if string(_list.MustMarshalJSON()) == "{}" {
		list = make([]models.ArticleModel, 0)
		res.OkWithList(list, int64(count), c)
		return
	}
	res.OkWithList(data, int64(count), c)
}
