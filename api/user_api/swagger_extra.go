package user_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
)

var (
	_ = res.Response{}
	_ = models.UserSpacePostModel{}
)

// UserCheckInStatusViewDoc 获取签到状态。
// @Tags 用户管理
// @Summary 获取签到状态
// @Description 返回当前用户今天是否已签到、当前积分、经验、等级和连续签到天数。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=UserCheckInStatusResponse} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/user_check_in_status [get]
func UserCheckInStatusViewDoc() {}

// UserCheckInViewDoc 用户签到。
// @Tags 用户管理
// @Summary 用户签到
// @Description 每日仅允许签到一次。游客账号不可签到；连续签到会获得额外积分和经验奖励。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=UserCheckInResponse} "签到成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/user_check_in [post]
func UserCheckInViewDoc() {}

// UserPublicProfileViewDoc 获取公开用户资料。
// @Tags 用户管理
// @Summary 获取用户空间资料
// @Description token 可选。未登录或访问他人空间时只会看到公开统计；本人或管理员可看到更多可管理状态。
// @Accept json
// @Produce json
// @Param token header string false "token，可选"
// @Param id path int true "用户 ID"
// @Success 200 {object} res.Response{data=userSpaceProfileResponse} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Router /api/users/{id}/profile [get]
func UserPublicProfileViewDoc() {}

// UserSpacePostListViewDoc 获取空间动态列表。
// @Tags 用户管理
// @Summary 获取空间动态列表
// @Description token 可选。访问他人空间时默认只返回公开动态；本人或管理员可查看私密动态。
// @Accept json
// @Produce json
// @Param token header string false "token，可选"
// @Param id path int true "空间用户 ID"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 10，最大 50"
// @Param key query string false "按动态内容搜索"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.UserSpacePostModel}} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Router /api/users/{id}/space/posts [get]
func UserSpacePostListViewDoc() {}

// UserSpacePostCreateViewDoc 发布空间动态。
// @Tags 用户管理
// @Summary 发布空间动态
// @Description 支持附带附件链接和是否私密。私密动态仅本人和管理员可见。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body userSpacePostRequest true "动态内容"
// @Success 200 {object} res.Response{data=models.UserSpacePostModel} "发布成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/space/posts [post]
func UserSpacePostCreateViewDoc() {}

// UserSpacePostRemoveViewDoc 删除空间动态。
// @Tags 用户管理
// @Summary 删除空间动态
// @Description 只有空间主人或管理员可以删除空间动态。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "动态 ID"
// @Success 200 {object} res.Response "删除成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/space/posts/{id} [delete]
func UserSpacePostRemoveViewDoc() {}

// UserSpaceMessageListViewDoc 获取空间留言列表。
// @Tags 用户管理
// @Summary 获取空间留言列表
// @Description token 可选。未登录时只能看到公开留言；登录后会额外看到自己留下的私密留言；空间主人和管理员可看全部。
// @Accept json
// @Produce json
// @Param token header string false "token，可选"
// @Param id path int true "空间用户 ID"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，默认 10，最大 50"
// @Param key query string false "按留言内容搜索"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.UserSpaceMessageModel}} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Router /api/users/{id}/space/messages [get]
func UserSpaceMessageListViewDoc() {}

// UserSpaceMessageCreateViewDoc 发表空间留言。
// @Tags 用户管理
// @Summary 发表空间留言
// @Description 可指定留言给哪个空间用户，并可选择是否私密。私密留言默认仅空间主人、留言人和管理员可见。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body userSpaceMessageRequest true "留言内容"
// @Success 200 {object} res.Response{data=models.UserSpaceMessageModel} "发表成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/space/messages [post]
func UserSpaceMessageCreateViewDoc() {}

// UserSpaceMessageRemoveViewDoc 删除空间留言。
// @Tags 用户管理
// @Summary 删除空间留言
// @Description 管理员、空间主人和留言作者都可以删除该留言。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "留言 ID"
// @Success 200 {object} res.Response "删除成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/space/messages/{id} [delete]
func UserSpaceMessageRemoveViewDoc() {}
