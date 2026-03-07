package social_api

import (
	"fmt"
	"sort"
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

type socialDirectMessageRequest struct {
	RevUserID uint   `json:"rev_user_id" binding:"required"`
	Content   string `json:"content"`
	MsgType   string `json:"msg_type"`
	FileID    uint   `json:"file_id"`
}

// ConversationListView 合并返回单聊和群聊会话列表。
func (SocialApi) ConversationListView(c *gin.Context) {
	claims := getClaims(c)
	readMap := fetchReadMap(claims.UserID)
	friendSet := map[uint]struct{}{}
	for _, item := range fetchFriendIDs(claims.UserID) {
		friendSet[item] = struct{}{}
	}

	conversationMap := map[string]*socialConversationItem{}

	var directMessages []models.SocialMessageModel
	global.DB.Where(
		"conversation_type = ? AND (send_user_id = ? OR receive_user_id = ?)",
		string(models.SocialConversationDirect), claims.UserID, claims.UserID,
	).Order("id desc").Find(&directMessages)
	for _, item := range directMessages {
		current, ok := conversationMap[item.ConversationKey]
		targetUserID := item.SendUserID
		targetName := item.SendUserNickName
		targetAvatar := item.SendUserAvatar
		if item.SendUserID == claims.UserID {
			targetUserID = item.ReceiveUserID
			targetName = item.ReceiveUserNickName
			targetAvatar = item.ReceiveUserAvatar
		}
		if !ok {
			_, isFriend := friendSet[targetUserID]
			current = &socialConversationItem{
				ConversationKey:   item.ConversationKey,
				ConversationType:  item.ConversationType,
				UserID:            targetUserID,
				Title:             targetName,
				Avatar:            targetAvatar,
				LatestMessage:     messagePreview(item),
				LatestMessageType: item.MsgType,
				LatestMessageID:   item.ID,
				LatestAt:          item.CreatedAt,
				IsFriend:          isFriend,
			}
			conversationMap[item.ConversationKey] = current
		}
		if item.SendUserID != claims.UserID && item.ID > readMap[item.ConversationKey] {
			current.UnreadCount++
		}
	}

	groupIDList := getCurrentUserGroupIDs(claims.UserID)
	groupMap := map[uint]models.SocialGroupModel{}
	if len(groupIDList) > 0 {
		var groups []models.SocialGroupModel
		global.DB.Where("id IN ?", groupIDList).Find(&groups)
		for _, item := range groups {
			groupMap[item.ID] = item
		}
		memberCountMap := map[uint]int{}
		var countRows []struct {
			GroupID uint
			Count   int64
		}
		global.DB.Model(&models.SocialGroupMemberModel{}).
			Select("group_id, count(1) as count").
			Where("group_id IN ?", groupIDList).
			Group("group_id").
			Scan(&countRows)
		for _, row := range countRows {
			memberCountMap[row.GroupID] = int(row.Count)
		}

		var groupMessages []models.SocialMessageModel
		global.DB.Where("conversation_type = ? AND group_id IN ?", string(models.SocialConversationGroup), groupIDList).
			Order("id desc").Find(&groupMessages)
		for _, item := range groupMessages {
			current, ok := conversationMap[item.ConversationKey]
			group := groupMap[item.GroupID]
			if !ok {
				current = &socialConversationItem{
					ConversationKey:   item.ConversationKey,
					ConversationType:  item.ConversationType,
					GroupID:           item.GroupID,
					Title:             group.Name,
					Avatar:            group.Avatar,
					LatestMessage:     messagePreview(item),
					LatestMessageType: item.MsgType,
					LatestMessageID:   item.ID,
					LatestAt:          item.CreatedAt,
					MemberCount:       memberCountMap[item.GroupID],
				}
				conversationMap[item.ConversationKey] = current
			}
			if item.SendUserID != claims.UserID && item.ID > readMap[item.ConversationKey] {
				current.UnreadCount++
			}
		}
	}

	userIDs := make([]uint, 0)
	for _, item := range conversationMap {
		if item.UserID > 0 {
			userIDs = append(userIDs, item.UserID)
		}
	}
	presenceMap := loadPresenceMap(userIDs, claims.UserID)
	for _, item := range conversationMap {
		if item.UserID == 0 {
			continue
		}
		presence := presenceMap[item.UserID]
		item.IsOnline = presence.IsOnline
		item.PresenceMode = presence.Mode
		item.PresenceText = presence.StatusText
		item.LastActiveAt = presence.LastActiveAt
	}

	list := make([]socialConversationItem, 0, len(conversationMap))
	for _, item := range conversationMap {
		list = append(list, *item)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].LatestAt.After(list[j].LatestAt) })
	res.OkWithData(list, c)
}

// DirectMessageListView 获取单聊消息，并自动标记已读。
func (SocialApi) DirectMessageListView(c *gin.Context) {
	var cr socialDirectMessageQuery
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if cr.UserID == 0 || cr.UserID == claims.UserID {
		res.FailWithMessage("私信对象无效", c)
		return
	}
	blocked, blockedBy := hasBlockBetween(claims.UserID, cr.UserID)
	if blocked || blockedBy {
		res.FailWithMessage("存在拉黑关系，无法查看会话", c)
		return
	}

	key := buildDirectConversationKey(claims.UserID, cr.UserID)
	var list []models.SocialMessageModel
	if err := global.DB.Where("conversation_key = ?", key).Order("id asc").Find(&list).Error; err != nil {
		res.FailWithMessage("获取私信记录失败", c)
		return
	}
	if len(list) > 0 {
		upsertConversationRead(claims.UserID, key, list[len(list)-1].ID)
	}
	res.OkWithData(buildDirectMessageResponses(claims.UserID, cr.UserID, list), c)
}

// DirectMessageCreateView 发送单聊消息。
func (SocialApi) DirectMessageCreateView(c *gin.Context) {
	var cr socialDirectMessageRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if cr.RevUserID == 0 || cr.RevUserID == claims.UserID {
		res.FailWithMessage("私信对象无效", c)
		return
	}
	relation := buildRelation(claims.UserID, cr.RevUserID)
	if !relation.CanDirectMessage {
		res.FailWithMessage("存在拉黑关系，无法发送私信", c)
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

	var sender, receiver models.UserModel
	if err := global.DB.Take(&sender, claims.UserID).Error; err != nil {
		res.FailWithMessage("发送人不存在", c)
		return
	}
	if err := global.DB.Take(&receiver, cr.RevUserID).Error; err != nil {
		res.FailWithMessage("接收人不存在", c)
		return
	}

	message := models.SocialMessageModel{
		ConversationKey:     buildDirectConversationKey(claims.UserID, cr.RevUserID),
		ConversationType:    string(models.SocialConversationDirect),
		SendUserID:          sender.ID,
		SendUserNickName:    sender.NickName,
		SendUserAvatar:      sender.Avatar,
		ReceiveUserID:       receiver.ID,
		ReceiveUserNickName: receiver.NickName,
		ReceiveUserAvatar:   receiver.Avatar,
		MsgType:             msgType,
		Content:             content,
	}
	if msgType == string(models.SocialMessageFile) {
		if !relation.CanSendFile {
			res.FailWithMessage("仅好友之间支持发送文件", c)
			return
		}
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
		res.FailWithMessage("发送私信失败", c)
		return
	}
	sendSocketEvent(claims.UserID, "new_direct_message", message)
	sendSocketEvent(receiver.ID, "new_direct_message", message)
	res.OkWithData(buildDirectMessageResponses(claims.UserID, cr.RevUserID, []models.SocialMessageModel{message})[0], c)
}
