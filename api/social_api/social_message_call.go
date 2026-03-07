package social_api

import (
	"fmt"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type socialMessageSearchQuery struct {
	UserID  uint   `form:"user_id"`
	GroupID uint   `form:"group_id"`
	Keyword string `form:"keyword" binding:"required"`
	Limit   int    `form:"limit"`
}

type socialMessageURI struct {
	ID uint `uri:"id" binding:"required"`
}

// MessageSearchView 搜索当前会话消息。
func (SocialApi) MessageSearchView(c *gin.Context) {
	var cr socialMessageSearchQuery
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	keyword := strings.TrimSpace(cr.Keyword)
	if keyword == "" {
		res.FailWithMessage("搜索关键词不能为空", c)
		return
	}
	if cr.Limit <= 0 {
		cr.Limit = 20
	}
	if cr.Limit > 100 {
		cr.Limit = 100
	}

	var (
		list []models.SocialMessageModel
		err  error
	)
	if cr.GroupID > 0 {
		if !isGroupMember(cr.GroupID, claims.UserID) {
			res.FailWithMessage("你不在该群组中", c)
			return
		}
		err = global.DB.Where("conversation_key = ? AND is_recalled = ? AND content LIKE ?",
			buildGroupConversationKey(cr.GroupID), false, "%"+keyword+"%").
			Order("id desc").Limit(cr.Limit).Find(&list).Error
		if err != nil {
			res.FailWithMessage("搜索群消息失败", c)
			return
		}
		res.OkWithData(buildGroupMessageResponses(claims.UserID, cr.GroupID, reverseMessageOrder(list)), c)
		return
	}
	if cr.UserID == 0 || cr.UserID == claims.UserID {
		res.FailWithMessage("请指定私信对象或群组", c)
		return
	}
	blocked, blockedBy := hasBlockBetween(claims.UserID, cr.UserID)
	if blocked || blockedBy {
		res.FailWithMessage("存在拉黑关系，无法搜索会话", c)
		return
	}
	err = global.DB.Where("conversation_key = ? AND is_recalled = ? AND content LIKE ?",
		buildDirectConversationKey(claims.UserID, cr.UserID), false, "%"+keyword+"%").
		Order("id desc").Limit(cr.Limit).Find(&list).Error
	if err != nil {
		res.FailWithMessage("搜索私信失败", c)
		return
	}
	res.OkWithData(buildDirectMessageResponses(claims.UserID, cr.UserID, reverseMessageOrder(list)), c)
}

// MessageRecallView 撤回消息。
func (SocialApi) MessageRecallView(c *gin.Context) {
	var uri socialMessageURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)

	var message models.SocialMessageModel
	if err := global.DB.Take(&message, uri.ID).Error; err != nil {
		res.FailWithMessage("消息不存在", c)
		return
	}
	if !canRecallMessage(message, claims.UserID) {
		res.FailWithMessage("无权撤回该消息", c)
		return
	}
	if message.IsRecalled {
		res.FailWithMessage("该消息已撤回", c)
		return
	}
	now := time.Now()
	updates := map[string]any{
		"is_recalled": true,
		"recalled_by": claims.UserID,
		"recalled_at": &now,
		"content":     "该消息已撤回",
		"file_id":     uint(0),
		"file_name":   "",
		"file_size":   int64(0),
		"file_mime":   "",
		"file_url":    "",
	}
	if err := global.DB.Model(&message).Updates(updates).Error; err != nil {
		res.FailWithMessage("撤回消息失败", c)
		return
	}
	for key, value := range updates {
		switch key {
		case "is_recalled":
			message.IsRecalled = value.(bool)
		case "recalled_by":
			message.RecalledBy = value.(uint)
		case "recalled_at":
			message.RecalledAt = value.(*time.Time)
		case "content":
			message.Content = value.(string)
		case "file_id":
			message.FileID = value.(uint)
		case "file_name":
			message.FileName = value.(string)
		case "file_size":
			message.FileSize = value.(int64)
		case "file_mime":
			message.FileMime = value.(string)
		case "file_url":
			message.FileURL = value.(string)
		}
	}

	recipients := []uint{}
	if message.ConversationType == string(models.SocialConversationGroup) {
		recipients = getGroupMemberIDs(message.GroupID)
		broadcastSocketEvent(recipients, "message_recalled", gin.H{
			"message_id":       message.ID,
			"group_id":         message.GroupID,
			"conversation_key": message.ConversationKey,
		})
		res.OkWithData(buildGroupMessageResponses(claims.UserID, message.GroupID, []models.SocialMessageModel{message})[0], c)
		return
	}
	recipients = dedupeUintSlice([]uint{message.SendUserID, message.ReceiveUserID})
	broadcastSocketEvent(recipients, "message_recalled", gin.H{
		"message_id":       message.ID,
		"user_id":          otherUserInDirectMessage(message, claims.UserID),
		"conversation_key": message.ConversationKey,
	})
	peerID := otherUserInDirectMessage(message, claims.UserID)
	res.OkWithData(buildDirectMessageResponses(claims.UserID, peerID, []models.SocialMessageModel{message})[0], c)
}

// CallLogListView 获取通话记录和未接来电。
func (SocialApi) CallLogListView(c *gin.Context) {
	claims := getClaims(c)
	status := strings.TrimSpace(c.Query("status"))
	limit := 30
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		fmt.Sscan(raw, &limit)
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	query := global.DB.Model(&models.SocialCallLogModel{}).
		Where("caller_user_id = ? OR callee_user_id = ?", claims.UserID, claims.UserID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var list []models.SocialCallLogModel
	if err := query.Order("id desc").Limit(limit).Find(&list).Error; err != nil {
		res.FailWithMessage("获取通话记录失败", c)
		return
	}
	result := make([]socialCallLogItem, 0, len(list))
	for _, item := range list {
		result = append(result, buildCallLogItem(item, claims.UserID))
	}
	res.OkWithData(result, c)
}

func reverseMessageOrder(list []models.SocialMessageModel) []models.SocialMessageModel {
	for left, right := 0, len(list)-1; left < right; left, right = left+1, right-1 {
		list[left], list[right] = list[right], list[left]
	}
	return list
}

func otherUserInDirectMessage(item models.SocialMessageModel, viewerID uint) uint {
	if item.SendUserID == viewerID {
		return item.ReceiveUserID
	}
	return item.SendUserID
}

func handleCallInvite(senderID uint, req socialSocketRequest) {
	callID := strings.TrimSpace(req.CallID)
	if callID == "" {
		callID = fmt.Sprintf("call_%d_%d", senderID, time.Now().UnixNano())
	}
	if req.TargetUserID == 0 || req.TargetUserID == senderID {
		sendSocketEvent(senderID, "socket_error", gin.H{"message": "目标用户无效"})
		return
	}
	relation := buildRelation(senderID, req.TargetUserID)
	if !relation.CanCall {
		sendSocketEvent(senderID, "socket_error", gin.H{"message": "仅好友之间支持语音通话"})
		return
	}

	users := loadUserMap([]uint{senderID, req.TargetUserID})
	sender, okSender := users[senderID]
	target, okTarget := users[req.TargetUserID]
	if !okSender || !okTarget {
		sendSocketEvent(senderID, "socket_error", gin.H{"message": "用户不存在"})
		return
	}
	now := time.Now()
	log := models.SocialCallLogModel{
		CallID:             callID,
		ConversationKey:    buildDirectConversationKey(senderID, req.TargetUserID),
		ConversationType:   string(models.SocialConversationDirect),
		CallerUserID:       senderID,
		CallerNickName:     sender.NickName,
		CallerAvatar:       sender.Avatar,
		CalleeUserID:       target.ID,
		CalleeNickName:     target.NickName,
		CalleeAvatar:       target.Avatar,
		Status:             string(models.SocialCallStatusRinging),
		StartedAt:          &now,
		LastOperatorUserID: senderID,
	}
	if err := global.DB.Where("call_id = ?", callID).Assign(log).FirstOrCreate(&log).Error; err != nil {
		sendSocketEvent(senderID, "socket_error", gin.H{"message": "创建通话记录失败"})
		return
	}

	if !isUserCurrentlyOnline(req.TargetUserID) {
		finalizeCallLog(&log, string(models.SocialCallStatusMissed), senderID, req.TargetUserID)
		sendSocketEvent(senderID, "call_reject", gin.H{
			"from_user_id": req.TargetUserID,
			"call_id":      callID,
			"payload":      gin.H{"reason": "offline"},
		})
		return
	}

	sendSocketEvent(req.TargetUserID, "call_invite", gin.H{
		"from_user_id": senderID,
		"call_id":      callID,
		"payload":      req.Payload,
	})
}

func handleCallAccept(receiverID uint, req socialSocketRequest) {
	log, err := findCallLogForUser(strings.TrimSpace(req.CallID), receiverID)
	if err != nil {
		sendSocketEvent(receiverID, "socket_error", gin.H{"message": "通话不存在"})
		return
	}
	if log.CalleeUserID != receiverID {
		sendSocketEvent(receiverID, "socket_error", gin.H{"message": "无权接听该通话"})
		return
	}
	if log.AnsweredAt == nil {
		now := time.Now()
		log.AnsweredAt = &now
		log.LastOperatorUserID = receiverID
		global.DB.Model(&log).Updates(map[string]any{
			"answered_at":           &now,
			"last_operator_user_id": receiverID,
		})
	}
	sendSocketEvent(req.TargetUserID, "call_accept", gin.H{
		"from_user_id": receiverID,
		"call_id":      log.CallID,
		"payload":      req.Payload,
	})
}

func handleCallReject(receiverID uint, req socialSocketRequest) {
	log, err := findCallLogForUser(strings.TrimSpace(req.CallID), receiverID)
	if err != nil {
		sendSocketEvent(receiverID, "socket_error", gin.H{"message": "通话不存在"})
		return
	}
	if log.CalleeUserID != receiverID {
		sendSocketEvent(receiverID, "socket_error", gin.H{"message": "无权拒绝该通话"})
		return
	}
	finalizeCallLog(&log, string(models.SocialCallStatusRejected), receiverID, 0)
	sendSocketEvent(req.TargetUserID, "call_reject", gin.H{
		"from_user_id": receiverID,
		"call_id":      log.CallID,
		"payload":      req.Payload,
	})
}

func handleCallEnd(actorID uint, req socialSocketRequest) {
	log, err := findCallLogForUser(strings.TrimSpace(req.CallID), actorID)
	if err != nil {
		sendSocketEvent(actorID, "socket_error", gin.H{"message": "通话不存在"})
		return
	}
	if isTerminalCallStatus(log.Status) {
		sendSocketEvent(req.TargetUserID, "call_end", gin.H{
			"from_user_id": actorID,
			"call_id":      log.CallID,
			"payload":      req.Payload,
		})
		return
	}

	status := string(models.SocialCallStatusCompleted)
	missedBy := uint(0)
	if log.AnsweredAt == nil {
		if actorID == log.CallerUserID {
			status = string(models.SocialCallStatusMissed)
			missedBy = log.CalleeUserID
		} else {
			status = string(models.SocialCallStatusCanceled)
		}
	}
	finalizeCallLog(&log, status, actorID, missedBy)
	sendSocketEvent(req.TargetUserID, "call_end", gin.H{
		"from_user_id": actorID,
		"call_id":      log.CallID,
		"payload":      req.Payload,
	})
}

func findCallLogForUser(callID string, userID uint) (models.SocialCallLogModel, error) {
	var log models.SocialCallLogModel
	err := global.DB.Where("call_id = ? AND (caller_user_id = ? OR callee_user_id = ?)", callID, userID, userID).Take(&log).Error
	return log, err
}

func finalizeCallLog(log *models.SocialCallLogModel, status string, operatorID uint, missedBy uint) {
	if log == nil || log.ID == 0 {
		return
	}
	now := time.Now()
	updates := map[string]any{
		"status":                status,
		"ended_at":              &now,
		"missed_by_user_id":     missedBy,
		"last_operator_user_id": operatorID,
	}
	if log.AnsweredAt != nil && status == string(models.SocialCallStatusCompleted) {
		updates["duration_sec"] = int(now.Sub(*log.AnsweredAt).Seconds())
		log.DurationSec = updates["duration_sec"].(int)
	}
	if err := global.DB.Model(log).Updates(updates).Error; err != nil {
		return
	}
	log.Status = status
	log.EndedAt = &now
	log.MissedByUserID = missedBy
	log.LastOperatorUserID = operatorID
	if duration, ok := updates["duration_sec"].(int); ok {
		log.DurationSec = duration
	}
	createCallRecordMessageIfNeeded(log)
}

func createCallRecordMessageIfNeeded(log *models.SocialCallLogModel) {
	if log == nil || log.RelatedMessageID != 0 {
		return
	}
	message := models.SocialMessageModel{
		ConversationKey:     log.ConversationKey,
		ConversationType:    string(models.SocialConversationDirect),
		SendUserID:          log.CallerUserID,
		SendUserNickName:    log.CallerNickName,
		SendUserAvatar:      log.CallerAvatar,
		ReceiveUserID:       log.CalleeUserID,
		ReceiveUserNickName: log.CalleeNickName,
		ReceiveUserAvatar:   log.CalleeAvatar,
		MsgType:             string(models.SocialMessageCall),
		Content:             formatCallRecordMessage(*log),
		RelatedCallID:       log.CallID,
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		return tx.Model(log).Update("related_message_id", message.ID).Error
	}); err != nil {
		return
	}
	log.RelatedMessageID = message.ID
	broadcastSocketEvent([]uint{log.CallerUserID, log.CalleeUserID}, "new_direct_message", message)
}

func formatCallRecordMessage(log models.SocialCallLogModel) string {
	switch log.Status {
	case string(models.SocialCallStatusRejected):
		return "语音通话已拒绝"
	case string(models.SocialCallStatusMissed):
		return "未接来电"
	case string(models.SocialCallStatusCanceled):
		return "语音通话已取消"
	case string(models.SocialCallStatusCompleted):
		return "语音通话 " + formatCallDuration(log.DurationSec)
	default:
		return "语音通话"
	}
}

func formatCallDuration(durationSec int) string {
	if durationSec < 0 {
		durationSec = 0
	}
	minute := durationSec / 60
	second := durationSec % 60
	return fmt.Sprintf("%02d:%02d", minute, second)
}

func isTerminalCallStatus(status string) bool {
	switch status {
	case string(models.SocialCallStatusRejected),
		string(models.SocialCallStatusMissed),
		string(models.SocialCallStatusCompleted),
		string(models.SocialCallStatusCanceled):
		return true
	default:
		return false
	}
}
