package social_api

import (
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// SummaryView 返回好友系统顶部摘要。
func (SocialApi) SummaryView(c *gin.Context) {
	claims := getClaims(c)
	friendIDs := fetchFriendIDs(claims.UserID)
	onlineCount := 0
	for _, item := range loadPresenceMap(friendIDs, claims.UserID) {
		if item.IsOnline {
			onlineCount++
		}
	}
	var blockCount int64
	global.DB.Model(&models.UserBlockModel{}).Where("user_id = ?", claims.UserID).Count(&blockCount)
	res.OkWithData(socialSummaryResponse{
		Presence:          loadPresenceMap([]uint{claims.UserID}, claims.UserID)[claims.UserID],
		FriendCount:       len(friendIDs),
		OnlineFriendCount: onlineCount,
		BlockCount:        int(blockCount),
	}, c)
}

// FriendListView 获取好友列表，默认按在线和最近聊天排序。
func (SocialApi) FriendListView(c *gin.Context) {
	claims := getClaims(c)
	res.OkWithData(buildFriendCards(claims.UserID, strings.TrimSpace(c.Query("key"))), c)
}

// RelationView 获取与某个用户的社交关系。
func (SocialApi) RelationView(c *gin.Context) {
	var uri socialUserURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	res.OkWithData(buildRelation(claims.UserID, uri.ID), c)
}

// FollowView 关注用户。
func (SocialApi) FollowView(c *gin.Context) {
	var uri socialUserURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if uri.ID == 0 || uri.ID == claims.UserID {
		res.FailWithMessage("不能关注自己", c)
		return
	}
	var user models.UserModel
	if err := global.DB.Take(&user, uri.ID).Error; err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	blocked, blockedBy := hasBlockBetween(claims.UserID, uri.ID)
	if blocked || blockedBy {
		res.FailWithMessage("存在拉黑关系，无法关注", c)
		return
	}

	model := models.UserFollowModel{
		UserID:       claims.UserID,
		FollowUserID: uri.ID,
	}
	if err := global.DB.FirstOrCreate(&model, models.UserFollowModel{UserID: claims.UserID, FollowUserID: uri.ID}).Error; err != nil {
		res.FailWithMessage("关注失败", c)
		return
	}
	if relation := buildRelation(claims.UserID, uri.ID); relation.IsFriend {
		sendSocketEvent(uri.ID, "friend_update", map[string]any{"user_id": claims.UserID, "type": "friend"})
		sendSocketEvent(claims.UserID, "friend_update", map[string]any{"user_id": uri.ID, "type": "friend"})
	}
	res.OkWithData(buildRelation(claims.UserID, uri.ID), c)
}

// UnfollowView 取消关注。
func (SocialApi) UnfollowView(c *gin.Context) {
	var uri socialUserURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if uri.ID == 0 || uri.ID == claims.UserID {
		res.FailWithMessage("目标用户无效", c)
		return
	}
	if err := global.DB.Where("user_id = ? AND follow_user_id = ?", claims.UserID, uri.ID).Delete(&models.UserFollowModel{}).Error; err != nil {
		res.FailWithMessage("取消关注失败", c)
		return
	}
	sendSocketEvent(uri.ID, "friend_update", map[string]any{"user_id": claims.UserID, "type": "unfollow"})
	res.OkWithData(buildRelation(claims.UserID, uri.ID), c)
}
