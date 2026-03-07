package social_api

import (
	"fmt"
	"sort"
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type socialGroupCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Notice    string `json:"notice"`
	MemberIDs []uint `json:"member_ids"`
}

type socialGroupMessageRequest struct {
	Content string `json:"content"`
	MsgType string `json:"msg_type"`
	FileID  uint   `json:"file_id"`
}

// GroupListView 获取当前用户所在群组。
func (SocialApi) GroupListView(c *gin.Context) {
	claims := getClaims(c)
	groupIDs := getCurrentUserGroupIDs(claims.UserID)
	var groups []models.SocialGroupModel
	if len(groupIDs) > 0 {
		global.DB.Where("id IN ?", groupIDs).Order("updated_at desc").Find(&groups)
	}
	result := make([]map[string]any, 0, len(groups))
	for _, item := range groups {
		groupNo := ensureGroupNo(&item)
		memberIDs := getGroupMemberIDs(item.ID)
		result = append(result, map[string]any{
			"id":               item.ID,
			"group_no":         groupNo,
			"name":             item.Name,
			"avatar":           item.Avatar,
			"notice":           item.Notice,
			"owner_user_id":    item.OwnerUserID,
			"member_count":     len(memberIDs),
			"conversation_key": buildGroupConversationKey(item.ID),
			"member_ids":       memberIDs,
			"created_at":       item.CreatedAt,
		})
	}
	res.OkWithData(result, c)
}

// GroupCreateView 创建好友群组。
func (SocialApi) GroupCreateView(c *gin.Context) {
	var cr socialGroupCreateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	name := strings.TrimSpace(cr.Name)
	if name == "" {
		res.FailWithMessage("群组名称不能为空", c)
		return
	}
	if len([]rune(name)) > 30 {
		name = string([]rune(name)[:30])
	}
	notice := strings.TrimSpace(cr.Notice)
	if len([]rune(notice)) > 100 {
		notice = string([]rune(notice)[:100])
	}

	friendSet := map[uint]struct{}{}
	for _, item := range fetchFriendIDs(claims.UserID) {
		friendSet[item] = struct{}{}
	}
	memberIDs := dedupeUintSlice(cr.MemberIDs)
	filteredMemberIDs := make([]uint, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		if memberID == 0 || memberID == claims.UserID {
			continue
		}
		if _, ok := friendSet[memberID]; !ok {
			res.FailWithMessage("仅可邀请好友入群", c)
			return
		}
		filteredMemberIDs = append(filteredMemberIDs, memberID)
	}
	if len(filteredMemberIDs) == 0 {
		res.FailWithMessage("请至少选择一位好友", c)
		return
	}
	if len(filteredMemberIDs)+1 > maxGroupMembers {
		res.FailWithMessage(fmt.Sprintf("群成员上限为 %d 人", maxGroupMembers), c)
		return
	}

	var owner models.UserModel
	if err := global.DB.Take(&owner, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}
	userMap := loadUserMap(filteredMemberIDs)
	group := models.SocialGroupModel{
		OwnerUserID: claims.UserID,
		Name:        name,
		Avatar:      owner.Avatar,
		Notice:      notice,
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		group.GroupNo = fmt.Sprintf("G%06d", group.ID)
		if err := tx.Model(&group).Update("group_no", group.GroupNo).Error; err != nil {
			return err
		}
		memberRows := []models.SocialGroupMemberModel{
			{
				GroupID:  group.ID,
				UserID:   owner.ID,
				Role:     string(models.SocialGroupRoleOwner),
				NickName: owner.NickName,
				Avatar:   owner.Avatar,
			},
		}
		for _, memberID := range filteredMemberIDs {
			user := userMap[memberID]
			memberRows = append(memberRows, models.SocialGroupMemberModel{
				GroupID:  group.ID,
				UserID:   user.ID,
				Role:     string(models.SocialGroupRoleMember),
				NickName: user.NickName,
				Avatar:   user.Avatar,
			})
		}
		return tx.Create(&memberRows).Error
	}); err != nil {
		res.FailWithMessage("创建群组失败", c)
		return
	}

	memberIDs = append(filteredMemberIDs, claims.UserID)
	sort.Slice(memberIDs, func(i, j int) bool { return memberIDs[i] < memberIDs[j] })
	broadcastSocketEvent(filteredMemberIDs, "group_created", map[string]any{
		"group_id":         group.ID,
		"name":             group.Name,
		"conversation_key": buildGroupConversationKey(group.ID),
	})
	res.OkWithData(map[string]any{
		"id":               group.ID,
		"group_no":         group.GroupNo,
		"name":             group.Name,
		"avatar":           group.Avatar,
		"notice":           group.Notice,
		"owner_user_id":    group.OwnerUserID,
		"member_ids":       memberIDs,
		"conversation_key": buildGroupConversationKey(group.ID),
	}, c)
}

// GroupMessageListView 获取群消息，并自动标记已读。
func (SocialApi) GroupMessageListView(c *gin.Context) {
	var uri socialGroupURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if !isGroupMember(uri.ID, claims.UserID) {
		res.FailWithMessage("你不在该群组中", c)
		return
	}
	key := buildGroupConversationKey(uri.ID)
	var list []models.SocialMessageModel
	if err := global.DB.Where("conversation_key = ?", key).Order("id asc").Find(&list).Error; err != nil {
		res.FailWithMessage("获取群消息失败", c)
		return
	}
	if len(list) > 0 {
		upsertConversationRead(claims.UserID, key, list[len(list)-1].ID)
	}
	res.OkWithData(buildGroupMessageResponses(claims.UserID, uri.ID, list), c)
}

// GroupMessageCreateView 发送群消息。
func (SocialApi) GroupMessageCreateView(c *gin.Context) {
	var uri socialGroupURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr socialGroupMessageRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if !isGroupMember(uri.ID, claims.UserID) {
		res.FailWithMessage("你不在该群组中", c)
		return
	}
	var sender models.UserModel
	if err := global.DB.Take(&sender, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}

	msgType := strings.TrimSpace(cr.MsgType)
	if msgType == "" {
		msgType = string(models.SocialMessageText)
	}
	content := strings.TrimSpace(cr.Content)
	if msgType == string(models.SocialMessageText) && content == "" {
		res.FailWithMessage("消息内容不能为空", c)
		return
	}
	if len([]rune(content)) > maxMessageLength {
		content = string([]rune(content)[:maxMessageLength])
	}

	message := models.SocialMessageModel{
		ConversationKey:  buildGroupConversationKey(uri.ID),
		ConversationType: string(models.SocialConversationGroup),
		GroupID:          uri.ID,
		SendUserID:       sender.ID,
		SendUserNickName: sender.NickName,
		SendUserAvatar:   sender.Avatar,
		MsgType:          msgType,
		Content:          content,
	}
	if msgType == string(models.SocialMessageFile) {
		if cr.FileID == 0 {
			res.FailWithMessage("请选择要发送的文件", c)
			return
		}
		var file models.SocialFileModel
		if err := global.DB.Take(&file, "id = ? AND user_id = ?", cr.FileID, claims.UserID).Error; err != nil {
			res.FailWithMessage("文件不存在或不属于当前用户", c)
			return
		}
		message.FileID = file.ID
		message.FileName = file.Name
		message.FileSize = file.Size
		message.FileMime = file.Mime
		message.FileURL = "/api/social/files/" + fmt.Sprint(file.ID) + "/download"
		if content == "" {
			message.Content = "发送了一个文件"
		}
	}
	if err := global.DB.Create(&message).Error; err != nil {
		res.FailWithMessage("发送群消息失败", c)
		return
	}
	memberIDs := getGroupMemberIDs(uri.ID)
	broadcastSocketEvent(memberIDs, "new_group_message", message)
	res.OkWithData(buildGroupMessageResponses(claims.UserID, uri.ID, []models.SocialMessageModel{message})[0], c)
}
