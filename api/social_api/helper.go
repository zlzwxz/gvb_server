package social_api

import (
	"fmt"
	"mime"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

const (
	maxGroupMembers     = 30
	maxStatusTextLength = 60
	maxMessageLength    = 2000
)

type socialUserURI struct {
	ID uint `uri:"id" binding:"required"`
}

type socialFileURI struct {
	ID uint `uri:"id" binding:"required"`
}

type socialGroupURI struct {
	ID uint `uri:"id" binding:"required"`
}

type socialDirectMessageQuery struct {
	UserID uint `form:"user_id" binding:"required"`
}

type socialPresenceRequest struct {
	Mode        string `json:"mode"`
	StatusText  string `json:"status_text"`
	IsInvisible bool   `json:"is_invisible"`
}

type socialFollowCard struct {
	UserID             uint       `json:"user_id"`
	BlogNo             uint       `json:"blog_no"`
	NickName           string     `json:"nick_name"`
	UserName           string     `json:"user_name"`
	Avatar             string     `json:"avatar"`
	IsFollowing        bool       `json:"is_following"`
	FollowsMe          bool       `json:"follows_me"`
	IsFriend           bool       `json:"is_friend"`
	IsBlocked          bool       `json:"is_blocked"`
	BlockedByThem      bool       `json:"blocked_by_them"`
	CanDirectMessage   bool       `json:"can_direct_message"`
	CanCreateGroup     bool       `json:"can_create_group"`
	CanSendFile        bool       `json:"can_send_file"`
	CanCall            bool       `json:"can_call"`
	IsOnline           bool       `json:"is_online"`
	PresenceMode       string     `json:"presence_mode"`
	PresenceText       string     `json:"presence_text"`
	IsInvisible        bool       `json:"is_invisible"`
	LastActiveAt       *time.Time `json:"last_active_at"`
	LastMessageAt      *time.Time `json:"last_message_at,omitempty"`
	LastMessagePreview string     `json:"last_message_preview,omitempty"`
	UnreadCount        int        `json:"unread_count"`
}

type socialConversationItem struct {
	ConversationKey   string     `json:"conversation_key"`
	ConversationType  string     `json:"conversation_type"`
	UserID            uint       `json:"user_id"`
	GroupID           uint       `json:"group_id"`
	Title             string     `json:"title"`
	Avatar            string     `json:"avatar"`
	LatestMessage     string     `json:"latest_message"`
	LatestMessageType string     `json:"latest_message_type"`
	LatestMessageID   uint       `json:"latest_message_id"`
	LatestAt          time.Time  `json:"latest_at"`
	UnreadCount       int        `json:"unread_count"`
	IsFriend          bool       `json:"is_friend"`
	IsOnline          bool       `json:"is_online"`
	PresenceMode      string     `json:"presence_mode"`
	PresenceText      string     `json:"presence_text"`
	LastActiveAt      *time.Time `json:"last_active_at"`
	MemberCount       int        `json:"member_count"`
}

type socialSummaryResponse struct {
	Presence          socialPresenceResponse `json:"presence"`
	FriendCount       int                    `json:"friend_count"`
	OnlineFriendCount int                    `json:"online_friend_count"`
	BlockCount        int                    `json:"block_count"`
}

type socialPresenceResponse struct {
	Mode         string     `json:"mode"`
	StatusText   string     `json:"status_text"`
	IsInvisible  bool       `json:"is_invisible"`
	IsOnline     bool       `json:"is_online"`
	LastActiveAt *time.Time `json:"last_active_at"`
}

type socialRelationResponse struct {
	UserID           uint `json:"user_id"`
	IsFollowing      bool `json:"is_following"`
	FollowsMe        bool `json:"follows_me"`
	IsFriend         bool `json:"is_friend"`
	IsBlocked        bool `json:"is_blocked"`
	BlockedByThem    bool `json:"blocked_by_them"`
	CanDirectMessage bool `json:"can_direct_message"`
	CanCreateGroup   bool `json:"can_create_group"`
	CanSendFile      bool `json:"can_send_file"`
	CanCall          bool `json:"can_call"`
}

type socialFileUploadItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
	Mime string `json:"mime"`
}

func getClaims(c *gin.Context) *jwts.CustomClaims {
	claimsAny, _ := c.Get("claims")
	return claimsAny.(*jwts.CustomClaims)
}

func resolveSocketToken(c *gin.Context) string {
	token := strings.TrimSpace(c.Query("token"))
	if token != "" {
		return token
	}
	token = strings.TrimSpace(c.Request.Header.Get("token"))
	if token != "" {
		return token
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}

func authenticateSocket(c *gin.Context) (*jwts.CustomClaims, error) {
	token := resolveSocketToken(c)
	if token == "" {
		return nil, fmt.Errorf("未携带 token")
	}
	claims, err := jwts.ParseToken(token)
	if err != nil {
		return nil, err
	}
	if redis_ser.CheckLogout(token) {
		return nil, fmt.Errorf("token 已失效")
	}
	return claims, nil
}

func normalizePresenceMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "online", "busy", "away", "vacation", "travel", "invisible":
		return strings.TrimSpace(strings.ToLower(mode))
	default:
		return "online"
	}
}

func buildDirectConversationKey(a, b uint) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("direct:%d:%d", a, b)
}

func buildGroupConversationKey(groupID uint) string {
	return fmt.Sprintf("group:%d", groupID)
}

func messagePreview(item models.SocialMessageModel) string {
	if item.IsRecalled {
		return "消息已撤回"
	}
	switch models.SocialMessageType(item.MsgType) {
	case models.SocialMessageFile:
		if strings.TrimSpace(item.FileName) != "" {
			return "发送了文件：" + strings.TrimSpace(item.FileName)
		}
		return "发送了一个文件"
	case models.SocialMessageCall:
		if strings.TrimSpace(item.Content) != "" {
			return strings.TrimSpace(item.Content)
		}
		return "发起了语音通话"
	default:
		text := strings.TrimSpace(item.Content)
		if text == "" {
			return "新消息"
		}
		return text
	}
}

func dedupeUintSlice(ids []uint) []uint {
	set := map[uint]struct{}{}
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := set[id]; ok {
			continue
		}
		set[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func loadUserMap(ids []uint) map[uint]models.UserModel {
	result := map[uint]models.UserModel{}
	ids = dedupeUintSlice(ids)
	if len(ids) == 0 {
		return result
	}
	var list []models.UserModel
	if err := global.DB.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return result
	}
	for _, item := range list {
		result[item.ID] = item
	}
	return result
}

func fetchFollowingMap(userID uint) map[uint]struct{} {
	result := map[uint]struct{}{}
	var list []models.UserFollowModel
	global.DB.Select("follow_user_id").Where("user_id = ?", userID).Find(&list)
	for _, item := range list {
		result[item.FollowUserID] = struct{}{}
	}
	return result
}

func fetchFollowerMap(userID uint) map[uint]struct{} {
	result := map[uint]struct{}{}
	var list []models.UserFollowModel
	global.DB.Select("user_id").Where("follow_user_id = ?", userID).Find(&list)
	for _, item := range list {
		result[item.UserID] = struct{}{}
	}
	return result
}

func fetchBlockMap(userID uint) map[uint]struct{} {
	result := map[uint]struct{}{}
	var list []models.UserBlockModel
	global.DB.Select("block_user_id").Where("user_id = ?", userID).Find(&list)
	for _, item := range list {
		result[item.BlockUserID] = struct{}{}
	}
	return result
}

func fetchBlockedByMap(userID uint) map[uint]struct{} {
	result := map[uint]struct{}{}
	var list []models.UserBlockModel
	global.DB.Select("user_id").Where("block_user_id = ?", userID).Find(&list)
	for _, item := range list {
		result[item.UserID] = struct{}{}
	}
	return result
}

func fetchFriendIDs(userID uint) []uint {
	following := fetchFollowingMap(userID)
	followers := fetchFollowerMap(userID)
	blocks := fetchBlockMap(userID)
	blockedBy := fetchBlockedByMap(userID)
	result := make([]uint, 0)
	for targetID := range following {
		if _, ok := followers[targetID]; !ok {
			continue
		}
		if _, ok := blocks[targetID]; ok {
			continue
		}
		if _, ok := blockedBy[targetID]; ok {
			continue
		}
		result = append(result, targetID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func isFollowing(userID, targetID uint) bool {
	var count int64
	global.DB.Model(&models.UserFollowModel{}).Where("user_id = ? AND follow_user_id = ?", userID, targetID).Count(&count)
	return count > 0
}

func hasBlockBetween(userID, targetID uint) (blocked, blockedBy bool) {
	var list []models.UserBlockModel
	global.DB.Where("(user_id = ? AND block_user_id = ?) OR (user_id = ? AND block_user_id = ?)", userID, targetID, targetID, userID).Find(&list)
	for _, item := range list {
		if item.UserID == userID {
			blocked = true
		}
		if item.UserID == targetID {
			blockedBy = true
		}
	}
	return
}

func buildRelation(viewerID, targetID uint) socialRelationResponse {
	response := socialRelationResponse{UserID: targetID}
	if viewerID == 0 || targetID == 0 || viewerID == targetID {
		return response
	}
	response.IsFollowing = isFollowing(viewerID, targetID)
	response.FollowsMe = isFollowing(targetID, viewerID)
	response.IsBlocked, response.BlockedByThem = hasBlockBetween(viewerID, targetID)
	response.IsFriend = response.IsFollowing && response.FollowsMe && !response.IsBlocked && !response.BlockedByThem
	response.CanDirectMessage = !response.IsBlocked && !response.BlockedByThem
	response.CanCreateGroup = response.IsFriend
	response.CanSendFile = response.IsFriend
	response.CanCall = response.IsFriend
	return response
}

func loadPresenceMap(userIDs []uint, viewerID uint) map[uint]socialPresenceResponse {
	result := map[uint]socialPresenceResponse{}
	userIDs = dedupeUintSlice(userIDs)
	if len(userIDs) == 0 {
		return result
	}
	onlineMap := snapshotOnlinePresence()
	var list []models.UserPresenceModel
	global.DB.Where("user_id IN ?", userIDs).Find(&list)
	for _, item := range list {
		visibleInvisible := item.IsInvisible && viewerID != item.UserID
		result[item.UserID] = socialPresenceResponse{
			Mode:         normalizePresenceMode(item.Mode),
			StatusText:   strings.TrimSpace(item.StatusText),
			IsInvisible:  item.IsInvisible,
			IsOnline:     onlineMap[item.UserID] && !visibleInvisible,
			LastActiveAt: item.LastActiveAt,
		}
	}
	for _, userID := range userIDs {
		if _, ok := result[userID]; ok {
			continue
		}
		result[userID] = socialPresenceResponse{
			Mode:         "online",
			StatusText:   "",
			IsInvisible:  false,
			IsOnline:     onlineMap[userID],
			LastActiveAt: nil,
		}
	}
	for userID, item := range result {
		if !item.IsOnline && item.Mode == "online" && !(item.IsInvisible && viewerID != userID) {
			item.Mode = "offline"
		}
		result[userID] = item
	}
	return result
}

func savePresence(userID uint, req socialPresenceRequest) socialPresenceResponse {
	mode := normalizePresenceMode(req.Mode)
	statusText := strings.TrimSpace(req.StatusText)
	if len([]rune(statusText)) > maxStatusTextLength {
		statusText = string([]rune(statusText)[:maxStatusTextLength])
	}
	now := time.Now()
	model := models.UserPresenceModel{
		UserID:       userID,
		Mode:         mode,
		StatusText:   statusText,
		IsInvisible:  req.IsInvisible,
		LastActiveAt: &now,
	}
	global.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"mode", "status_text", "is_invisible", "last_active_at", "updated_at"}),
	}).Create(&model)
	return loadPresenceMap([]uint{userID}, userID)[userID]
}

func touchPresence(userID uint) {
	now := time.Now()
	model := models.UserPresenceModel{
		UserID:       userID,
		Mode:         "online",
		LastActiveAt: &now,
	}
	global.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"last_active_at": &now,
			"updated_at":     now,
		}),
	}).Create(&model)
}

func upsertConversationRead(userID uint, conversationKey string, lastID uint) {
	if userID == 0 || conversationKey == "" || lastID == 0 {
		return
	}
	now := time.Now()
	model := models.SocialConversationReadModel{
		UserID:            userID,
		ConversationKey:   conversationKey,
		LastReadMessageID: lastID,
		LastReadAt:        &now,
	}
	global.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "conversation_key"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_read_message_id", "last_read_at", "updated_at"}),
	}).Create(&model)
}

func fetchReadMap(userID uint) map[string]uint {
	result := map[string]uint{}
	var list []models.SocialConversationReadModel
	global.DB.Where("user_id = ?", userID).Find(&list)
	for _, item := range list {
		result[item.ConversationKey] = item.LastReadMessageID
	}
	return result
}

func loadDirectActivityMap(userID uint, targetIDs []uint) map[uint]models.SocialMessageModel {
	result := map[uint]models.SocialMessageModel{}
	targetIDs = dedupeUintSlice(targetIDs)
	if len(targetIDs) == 0 {
		return result
	}
	var list []models.SocialMessageModel
	global.DB.Where(
		"conversation_type = ? AND ((send_user_id = ? AND receive_user_id IN ?) OR (send_user_id IN ? AND receive_user_id = ?))",
		string(models.SocialConversationDirect), userID, targetIDs, targetIDs, userID,
	).Order("id desc").Find(&list)
	for _, item := range list {
		targetID := item.SendUserID
		if targetID == userID {
			targetID = item.ReceiveUserID
		}
		if _, ok := result[targetID]; ok {
			continue
		}
		result[targetID] = item
	}
	return result
}

func isGroupMember(groupID, userID uint) bool {
	var count int64
	global.DB.Model(&models.SocialGroupMemberModel{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
	return count > 0
}

func getGroupMemberIDs(groupID uint) []uint {
	var list []models.SocialGroupMemberModel
	global.DB.Select("user_id").Where("group_id = ?", groupID).Find(&list)
	result := make([]uint, 0, len(list))
	for _, item := range list {
		result = append(result, item.UserID)
	}
	return result
}

func buildFriendCards(userID uint, key string) []socialFollowCard {
	friendIDs := fetchFriendIDs(userID)
	userMap := loadUserMap(friendIDs)
	presenceMap := loadPresenceMap(friendIDs, userID)
	activityMap := loadDirectActivityMap(userID, friendIDs)
	readMap := fetchReadMap(userID)
	text := strings.TrimSpace(strings.ToLower(key))
	result := make([]socialFollowCard, 0)
	for _, friendID := range friendIDs {
		user, ok := userMap[friendID]
		if !ok {
			continue
		}
		searchText := strings.ToLower(strings.TrimSpace(user.NickName + " " + user.UserName))
		if text != "" && !strings.Contains(searchText, text) {
			continue
		}
		card := socialFollowCard{
			UserID:           user.ID,
			BlogNo:           user.ID,
			NickName:         user.NickName,
			UserName:         user.UserName,
			Avatar:           user.Avatar,
			IsFollowing:      true,
			FollowsMe:        true,
			IsFriend:         true,
			CanDirectMessage: true,
			CanCreateGroup:   true,
			CanSendFile:      true,
			CanCall:          true,
		}
		if presence, ok := presenceMap[friendID]; ok {
			card.IsOnline = presence.IsOnline
			card.PresenceMode = presence.Mode
			card.PresenceText = presence.StatusText
			card.IsInvisible = presence.IsInvisible
			card.LastActiveAt = presence.LastActiveAt
		}
		if latest, ok := activityMap[friendID]; ok {
			card.LastMessageAt = &latest.CreatedAt
			card.LastMessagePreview = messagePreview(latest)
			key := buildDirectConversationKey(userID, friendID)
			lastReadID := readMap[key]
			var unread int64
			global.DB.Model(&models.SocialMessageModel{}).
				Where("conversation_key = ? AND send_user_id = ? AND id > ?", key, friendID, lastReadID).
				Count(&unread)
			card.UnreadCount = int(unread)
		}
		result = append(result, card)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.IsOnline != right.IsOnline {
			return left.IsOnline
		}
		if left.LastMessageAt == nil && right.LastMessageAt == nil {
			return left.UserID < right.UserID
		}
		if left.LastMessageAt == nil {
			return false
		}
		if right.LastMessageAt == nil {
			return true
		}
		return left.LastMessageAt.After(*right.LastMessageAt)
	})
	return result
}

func guessMimeType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	if ext == "" {
		return "application/octet-stream"
	}
	if value := mime.TypeByExtension(ext); value != "" {
		return value
	}
	return "application/octet-stream"
}
