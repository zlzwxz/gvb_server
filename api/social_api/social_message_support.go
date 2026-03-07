package social_api

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"

	"gorm.io/gorm/clause"
)

type socialMessageResponse struct {
	models.SocialMessageModel
	IsMine         bool   `json:"is_mine"`
	IsRead         bool   `json:"is_read"`
	ReadCount      int    `json:"read_count"`
	MemberCount    int    `json:"member_count"`
	ReadUserIDs    []uint `json:"read_user_ids"`
	CanRecall      bool   `json:"can_recall"`
	ReadStatusText string `json:"read_status_text"`
}

type socialGroupMemberItem struct {
	UserID    uint   `json:"user_id"`
	BlogNo    uint   `json:"blog_no"`
	UserName  string `json:"user_name"`
	NickName  string `json:"nick_name"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	RoleLabel string `json:"role_label"`
}

type socialDiscoveryUserItem struct {
	UserID     uint                   `json:"user_id"`
	BlogNo     uint                   `json:"blog_no"`
	UserName   string                 `json:"user_name"`
	NickName   string                 `json:"nick_name"`
	Avatar     string                 `json:"avatar"`
	Relation   socialRelationResponse `json:"relation"`
	IsSelf     bool                   `json:"is_self"`
	IsOnline   bool                   `json:"is_online"`
	LastActive *time.Time             `json:"last_active_at"`
}

type socialDiscoveryGroupItem struct {
	ID          uint   `json:"id"`
	GroupNo     string `json:"group_no"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Notice      string `json:"notice"`
	OwnerUserID uint   `json:"owner_user_id"`
	MemberCount int    `json:"member_count"`
	IsJoined    bool   `json:"is_joined"`
}

type socialCallLogItem struct {
	ID              uint       `json:"id"`
	CallID          string     `json:"call_id"`
	ConversationKey string     `json:"conversation_key"`
	Status          string     `json:"status"`
	StatusLabel     string     `json:"status_label"`
	Direction       string     `json:"direction"`
	PartnerUserID   uint       `json:"partner_user_id"`
	PartnerNickName string     `json:"partner_nick_name"`
	PartnerAvatar   string     `json:"partner_avatar"`
	StartedAt       *time.Time `json:"started_at"`
	AnsweredAt      *time.Time `json:"answered_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationSec     int        `json:"duration_sec"`
	IsMissed        bool       `json:"is_missed"`
}

func groupRoleLabel(role string) string {
	switch strings.TrimSpace(role) {
	case string(models.SocialGroupRoleOwner):
		return "群主"
	case string(models.SocialGroupRoleAdmin):
		return "管理员"
	default:
		return "成员"
	}
}

func callStatusLabel(status string) string {
	switch strings.TrimSpace(status) {
	case string(models.SocialCallStatusRejected):
		return "已拒绝"
	case string(models.SocialCallStatusMissed):
		return "未接来电"
	case string(models.SocialCallStatusCompleted):
		return "已完成"
	case string(models.SocialCallStatusCanceled):
		return "已取消"
	default:
		return "呼叫中"
	}
}

func ensureGroupNo(group *models.SocialGroupModel) string {
	if group == nil {
		return ""
	}
	if strings.TrimSpace(group.GroupNo) != "" {
		return strings.TrimSpace(group.GroupNo)
	}
	if group.ID == 0 {
		return ""
	}
	groupNo := fmt.Sprintf("G%06d", group.ID)
	global.DB.Model(group).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND (group_no IS NULL OR group_no = '')", group.ID).
		Update("group_no", groupNo)
	group.GroupNo = groupNo
	return groupNo
}

func getGroupMemberModels(groupID uint) []models.SocialGroupMemberModel {
	var list []models.SocialGroupMemberModel
	global.DB.Where("group_id = ?", groupID).Order("id asc").Find(&list)
	return list
}

func getGroupMemberRole(groupID, userID uint) string {
	if groupID == 0 || userID == 0 {
		return ""
	}
	var member models.SocialGroupMemberModel
	if err := global.DB.Take(&member, "group_id = ? AND user_id = ?", groupID, userID).Error; err != nil {
		return ""
	}
	return strings.TrimSpace(member.Role)
}

func isGroupManagerRole(role string) bool {
	switch strings.TrimSpace(role) {
	case string(models.SocialGroupRoleOwner), string(models.SocialGroupRoleAdmin):
		return true
	default:
		return false
	}
}

func canOperateGroup(groupID uint, userID uint) bool {
	return isGroupManagerRole(getGroupMemberRole(groupID, userID))
}

func canKickGroupMember(groupID uint, actorID uint, targetRole string, targetUserID uint) bool {
	role := getGroupMemberRole(groupID, actorID)
	if role == string(models.SocialGroupRoleOwner) {
		return actorID != targetUserID
	}
	if role == string(models.SocialGroupRoleAdmin) {
		return targetRole == string(models.SocialGroupRoleMember) && actorID != targetUserID
	}
	return false
}

func isUserCurrentlyOnline(userID uint) bool {
	return snapshotOnlinePresence()[userID]
}

func buildGroupMemberItems(groupID uint) []socialGroupMemberItem {
	members := getGroupMemberModels(groupID)
	userIDs := make([]uint, 0, len(members))
	for _, item := range members {
		userIDs = append(userIDs, item.UserID)
	}
	userMap := loadUserMap(userIDs)
	result := make([]socialGroupMemberItem, 0, len(members))
	for _, item := range members {
		user := userMap[item.UserID]
		role := strings.TrimSpace(item.Role)
		result = append(result, socialGroupMemberItem{
			UserID:    item.UserID,
			BlogNo:    item.UserID,
			UserName:  user.UserName,
			NickName:  firstNonEmpty(item.NickName, user.NickName, user.UserName, fmt.Sprintf("用户%d", item.UserID)),
			Avatar:    firstNonEmpty(item.Avatar, user.Avatar),
			Role:      role,
			RoleLabel: groupRoleLabel(role),
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		leftWeight := groupRoleWeight(result[i].Role)
		rightWeight := groupRoleWeight(result[j].Role)
		if leftWeight != rightWeight {
			return leftWeight < rightWeight
		}
		return result[i].UserID < result[j].UserID
	})
	return result
}

func groupRoleWeight(role string) int {
	switch strings.TrimSpace(role) {
	case string(models.SocialGroupRoleOwner):
		return 1
	case string(models.SocialGroupRoleAdmin):
		return 2
	default:
		return 3
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text != "" {
			return text
		}
	}
	return ""
}

func canRecallMessage(item models.SocialMessageModel, viewerID uint) bool {
	if viewerID == 0 || item.ID == 0 || item.IsRecalled {
		return false
	}
	if isPlatformAdmin(viewerID) {
		return true
	}
	if item.SendUserID == viewerID {
		return true
	}
	if item.ConversationType == string(models.SocialConversationGroup) && canOperateGroup(item.GroupID, viewerID) {
		return true
	}
	return false
}

func buildDirectMessageResponses(viewerID, peerID uint, list []models.SocialMessageModel) []socialMessageResponse {
	key := buildDirectConversationKey(viewerID, peerID)
	readMap := fetchReadMap(peerID)
	peerReadCursor := readMap[key]
	viewerReadCursor := fetchReadMap(viewerID)[key]

	result := make([]socialMessageResponse, 0, len(list))
	for _, item := range list {
		isMine := item.SendUserID == viewerID
		isRead := viewerReadCursor >= item.ID
		if isMine {
			isRead = peerReadCursor >= item.ID
		}
		statusText := "已读"
		if !isRead && isMine {
			statusText = "未读"
		}
		result = append(result, socialMessageResponse{
			SocialMessageModel: item,
			IsMine:             isMine,
			IsRead:             isRead,
			ReadCount:          boolToInt(isRead),
			MemberCount:        1,
			ReadUserIDs:        nil,
			CanRecall:          canRecallMessage(item, viewerID),
			ReadStatusText:     statusText,
		})
	}
	return result
}

func buildGroupMessageResponses(viewerID, groupID uint, list []models.SocialMessageModel) []socialMessageResponse {
	memberIDs := getGroupMemberIDs(groupID)
	readCursorMap := map[uint]uint{}
	var readRows []models.SocialConversationReadModel
	global.DB.Where("conversation_key = ? AND user_id IN ?", buildGroupConversationKey(groupID), memberIDs).Find(&readRows)
	for _, row := range readRows {
		readCursorMap[row.UserID] = row.LastReadMessageID
	}

	result := make([]socialMessageResponse, 0, len(list))
	for _, item := range list {
		readUserIDs := make([]uint, 0, len(memberIDs))
		for _, userID := range memberIDs {
			if userID == item.SendUserID {
				readUserIDs = append(readUserIDs, userID)
				continue
			}
			if readCursorMap[userID] >= item.ID {
				readUserIDs = append(readUserIDs, userID)
			}
		}
		sort.Slice(readUserIDs, func(i, j int) bool { return readUserIDs[i] < readUserIDs[j] })
		result = append(result, socialMessageResponse{
			SocialMessageModel: item,
			IsMine:             item.SendUserID == viewerID,
			IsRead:             containsUint(readUserIDs, viewerID),
			ReadCount:          len(readUserIDs),
			MemberCount:        len(memberIDs),
			ReadUserIDs:        readUserIDs,
			CanRecall:          canRecallMessage(item, viewerID),
			ReadStatusText:     fmt.Sprintf("%d/%d 已读", len(readUserIDs), len(memberIDs)),
		})
	}
	return result
}

func containsUint(list []uint, target uint) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func buildCallLogItem(log models.SocialCallLogModel, viewerID uint) socialCallLogItem {
	direction := "outgoing"
	partnerID := log.CalleeUserID
	partnerName := log.CalleeNickName
	partnerAvatar := log.CalleeAvatar
	if viewerID == log.CalleeUserID {
		direction = "incoming"
		partnerID = log.CallerUserID
		partnerName = log.CallerNickName
		partnerAvatar = log.CallerAvatar
	}
	return socialCallLogItem{
		ID:              log.ID,
		CallID:          log.CallID,
		ConversationKey: log.ConversationKey,
		Status:          log.Status,
		StatusLabel:     callStatusLabel(log.Status),
		Direction:       direction,
		PartnerUserID:   partnerID,
		PartnerNickName: partnerName,
		PartnerAvatar:   partnerAvatar,
		StartedAt:       log.StartedAt,
		AnsweredAt:      log.AnsweredAt,
		EndedAt:         log.EndedAt,
		DurationSec:     log.DurationSec,
		IsMissed:        log.Status == string(models.SocialCallStatusMissed) && log.MissedByUserID == viewerID,
	}
}

func isPlatformAdmin(userID uint) bool {
	if userID == 0 {
		return false
	}
	var user models.UserModel
	if err := global.DB.Select("id", "role").Take(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == ctype.PermissionAdmin
}
