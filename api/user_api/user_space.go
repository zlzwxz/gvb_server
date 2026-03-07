package user_api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/board_ser"
	"gvb-server/service/es_ser"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"
	"gvb-server/utils/sanitize"
)

type userIDUri struct {
	ID uint `uri:"id" binding:"required"`
}

type spaceItemIDUri struct {
	ID uint `uri:"id" binding:"required"`
}

type userSpaceListQuery struct {
	models.PageInfo
}

type userSpacePostRequest struct {
	Content     string                   `json:"content" binding:"required"`
	Attachments []models.SpaceAttachment `json:"attachments"`
	IsPrivate   bool                     `json:"is_private"`
}

type userSpaceMessageRequest struct {
	SpaceUserID uint   `json:"space_user_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	IsPrivate   bool   `json:"is_private"`
}

type userSpaceProfileResponse struct {
	ID             uint     `json:"id"`
	NickName       string   `json:"nick_name"`
	UserName       string   `json:"user_name"`
	Avatar         string   `json:"avatar"`
	Sign           string   `json:"sign"`
	Link           string   `json:"link"`
	Role           string   `json:"role"`
	Level          int      `json:"level"`
	Points         int      `json:"points"`
	Experience     int      `json:"experience"`
	CreatedAt      string   `json:"created_at"`
	IsSelf         bool     `json:"is_self"`
	CanManageSpace bool     `json:"can_manage_space"`
	ManagedBoards  []string `json:"managed_boards"`
	Stats          struct {
		PostCount      int64 `json:"post_count"`
		GuestbookCount int64 `json:"guestbook_count"`
		ArticleCount   int   `json:"article_count"`
	} `json:"stats"`
}

func optionalUserClaims(c *gin.Context) *jwts.CustomClaims {
	if claimsAny, ok := c.Get("claims"); ok {
		if claims, ok := claimsAny.(*jwts.CustomClaims); ok {
			return claims
		}
	}

	token := strings.TrimSpace(c.GetHeader("token"))
	if token == "" {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}
	if token == "" {
		return nil
	}
	claims, err := jwts.ParseToken(token)
	if err != nil || redis_ser.CheckLogout(token) {
		return nil
	}
	return claims
}

func canManageUserSpace(spaceUserID uint, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	return claims.Role == int(ctype.PermissionAdmin) || claims.UserID == spaceUserID
}

func normalizeSpacePagination(cr *userSpaceListQuery) {
	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 10
	}
	if cr.Limit > 50 {
		cr.Limit = 50
	}
}

func sanitizeSpaceAttachments(items []models.SpaceAttachment) []models.SpaceAttachment {
	result := make([]models.SpaceAttachment, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		urlValue := sanitize.CleanURL(item.URL, true)
		if urlValue == "" {
			continue
		}
		if name == "" {
			name = "附件"
		}
		result = append(result, models.SpaceAttachment{
			FileID: item.FileID,
			Name:   strings.ReplaceAll(strings.ReplaceAll(name, "<", ""), ">", ""),
			URL:    urlValue,
			Size:   item.Size,
		})
	}
	return result
}

func listManagedBoardNames(userID uint) []string {
	if userID == 0 {
		return []string{}
	}
	var boards []models.BoardModel
	if err := global.DB.Order("sort asc").Order("id asc").Find(&boards).Error; err != nil {
		return []string{}
	}
	list := make([]string, 0)
	for _, board := range boards {
		if board_ser.IsUserBoardManager(board, userID) {
			list = append(list, board.Name)
		}
	}
	return list
}

// UserPublicProfileView 获取公开用户资料和空间统计。
func (UserApi) UserPublicProfileView(c *gin.Context) {
	var uri userIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, uri.ID).Error; err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}

	claims := optionalUserClaims(c)
	canManage := canManageUserSpace(user.ID, claims)

	var postCount int64
	postQuery := global.DB.Model(&models.UserSpacePostModel{}).Where("user_id = ?", user.ID)
	if !canManage {
		postQuery = postQuery.Where("is_private = ?", false)
	}
	_ = postQuery.Count(&postCount).Error

	var messageCount int64
	messageQuery := global.DB.Model(&models.UserSpaceMessageModel{}).Where("space_user_id = ?", user.ID)
	if !canManage {
		if claims != nil && claims.UserID > 0 {
			messageQuery = messageQuery.Where("is_private = ? OR user_id = ?", false, claims.UserID)
		} else {
			messageQuery = messageQuery.Where("is_private = ?", false)
		}
	}
	_ = messageQuery.Count(&messageCount).Error

	articleList, articleCount, err := es_ser.CommList(es_ser.Option{
		PageInfo: models.PageInfo{Page: 1, Limit: 1},
		Query:    buildUserArticleQuery(user.ID, canManage),
	})
	if err != nil {
		articleCount = 0
	}
	_ = articleList

	response := userSpaceProfileResponse{
		ID:             user.ID,
		NickName:       user.NickName,
		UserName:       user.UserName,
		Avatar:         user.Avatar,
		Sign:           user.Sign,
		Link:           user.Link,
		Role:           user.Role.String(),
		Level:          user.Level,
		Points:         user.Points,
		Experience:     user.Experience,
		CreatedAt:      user.CreatedAt.Format("2006-01-02 15:04:05"),
		IsSelf:         claims != nil && claims.UserID == user.ID,
		CanManageSpace: canManage,
		ManagedBoards:  listManagedBoardNames(user.ID),
	}
	response.Stats.PostCount = postCount
	response.Stats.GuestbookCount = messageCount
	response.Stats.ArticleCount = articleCount
	res.OkWithData(response, c)
}

// UserSpacePostListView 获取用户空间动态列表。
func (UserApi) UserSpacePostListView(c *gin.Context) {
	var uri userIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr userSpaceListQuery
	_ = c.ShouldBindQuery(&cr)
	normalizeSpacePagination(&cr)

	claims := optionalUserClaims(c)
	canManage := canManageUserSpace(uri.ID, claims)

	query := global.DB.Model(&models.UserSpacePostModel{}).Where("user_id = ?", uri.ID)
	if !canManage {
		query = query.Where("is_private = ?", false)
	}
	if strings.TrimSpace(cr.Key) != "" {
		query = query.Where("content LIKE ?", "%"+strings.TrimSpace(cr.Key)+"%")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取空间动态数量失败", c)
		return
	}

	var list []models.UserSpacePostModel
	if err := query.Order("created_at desc").Limit(cr.Limit).Offset((cr.Page - 1) * cr.Limit).Find(&list).Error; err != nil {
		res.FailWithMessage("获取空间动态失败", c)
		return
	}
	res.OkWithList(list, count, c)
}

// UserSpacePostCreateView 发表空间动态。
func (UserApi) UserSpacePostCreateView(c *gin.Context) {
	var cr userSpacePostRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	content := strings.TrimSpace(cr.Content)
	if content == "" {
		res.FailWithMessage("动态内容不能为空", c)
		return
	}
	if len([]rune(content)) > 5000 {
		res.FailWithMessage("动态内容不能超过 5000 个字符", c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}

	model := models.UserSpacePostModel{
		UserID:       user.ID,
		UserNickName: user.NickName,
		UserAvatar:   user.Avatar,
		Content:      content,
		Attachments:  models.SpaceAttachmentList(sanitizeSpaceAttachments(cr.Attachments)),
		IsPrivate:    cr.IsPrivate,
	}
	if err := global.DB.Create(&model).Error; err != nil {
		res.FailWithMessage("发布动态失败", c)
		return
	}
	res.OkWithData(model, c)
}

// UserSpacePostRemoveView 删除空间动态。
func (UserApi) UserSpacePostRemoveView(c *gin.Context) {
	var uri spaceItemIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	var model models.UserSpacePostModel
	if err := global.DB.Take(&model, uri.ID).Error; err != nil {
		res.FailWithMessage("动态不存在", c)
		return
	}
	if !canManageUserSpace(model.UserID, claims) {
		res.FailWithMessage("无权删除该动态", c)
		return
	}
	if err := global.DB.Delete(&model).Error; err != nil {
		res.FailWithMessage("删除动态失败", c)
		return
	}
	res.OkWithMessage("动态删除成功", c)
}

// UserSpaceMessageListView 获取空间留言列表。
func (UserApi) UserSpaceMessageListView(c *gin.Context) {
	var uri userIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr userSpaceListQuery
	_ = c.ShouldBindQuery(&cr)
	normalizeSpacePagination(&cr)

	claims := optionalUserClaims(c)
	canManage := canManageUserSpace(uri.ID, claims)

	query := global.DB.Model(&models.UserSpaceMessageModel{}).Where("space_user_id = ?", uri.ID)
	if !canManage {
		if claims != nil && claims.UserID > 0 {
			query = query.Where("is_private = ? OR user_id = ?", false, claims.UserID)
		} else {
			query = query.Where("is_private = ?", false)
		}
	}
	if strings.TrimSpace(cr.Key) != "" {
		query = query.Where("content LIKE ?", "%"+strings.TrimSpace(cr.Key)+"%")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取空间留言数量失败", c)
		return
	}

	var list []models.UserSpaceMessageModel
	if err := query.Order("created_at desc").Limit(cr.Limit).Offset((cr.Page - 1) * cr.Limit).Find(&list).Error; err != nil {
		res.FailWithMessage("获取空间留言失败", c)
		return
	}
	res.OkWithList(list, count, c)
}

// UserSpaceMessageCreateView 发表空间留言。
func (UserApi) UserSpaceMessageCreateView(c *gin.Context) {
	var cr userSpaceMessageRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	content := strings.TrimSpace(cr.Content)
	if content == "" {
		res.FailWithMessage("留言内容不能为空", c)
		return
	}
	if len([]rune(content)) > 1000 {
		res.FailWithMessage("留言内容不能超过 1000 个字符", c)
		return
	}
	if cr.SpaceUserID == 0 {
		res.FailWithMessage("空间用户不能为空", c)
		return
	}

	var owner models.UserModel
	if err := global.DB.Take(&owner, cr.SpaceUserID).Error; err != nil {
		res.FailWithMessage("空间用户不存在", c)
		return
	}
	var sender models.UserModel
	if err := global.DB.Take(&sender, claims.UserID).Error; err != nil {
		res.FailWithMessage("留言用户不存在", c)
		return
	}

	model := models.UserSpaceMessageModel{
		SpaceUserID:  owner.ID,
		UserID:       sender.ID,
		UserNickName: sender.NickName,
		UserAvatar:   sender.Avatar,
		Content:      content,
		IsPrivate:    cr.IsPrivate,
	}
	if err := global.DB.Create(&model).Error; err != nil {
		res.FailWithMessage("发表留言失败", c)
		return
	}
	res.OkWithData(model, c)
}

// UserSpaceMessageRemoveView 删除空间留言。
func (UserApi) UserSpaceMessageRemoveView(c *gin.Context) {
	var uri spaceItemIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claimsAny, _ := c.Get("claims")
	claims := claimsAny.(*jwts.CustomClaims)

	var model models.UserSpaceMessageModel
	if err := global.DB.Take(&model, uri.ID).Error; err != nil {
		res.FailWithMessage("留言不存在", c)
		return
	}
	if !(claims.Role == int(ctype.PermissionAdmin) || claims.UserID == model.SpaceUserID || claims.UserID == model.UserID) {
		res.FailWithMessage("无权删除该留言", c)
		return
	}
	if err := global.DB.Delete(&model).Error; err != nil {
		res.FailWithMessage("删除留言失败", c)
		return
	}
	res.OkWithMessage("留言删除成功", c)
}

func buildUserArticleQuery(userID uint, canManage bool) elastic.Query {
	boolQuery := elastic.NewBoolQuery().Must(elastic.NewTermQuery("user_id", userID))
	if !canManage {
		boolQuery.Must(publicSpaceVisibleArticleQuery())
	}
	return boolQuery
}

func publicSpaceVisibleArticleQuery() elastic.Query {
	reviewQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery("review_status", int(ctype.ArticleReviewApproved)),
			elastic.NewTermQuery("review_status", int(ctype.ArticleReviewLegacy)),
			elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("review_status")),
		).
		MinimumNumberShouldMatch(1)
	privateQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery("is_private", false),
			elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("is_private")),
		).
		MinimumNumberShouldMatch(1)
	return elastic.NewBoolQuery().Must(reviewQuery, privateQuery)
}
