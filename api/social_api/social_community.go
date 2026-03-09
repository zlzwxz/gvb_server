package social_api

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"
	"gvb-server/utils/sanitize"
)

type communityPostURI struct {
	ID uint `uri:"id" binding:"required"`
}

type communityListQuery struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Key      string `form:"key"`
	Scene    string `form:"scene"`
	Category string `form:"category"`
	Status   string `form:"status"`
	Sort     string `form:"sort"`
	UserID   uint   `form:"user_id"`
}

type communityCreateRequest struct {
	Scene       string                   `json:"scene"`
	Title       string                   `json:"title" binding:"required"`
	Summary     string                   `json:"summary"`
	Content     string                   `json:"content" binding:"required"`
	Category    string                   `json:"category"`
	Tags        []string                 `json:"tags"`
	CoverImage  string                   `json:"cover_image"`
	Attachments []models.SpaceAttachment `json:"attachments"`
	Budget      int                      `json:"budget"`
	RewardUnit  string                   `json:"reward_unit"`
	Deadline    string                   `json:"deadline"`
}

type communityReplyRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID uint   `json:"parent_id"`
}

type communityStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type communityPinRequest struct {
	IsPinned bool `json:"is_pinned"`
}

type communityMetaCategoryItem struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type communityMetaResponse struct {
	Scene      string                      `json:"scene"`
	Total      int64                       `json:"total"`
	Categories []communityMetaCategoryItem `json:"categories"`
	Status     map[string]int64            `json:"status"`
}

type communityPostListItem struct {
	ID               uint                       `json:"id"`
	Scene            string                     `json:"scene"`
	SceneLabel       string                     `json:"scene_label"`
	UserID           uint                       `json:"user_id"`
	UserNickName     string                     `json:"user_nick_name"`
	UserAvatar       string                     `json:"user_avatar"`
	Title            string                     `json:"title"`
	Summary          string                     `json:"summary"`
	Content          string                     `json:"content"`
	Category         string                     `json:"category"`
	Tags             []string                   `json:"tags"`
	CoverImage       string                     `json:"cover_image"`
	Attachments      models.SpaceAttachmentList `json:"attachments"`
	Status           string                     `json:"status"`
	StatusLabel      string                     `json:"status_label"`
	Budget           int                        `json:"budget"`
	RewardUnit       string                     `json:"reward_unit"`
	Deadline         *time.Time                 `json:"deadline"`
	AcceptedUserID   uint                       `json:"accepted_user_id"`
	AcceptedUserNick string                     `json:"accepted_user_nick"`
	ReplyCount       int                        `json:"reply_count"`
	ViewCount        int                        `json:"view_count"`
	LastReplyAt      *time.Time                 `json:"last_reply_at"`
	LastReplyNick    string                     `json:"last_reply_nick"`
	LastReplyPreview string                     `json:"last_reply_preview"`
	IsPinned         bool                       `json:"is_pinned"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	CanEdit          bool                       `json:"can_edit"`
	CanReply         bool                       `json:"can_reply"`
	CanAccept        bool                       `json:"can_accept"`
	CanManageStatus  bool                       `json:"can_manage_status"`
}

type communityReplyItem struct {
	ID             uint      `json:"id"`
	PostID         uint      `json:"post_id"`
	ParentID       uint      `json:"parent_id"`
	UserID         uint      `json:"user_id"`
	UserNickName   string    `json:"user_nick_name"`
	UserAvatar     string    `json:"user_avatar"`
	Content        string    `json:"content"`
	IsOfficial     bool      `json:"is_official"`
	QuotedUserID   uint      `json:"quoted_user_id"`
	QuotedUserNick string    `json:"quoted_user_nick"`
	CreatedAt      time.Time `json:"created_at"`
	CanDelete      bool      `json:"can_delete"`
}

type communityDetailResponse struct {
	Post    communityPostListItem   `json:"post"`
	Replies []communityReplyItem    `json:"replies"`
	Related []communityPostListItem `json:"related"`
}

func optionalSocialClaims(c *gin.Context) *jwts.CustomClaims {
	token := resolveSocketToken(c)
	if token == "" {
		return nil
	}
	claims, err := jwts.ParseToken(token)
	if err != nil || claims == nil || redis_ser.CheckLogout(token) {
		return nil
	}
	return claims
}

func normalizeCommunityPagination(cr *communityListQuery, defaultLimit int) {
	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = defaultLimit
	}
	if cr.Limit > 50 {
		cr.Limit = 50
	}
}

func normalizeCommunityText(value string, limit int) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r", "\n"), "\t", " "))
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}

func normalizeCommunityCategory(value string) string {
	value = normalizeCommunityText(value, 24)
	if value == "" {
		return "综合交流"
	}
	return value
}

func normalizeCommunityTagList(items []string) ctype.Array {
	result := make([]string, 0, len(items))
	set := map[string]struct{}{}
	for _, item := range items {
		tag := normalizeCommunityText(item, 16)
		tag = strings.Trim(tag, "# ")
		if tag == "" {
			continue
		}
		if _, ok := set[tag]; ok {
			continue
		}
		set[tag] = struct{}{}
		result = append(result, tag)
	}
	return ctype.Array(result)
}

func sanitizeCommunityAttachments(items []models.SpaceAttachment) models.SpaceAttachmentList {
	result := make(models.SpaceAttachmentList, 0, len(items))
	for _, item := range items {
		urlValue := sanitize.CleanURL(item.URL, true)
		if urlValue == "" {
			continue
		}
		name := normalizeCommunityText(item.Name, 60)
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

func buildCommunitySummary(summary string, content string) string {
	summary = normalizeCommunityText(summary, 120)
	if summary != "" {
		return summary
	}
	content = strings.ReplaceAll(content, "\n", " ")
	return normalizeCommunityText(content, 120)
}

func parseCommunityDeadline(raw string) *time.Time {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		value, err := time.ParseInLocation(layout, text, time.Local)
		if err == nil {
			return &value
		}
	}
	return nil
}

func communityStatusLabel(scene string, status string) string {
	scene = models.NormalizeCommunityScene(scene)
	status = models.NormalizeCommunityStatus(scene, status)
	if scene == models.CommunitySceneBounty {
		switch status {
		case models.CommunityStatusOpen:
			return "招募中"
		case models.CommunityStatusInProgress:
			return "进行中"
		case models.CommunityStatusResolved:
			return "已完成"
		case models.CommunityStatusClosed:
			return "已关闭"
		default:
			return "招募中"
		}
	}
	if status == models.CommunityStatusClosed {
		return "已归档"
	}
	return "交流中"
}

func communitySceneLabel(scene string) string {
	if models.NormalizeCommunityScene(scene) == models.CommunitySceneBounty {
		return "赏金"
	}
	return "交流"
}

func communityCanEdit(post models.CommunityPostModel, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	return claims.Role == int(ctype.PermissionAdmin) || claims.UserID == post.UserID
}

func communityCanReply(post models.CommunityPostModel, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	if post.Scene == models.CommunitySceneBounty && post.Status == models.CommunityStatusClosed {
		return claims.Role == int(ctype.PermissionAdmin) || claims.UserID == post.UserID
	}
	return true
}

func communityCanAccept(post models.CommunityPostModel, claims *jwts.CustomClaims) bool {
	if claims == nil || post.Scene != models.CommunitySceneBounty {
		return false
	}
	if claims.UserID == post.UserID || models.NormalizeCommunityStatus(post.Scene, post.Status) != models.CommunityStatusOpen {
		return false
	}
	return true
}

func communityCanManageStatus(post models.CommunityPostModel, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	return claims.Role == int(ctype.PermissionAdmin) || claims.UserID == post.UserID
}

func communityPostToItem(post models.CommunityPostModel, claims *jwts.CustomClaims) communityPostListItem {
	return communityPostListItem{
		ID:               post.ID,
		Scene:            post.Scene,
		SceneLabel:       communitySceneLabel(post.Scene),
		UserID:           post.UserID,
		UserNickName:     post.UserNickName,
		UserAvatar:       post.UserAvatar,
		Title:            post.Title,
		Summary:          post.Summary,
		Content:          post.Content,
		Category:         normalizeCommunityCategory(post.Category),
		Tags:             []string(post.Tags),
		CoverImage:       sanitize.CleanURL(post.CoverImage, true),
		Attachments:      post.Attachments,
		Status:           models.NormalizeCommunityStatus(post.Scene, post.Status),
		StatusLabel:      communityStatusLabel(post.Scene, post.Status),
		Budget:           post.Budget,
		RewardUnit:       models.CommunityRewardUnitLabel(post.RewardUnit),
		Deadline:         post.Deadline,
		AcceptedUserID:   post.AcceptedUserID,
		AcceptedUserNick: post.AcceptedUserNick,
		ReplyCount:       post.ReplyCount,
		ViewCount:        post.ViewCount,
		LastReplyAt:      post.LastReplyAt,
		LastReplyNick:    post.LastReplyNick,
		LastReplyPreview: post.LastReplyPreview,
		IsPinned:         post.IsPinned,
		CreatedAt:        post.CreatedAt,
		UpdatedAt:        post.UpdatedAt,
		CanEdit:          communityCanEdit(post, claims),
		CanReply:         communityCanReply(post, claims),
		CanAccept:        communityCanAccept(post, claims),
		CanManageStatus:  communityCanManageStatus(post, claims),
	}
}

func communityReplyToItem(reply models.CommunityReplyModel, claims *jwts.CustomClaims) communityReplyItem {
	return communityReplyItem{
		ID:             reply.ID,
		PostID:         reply.PostID,
		ParentID:       reply.ParentID,
		UserID:         reply.UserID,
		UserNickName:   reply.UserNickName,
		UserAvatar:     reply.UserAvatar,
		Content:        reply.Content,
		IsOfficial:     reply.IsOfficial,
		QuotedUserID:   reply.QuotedUserID,
		QuotedUserNick: reply.QuotedUserNick,
		CreatedAt:      reply.CreatedAt,
		CanDelete:      claims != nil && (claims.Role == int(ctype.PermissionAdmin) || claims.UserID == reply.UserID),
	}
}

func communityBaseQuery(scene string) *gorm.DB {
	query := global.DB.Model(&models.CommunityPostModel{})
	scene = strings.TrimSpace(strings.ToLower(scene))
	if scene == "" || scene == "all" {
		return query
	}
	return query.Where("scene = ?", models.NormalizeCommunityScene(scene))
}

func applyCommunityFilters(query *gorm.DB, cr communityListQuery) *gorm.DB {
	key := strings.TrimSpace(cr.Key)
	if key != "" {
		like := "%" + key + "%"
		query = query.Where(
			"title LIKE ? OR summary LIKE ? OR content LIKE ? OR category LIKE ? OR user_nick_name LIKE ?",
			like, like, like, like, like,
		)
	}
	if cr.UserID > 0 {
		query = query.Where("user_id = ?", cr.UserID)
	}
	if strings.TrimSpace(cr.Category) != "" {
		query = query.Where("category = ?", normalizeCommunityCategory(cr.Category))
	}
	if strings.TrimSpace(cr.Status) != "" {
		scene := strings.TrimSpace(cr.Scene)
		status := cr.Status
		if scene == "" || strings.EqualFold(scene, "all") {
			status = strings.ToLower(strings.TrimSpace(status))
		} else {
			status = models.NormalizeCommunityStatus(scene, status)
		}
		query = query.Where("status = ?", status)
	}
	return query
}

func applyCommunitySort(query *gorm.DB, scene string, sort string) *gorm.DB {
	scene = models.NormalizeCommunityScene(scene)
	sort = strings.ToLower(strings.TrimSpace(sort))
	switch sort {
	case "latest", "newest":
		return query.Order("is_pinned desc").Order("created_at desc")
	case "hot":
		return query.Order("is_pinned desc").Order("reply_count desc").Order("last_reply_at desc").Order("created_at desc")
	case "reward":
		return query.Order("is_pinned desc").Order("budget desc").Order("created_at desc")
	case "deadline":
		return query.Order("is_pinned desc").Order("deadline asc").Order("created_at desc")
	}
	if scene == models.CommunitySceneBounty {
		return query.Order("is_pinned desc").
			Order("CASE status WHEN 'open' THEN 0 WHEN 'in_progress' THEN 1 WHEN 'resolved' THEN 2 WHEN 'closed' THEN 3 ELSE 4 END").
			Order("budget desc").
			Order("created_at desc")
	}
	return query.Order("is_pinned desc").Order("last_reply_at desc").Order("created_at desc")
}

func loadCommunityRelated(post models.CommunityPostModel, claims *jwts.CustomClaims) []communityPostListItem {
	var list []models.CommunityPostModel
	query := global.DB.Model(&models.CommunityPostModel{}).
		Where("scene = ? AND id <> ?", post.Scene, post.ID)
	if strings.TrimSpace(post.Category) != "" {
		query = query.Where("category = ?", post.Category)
	}
	if err := query.Order("is_pinned desc").Order("created_at desc").Limit(4).Find(&list).Error; err != nil {
		return []communityPostListItem{}
	}
	result := make([]communityPostListItem, 0, len(list))
	for _, item := range list {
		result = append(result, communityPostToItem(item, claims))
	}
	return result
}

func loadCommunityReplies(postID uint, claims *jwts.CustomClaims) []communityReplyItem {
	var replies []models.CommunityReplyModel
	if err := global.DB.Where("post_id = ?", postID).Order("created_at asc").Find(&replies).Error; err != nil {
		return []communityReplyItem{}
	}
	result := make([]communityReplyItem, 0, len(replies))
	for _, item := range replies {
		result = append(result, communityReplyToItem(item, claims))
	}
	return result
}

func (SocialApi) CommunityMetaView(c *gin.Context) {
	var cr communityListQuery
	_ = c.ShouldBindQuery(&cr)
	scene := cr.Scene
	if strings.TrimSpace(scene) == "" {
		scene = models.CommunityScenePlaza
	}

	query := communityBaseQuery(scene)

	var total int64
	_ = query.Count(&total).Error

	statusMap := map[string]int64{}
	var statusRows []struct {
		Status string
		Count  int64
	}
	_ = query.Session(&gorm.Session{}).
		Select("status, COUNT(1) AS count").
		Group("status").
		Scan(&statusRows).Error
	for _, row := range statusRows {
		statusMap[row.Status] = row.Count
	}

	var categoryRows []communityMetaCategoryItem
	_ = query.Session(&gorm.Session{}).
		Select("COALESCE(NULLIF(category, ''), '综合交流') AS category, COUNT(1) AS count").
		Group("COALESCE(NULLIF(category, ''), '综合交流')").
		Order("count desc, category asc").
		Scan(&categoryRows).Error
	for index := range categoryRows {
		categoryRows[index].Category = normalizeCommunityCategory(categoryRows[index].Category)
	}

	res.OkWithData(communityMetaResponse{
		Scene:      models.NormalizeCommunityScene(scene),
		Total:      total,
		Categories: categoryRows,
		Status:     statusMap,
	}, c)
}

func (SocialApi) CommunityPostListView(c *gin.Context) {
	var cr communityListQuery
	_ = c.ShouldBindQuery(&cr)
	normalizeCommunityPagination(&cr, 12)
	if strings.TrimSpace(cr.Scene) == "" {
		cr.Scene = models.CommunityScenePlaza
	}

	claims := optionalSocialClaims(c)
	query := applyCommunityFilters(communityBaseQuery(cr.Scene), cr)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取社区帖子数量失败", c)
		return
	}

	var posts []models.CommunityPostModel
	if err := applyCommunitySort(query, cr.Scene, cr.Sort).
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&posts).Error; err != nil {
		res.FailWithMessage("获取社区帖子失败", c)
		return
	}

	list := make([]communityPostListItem, 0, len(posts))
	for _, item := range posts {
		list = append(list, communityPostToItem(item, claims))
	}
	res.OkWithList(list, count, c)
}

func (SocialApi) CommunityPostDetailView(c *gin.Context) {
	var uri communityPostURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	claims := optionalSocialClaims(c)
	var post models.CommunityPostModel
	if err := global.DB.Take(&post, uri.ID).Error; err != nil {
		res.FailWithMessage("帖子不存在", c)
		return
	}

	_ = global.DB.Model(&post).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
	post.ViewCount++

	res.OkWithData(communityDetailResponse{
		Post:    communityPostToItem(post, claims),
		Replies: loadCommunityReplies(post.ID, claims),
		Related: loadCommunityRelated(post, claims),
	}, c)
}

func (SocialApi) CommunityPostCreateView(c *gin.Context) {
	var cr communityCreateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}

	scene := models.NormalizeCommunityScene(cr.Scene)
	title := normalizeCommunityText(cr.Title, 80)
	content := normalizeCommunityText(cr.Content, 5000)
	if title == "" || content == "" {
		res.FailWithMessage("标题和内容不能为空", c)
		return
	}

	budget := cr.Budget
	if budget < 0 {
		budget = 0
	}
	if scene == models.CommunitySceneBounty && budget <= 0 {
		res.FailWithMessage("赏金帖子请设置大于 0 的赏金金额", c)
		return
	}

	model := models.CommunityPostModel{
		Scene:       scene,
		UserID:      user.ID,
		UserNickName: func() string {
			if strings.TrimSpace(user.NickName) != "" {
				return user.NickName
			}
			return user.UserName
		}(),
		UserAvatar:  user.Avatar,
		Title:       title,
		Summary:     buildCommunitySummary(cr.Summary, content),
		Content:     content,
		Category:    normalizeCommunityCategory(cr.Category),
		Tags:        normalizeCommunityTagList(cr.Tags),
		CoverImage:  sanitize.CleanURL(cr.CoverImage, true),
		Attachments: sanitizeCommunityAttachments(cr.Attachments),
		Status:      models.NormalizeCommunityStatus(scene, ""),
		Budget:      budget,
		RewardUnit:  models.CommunityRewardUnitLabel(cr.RewardUnit),
		Deadline:    parseCommunityDeadline(cr.Deadline),
	}
	if err := global.DB.Create(&model).Error; err != nil {
		res.FailWithMessage("发布失败", c)
		return
	}
	res.OkWithData(communityPostToItem(model, claims), c)
}

func (SocialApi) CommunityPostRemoveView(c *gin.Context) {
	var uri communityPostURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var post models.CommunityPostModel
	if err := global.DB.Take(&post, uri.ID).Error; err != nil {
		res.FailWithMessage("帖子不存在", c)
		return
	}
	if !communityCanEdit(post, claims) {
		res.FailWithMessage("无权删除该帖子", c)
		return
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.CommunityReplyModel{}).Error; err != nil {
			return err
		}
		return tx.Delete(&post).Error
	}); err != nil {
		res.FailWithMessage("删除帖子失败", c)
		return
	}
	res.OkWithMessage("帖子已删除", c)
}

func (SocialApi) CommunityReplyCreateView(c *gin.Context) {
	var uri communityPostURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr communityReplyRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var post models.CommunityPostModel
	if err := global.DB.Take(&post, uri.ID).Error; err != nil {
		res.FailWithMessage("帖子不存在", c)
		return
	}
	if !communityCanReply(post, claims) {
		res.FailWithMessage("当前状态不可回复", c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}

	content := normalizeCommunityText(cr.Content, 2000)
	if content == "" {
		res.FailWithMessage("回复内容不能为空", c)
		return
	}

	reply := models.CommunityReplyModel{
		PostID:       post.ID,
		ParentID:     cr.ParentID,
		UserID:       user.ID,
		UserNickName: func() string {
			if strings.TrimSpace(user.NickName) != "" {
				return user.NickName
			}
			return user.UserName
		}(),
		UserAvatar: user.Avatar,
		Content:    content,
		IsOfficial: claims.Role == int(ctype.PermissionAdmin),
	}
	if cr.ParentID > 0 {
		var parent models.CommunityReplyModel
		if err := global.DB.Take(&parent, "id = ? AND post_id = ?", cr.ParentID, post.ID).Error; err == nil {
			reply.QuotedUserID = parent.UserID
			reply.QuotedUserNick = parent.UserNickName
		}
	}

	now := time.Now()
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&reply).Error; err != nil {
			return err
		}
		return tx.Model(&post).Updates(map[string]any{
			"reply_count":         gorm.Expr("reply_count + ?", 1),
			"last_reply_at":       &now,
			"last_reply_nick":     reply.UserNickName,
			"last_reply_preview":  buildCommunitySummary("", reply.Content),
			"updated_at":          now,
		}).Error
	}); err != nil {
		res.FailWithMessage("发布回复失败", c)
		return
	}
	res.OkWithData(communityReplyToItem(reply, claims), c)
}

func (SocialApi) CommunityBountyAcceptView(c *gin.Context) {
	var uri communityPostURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var post models.CommunityPostModel
	if err := global.DB.Take(&post, uri.ID).Error; err != nil {
		res.FailWithMessage("赏金帖子不存在", c)
		return
	}
	if post.Scene != models.CommunitySceneBounty {
		res.FailWithMessage("只有赏金帖子可接单", c)
		return
	}
	if !communityCanAccept(post, claims) {
		res.FailWithMessage("当前不可接单", c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}
	post.AcceptedUserID = user.ID
	if strings.TrimSpace(user.NickName) != "" {
		post.AcceptedUserNick = user.NickName
	} else {
		post.AcceptedUserNick = user.UserName
	}
	post.Status = models.CommunityStatusInProgress
	if err := global.DB.Model(&post).Updates(map[string]any{
		"accepted_user_id":   post.AcceptedUserID,
		"accepted_user_nick": post.AcceptedUserNick,
		"status":             post.Status,
	}).Error; err != nil {
		res.FailWithMessage("接单失败", c)
		return
	}
	res.OkWithData(communityPostToItem(post, claims), c)
}

func (SocialApi) CommunityPostStatusView(c *gin.Context) {
	var uri communityPostURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr communityStatusRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var post models.CommunityPostModel
	if err := global.DB.Take(&post, uri.ID).Error; err != nil {
		res.FailWithMessage("帖子不存在", c)
		return
	}
	if !communityCanManageStatus(post, claims) {
		res.FailWithMessage("无权修改帖子状态", c)
		return
	}
	post.Status = models.NormalizeCommunityStatus(post.Scene, cr.Status)
	if err := global.DB.Model(&post).Update("status", post.Status).Error; err != nil {
		res.FailWithMessage("更新帖子状态失败", c)
		return
	}
	res.OkWithData(communityPostToItem(post, claims), c)
}

func (SocialApi) CommunityPostPinView(c *gin.Context) {
	var uri communityPostURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr communityPinRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var post models.CommunityPostModel
	if err := global.DB.Take(&post, uri.ID).Error; err != nil {
		res.FailWithMessage("帖子不存在", c)
		return
	}
	post.IsPinned = cr.IsPinned
	if err := global.DB.Model(&post).Update("is_pinned", post.IsPinned).Error; err != nil {
		res.FailWithMessage("置顶状态更新失败", c)
		return
	}
	res.OkWithData(communityPostToItem(post, getClaims(c)), c)
}

func (SocialApi) AdminCommunityPostListView(c *gin.Context) {
	var cr communityListQuery
	_ = c.ShouldBindQuery(&cr)
	normalizeCommunityPagination(&cr, 15)

	query := applyCommunityFilters(communityBaseQuery(cr.Scene), cr)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取社区内容数量失败", c)
		return
	}

	var posts []models.CommunityPostModel
	if err := applyCommunitySort(query, cr.Scene, cr.Sort).
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Find(&posts).Error; err != nil {
		res.FailWithMessage("获取社区内容失败", c)
		return
	}

	claims := getClaims(c)
	list := make([]communityPostListItem, 0, len(posts))
	for _, item := range posts {
		list = append(list, communityPostToItem(item, claims))
	}
	res.OkWithList(list, count, c)
}
