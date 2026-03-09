package article_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
)

var (
	_ = res.Response{}
	_ = models.ArticleReportModel{}
)

// ArticleReportCreateViewDoc 提交文章举报。
// @Tags 文章管理
// @Summary 提交文章举报
// @Description 举报自己的文章会被拦截；同一用户对同一篇文章存在待处理举报时，不能重复提交。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body articleReportCreateRequest true "举报内容"
// @Success 200 {object} res.Response "提交成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles/reports [post]
func ArticleReportCreateViewDoc() {}

// ArticleReportListViewDoc 获取文章举报列表。
// @Tags 文章管理
// @Summary 获取文章举报列表
// @Description 管理员可查看全部举报；版主仅能查看自己管理板块下的举报。status 可按处理状态筛选。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 10，最大 100"
// @Param key query string false "按文章标题、举报人、原因搜索"
// @Param status query int false "举报状态"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.ArticleReportModel}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles/reports [get]
func ArticleReportListViewDoc() {}

// ArticleReportHandleViewDoc 处理文章举报。
// @Tags 文章管理
// @Summary 处理文章举报
// @Description status=2 表示转入文章复审，status=3 表示忽略举报。只有管理员或对应板块版主可处理。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body articleReportHandleRequest true "处理结果"
// @Success 200 {object} res.Response "处理成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles/reports [put]
func ArticleReportHandleViewDoc() {}

// ArticleReviewViewDoc 审核文章。
// @Tags 文章管理
// @Summary 审核文章
// @Description review_status 仅允许传通过或驳回。管理员可全局审核，版主可审核自己管理板块下的文章。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body ArticleReviewRequest true "审核参数"
// @Success 200 {object} res.Response "审核成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/articles/review [put]
func ArticleReviewViewDoc() {}
