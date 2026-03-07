package social_api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type socialDiscoveryQuery struct {
	Key string `form:"key" binding:"required"`
}

type socialGroupJoinRequest struct {
	GroupNo string `json:"group_no" binding:"required"`
}

type socialGroupMemberSaveRequest struct {
	MemberIDs []uint `json:"member_ids" binding:"required"`
}

type socialGroupMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type socialGroupTransferRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

type socialGroupMemberURI struct {
	ID     uint `uri:"id" binding:"required"`
	UserID uint `uri:"user_id" binding:"required"`
}

// DiscoveryView 搜索博客号/群组号，方便加好友和进群。
func (SocialApi) DiscoveryView(c *gin.Context) {
	var cr socialDiscoveryQuery
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	key := strings.TrimSpace(cr.Key)
	if key == "" {
		res.FailWithMessage("请输入博客号、昵称或群组号", c)
		return
	}

	claims := getClaims(c)
	userIDs := make([]uint, 0)
	userResult := make([]socialDiscoveryUserItem, 0)
	userQuery := global.DB.Model(&models.UserModel{}).Where("role <> ?", ctype.PermissionDisableUser)
	if id, err := strconv.ParseUint(key, 10, 64); err == nil && id > 0 {
		userQuery = userQuery.Where("id = ? OR nick_name LIKE ? OR user_name LIKE ?", id, "%"+key+"%", "%"+key+"%")
	} else {
		userQuery = userQuery.Where("nick_name LIKE ? OR user_name LIKE ?", "%"+key+"%", "%"+key+"%")
	}
	var users []models.UserModel
	userQuery.Order("id asc").Limit(10).Find(&users)
	for _, item := range users {
		userIDs = append(userIDs, item.ID)
	}
	presenceMap := loadPresenceMap(userIDs, claims.UserID)
	for _, item := range users {
		presence := presenceMap[item.ID]
		userResult = append(userResult, socialDiscoveryUserItem{
			UserID:     item.ID,
			BlogNo:     item.ID,
			UserName:   item.UserName,
			NickName:   item.NickName,
			Avatar:     item.Avatar,
			Relation:   buildRelation(claims.UserID, item.ID),
			IsSelf:     item.ID == claims.UserID,
			IsOnline:   presence.IsOnline,
			LastActive: presence.LastActiveAt,
		})
	}

	groupResult := make([]socialDiscoveryGroupItem, 0)
	var groups []models.SocialGroupModel
	groupQuery := global.DB.Model(&models.SocialGroupModel{})
	groupQuery = groupQuery.Where("group_no = ? OR name LIKE ?", key, "%"+key+"%")
	groupQuery.Order("updated_at desc").Limit(10).Find(&groups)
	currentGroupIDs := map[uint]struct{}{}
	for _, groupID := range getCurrentUserGroupIDs(claims.UserID) {
		currentGroupIDs[groupID] = struct{}{}
	}
	for _, item := range groups {
		groupNo := ensureGroupNo(&item)
		groupResult = append(groupResult, socialDiscoveryGroupItem{
			ID:          item.ID,
			GroupNo:     groupNo,
			Name:        item.Name,
			Avatar:      item.Avatar,
			Notice:      item.Notice,
			OwnerUserID: item.OwnerUserID,
			MemberCount: len(getGroupMemberIDs(item.ID)),
			IsJoined:    containsGroup(currentGroupIDs, item.ID),
		})
	}

	res.OkWithData(gin.H{
		"users":  userResult,
		"groups": groupResult,
	}, c)
}

// GroupDetailView 获取群详情与成员列表。
func (SocialApi) GroupDetailView(c *gin.Context) {
	var uri socialGroupURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var group models.SocialGroupModel
	if err := global.DB.Take(&group, uri.ID).Error; err != nil {
		res.FailWithMessage("群组不存在", c)
		return
	}
	if !isGroupMember(group.ID, claims.UserID) {
		res.FailWithMessage("你不在该群组中", c)
		return
	}

	role := getGroupMemberRole(group.ID, claims.UserID)
	res.OkWithData(gin.H{
		"id":                group.ID,
		"group_no":          ensureGroupNo(&group),
		"name":              group.Name,
		"avatar":            group.Avatar,
		"notice":            group.Notice,
		"owner_user_id":     group.OwnerUserID,
		"viewer_role":       role,
		"viewer_role_label": groupRoleLabel(role),
		"can_manage":        isGroupManagerRole(role),
		"members":           buildGroupMemberItems(group.ID),
		"conversation_key":  buildGroupConversationKey(group.ID),
	}, c)
}

// GroupJoinView 根据群组号加入群组。
func (SocialApi) GroupJoinView(c *gin.Context) {
	var cr socialGroupJoinRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	groupNo := strings.ToUpper(strings.TrimSpace(cr.GroupNo))
	if groupNo == "" {
		res.FailWithMessage("群组号不能为空", c)
		return
	}

	var group models.SocialGroupModel
	if err := global.DB.Where("group_no = ?", groupNo).Take(&group).Error; err != nil {
		res.FailWithMessage("群组不存在", c)
		return
	}
	if isGroupMember(group.ID, claims.UserID) {
		res.FailWithMessage("你已经在群里了", c)
		return
	}
	memberIDs := getGroupMemberIDs(group.ID)
	if len(memberIDs) >= maxGroupMembers {
		res.FailWithMessage(fmt.Sprintf("群成员上限为 %d 人", maxGroupMembers), c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMessage("当前用户不存在", c)
		return
	}
	member := models.SocialGroupMemberModel{
		GroupID:  group.ID,
		UserID:   user.ID,
		Role:     string(models.SocialGroupRoleMember),
		NickName: user.NickName,
		Avatar:   user.Avatar,
	}
	if err := global.DB.Create(&member).Error; err != nil {
		res.FailWithMessage("加入群组失败", c)
		return
	}
	touchGroupUpdatedAt(group.ID)
	sendSocketEvent(user.ID, "group_created", gin.H{
		"group_id":         group.ID,
		"name":             group.Name,
		"conversation_key": buildGroupConversationKey(group.ID),
	})
	broadcastSocketEvent(getGroupMemberIDs(group.ID), "group_updated", gin.H{
		"group_id": group.ID,
	})
	res.OkWithData(gin.H{
		"id":               group.ID,
		"group_no":         ensureGroupNo(&group),
		"name":             group.Name,
		"conversation_key": buildGroupConversationKey(group.ID),
	}, c)
}

// GroupMemberAddView 群主或管理员拉人进群。
func (SocialApi) GroupMemberAddView(c *gin.Context) {
	var uri socialGroupURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr socialGroupMemberSaveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if !canOperateGroup(uri.ID, claims.UserID) {
		res.FailWithMessage("无权管理该群组", c)
		return
	}

	friendSet := map[uint]struct{}{}
	for _, item := range fetchFriendIDs(claims.UserID) {
		friendSet[item] = struct{}{}
	}
	userMap := loadUserMap(cr.MemberIDs)
	existingMembers := map[uint]struct{}{}
	for _, userID := range getGroupMemberIDs(uri.ID) {
		existingMembers[userID] = struct{}{}
	}
	createdRows := make([]models.SocialGroupMemberModel, 0)
	for _, memberID := range dedupeUintSlice(cr.MemberIDs) {
		if memberID == 0 {
			continue
		}
		if _, ok := existingMembers[memberID]; ok {
			continue
		}
		if _, ok := friendSet[memberID]; !ok {
			continue
		}
		user, ok := userMap[memberID]
		if !ok {
			continue
		}
		createdRows = append(createdRows, models.SocialGroupMemberModel{
			GroupID:  uri.ID,
			UserID:   user.ID,
			Role:     string(models.SocialGroupRoleMember),
			NickName: user.NickName,
			Avatar:   user.Avatar,
		})
	}
	if len(createdRows) == 0 {
		res.FailWithMessage("没有可加入的好友", c)
		return
	}
	if len(existingMembers)+len(createdRows) > maxGroupMembers {
		res.FailWithMessage(fmt.Sprintf("群成员上限为 %d 人", maxGroupMembers), c)
		return
	}
	if err := global.DB.Create(&createdRows).Error; err != nil {
		res.FailWithMessage("拉人进群失败", c)
		return
	}
	touchGroupUpdatedAt(uri.ID)
	memberIDs := make([]uint, 0, len(createdRows))
	for _, item := range createdRows {
		memberIDs = append(memberIDs, item.UserID)
	}
	for _, userID := range memberIDs {
		sendSocketEvent(userID, "group_created", gin.H{
			"group_id":         uri.ID,
			"conversation_key": buildGroupConversationKey(uri.ID),
		})
	}
	broadcastSocketEvent(getGroupMemberIDs(uri.ID), "group_updated", gin.H{
		"group_id": uri.ID,
	})
	res.OkWithData(gin.H{
		"members": buildGroupMemberItems(uri.ID),
	}, c)
}

// GroupMemberRoleUpdateView 设置群管理员或取消管理员。
func (SocialApi) GroupMemberRoleUpdateView(c *gin.Context) {
	var uri socialGroupMemberURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr socialGroupMemberRoleRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if getGroupMemberRole(uri.ID, claims.UserID) != string(models.SocialGroupRoleOwner) {
		res.FailWithMessage("只有群主可以设置管理员", c)
		return
	}

	role := strings.TrimSpace(strings.ToLower(cr.Role))
	if role != string(models.SocialGroupRoleAdmin) && role != string(models.SocialGroupRoleMember) {
		res.FailWithMessage("角色只能是 admin 或 member", c)
		return
	}
	var target models.SocialGroupMemberModel
	if err := global.DB.Take(&target, "group_id = ? AND user_id = ?", uri.ID, uri.UserID).Error; err != nil {
		res.FailWithMessage("群成员不存在", c)
		return
	}
	if target.Role == string(models.SocialGroupRoleOwner) {
		res.FailWithMessage("不能修改群主角色", c)
		return
	}
	if err := global.DB.Model(&target).Update("role", role).Error; err != nil {
		res.FailWithMessage("更新角色失败", c)
		return
	}
	touchGroupUpdatedAt(uri.ID)
	broadcastSocketEvent(getGroupMemberIDs(uri.ID), "group_updated", gin.H{"group_id": uri.ID})
	res.OkWithData(gin.H{"members": buildGroupMemberItems(uri.ID)}, c)
}

// GroupMemberRemoveView 踢人或主动退群。
func (SocialApi) GroupMemberRemoveView(c *gin.Context) {
	var uri socialGroupMemberURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var target models.SocialGroupMemberModel
	if err := global.DB.Take(&target, "group_id = ? AND user_id = ?", uri.ID, uri.UserID).Error; err != nil {
		res.FailWithMessage("群成员不存在", c)
		return
	}
	if uri.UserID == claims.UserID {
		if target.Role == string(models.SocialGroupRoleOwner) {
			res.FailWithMessage("群主请先转让群主后再退出", c)
			return
		}
	} else if !canKickGroupMember(uri.ID, claims.UserID, target.Role, uri.UserID) {
		res.FailWithMessage("无权移除该群成员", c)
		return
	}
	if err := global.DB.Delete(&target).Error; err != nil {
		res.FailWithMessage("移除群成员失败", c)
		return
	}
	touchGroupUpdatedAt(uri.ID)
	sendSocketEvent(uri.UserID, "group_removed", gin.H{"group_id": uri.ID})
	broadcastSocketEvent(getGroupMemberIDs(uri.ID), "group_updated", gin.H{"group_id": uri.ID})
	res.OkWithData(gin.H{"members": buildGroupMemberItems(uri.ID)}, c)
}

// GroupTransferOwnerView 转让群主。
func (SocialApi) GroupTransferOwnerView(c *gin.Context) {
	var uri socialGroupURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr socialGroupTransferRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if getGroupMemberRole(uri.ID, claims.UserID) != string(models.SocialGroupRoleOwner) {
		res.FailWithMessage("只有群主可以转让群主", c)
		return
	}
	if claims.UserID == cr.UserID || cr.UserID == 0 {
		res.FailWithMessage("请选择新的群主", c)
		return
	}
	var target models.SocialGroupMemberModel
	if err := global.DB.Take(&target, "group_id = ? AND user_id = ?", uri.ID, cr.UserID).Error; err != nil {
		res.FailWithMessage("目标成员不存在", c)
		return
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.SocialGroupModel{}).Where("id = ?", uri.ID).Updates(map[string]any{
			"owner_user_id": cr.UserID,
			"updated_at":    time.Now(),
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SocialGroupMemberModel{}).
			Where("group_id = ? AND user_id = ?", uri.ID, claims.UserID).
			Update("role", string(models.SocialGroupRoleAdmin)).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SocialGroupMemberModel{}).
			Where("group_id = ? AND user_id = ?", uri.ID, cr.UserID).
			Update("role", string(models.SocialGroupRoleOwner)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		res.FailWithMessage("转让群主失败", c)
		return
	}
	broadcastSocketEvent(getGroupMemberIDs(uri.ID), "group_updated", gin.H{"group_id": uri.ID})
	res.OkWithData(gin.H{"members": buildGroupMemberItems(uri.ID)}, c)
}

func containsGroup(groupMap map[uint]struct{}, groupID uint) bool {
	_, ok := groupMap[groupID]
	return ok
}

func touchGroupUpdatedAt(groupID uint) {
	if groupID == 0 {
		return
	}
	global.DB.Model(&models.SocialGroupModel{}).Where("id = ?", groupID).Update("updated_at", time.Now())
}
