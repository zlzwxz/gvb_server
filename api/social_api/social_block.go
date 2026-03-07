package social_api

import (
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

type socialBlockRequest struct {
	Reason string `json:"reason"`
}

// BlockListView 获取自己的黑名单。
func (SocialApi) BlockListView(c *gin.Context) {
	claims := getClaims(c)
	var blocks []models.UserBlockModel
	if err := global.DB.Where("user_id = ?", claims.UserID).Order("created_at desc").Find(&blocks).Error; err != nil {
		res.FailWithMessage("获取黑名单失败", c)
		return
	}
	userIDs := make([]uint, 0, len(blocks))
	for _, item := range blocks {
		userIDs = append(userIDs, item.BlockUserID)
	}
	userMap := loadUserMap(userIDs)
	result := make([]map[string]any, 0, len(blocks))
	for _, item := range blocks {
		user := userMap[item.BlockUserID]
		result = append(result, map[string]any{
			"id":         item.ID,
			"user_id":    item.BlockUserID,
			"nick_name":  user.NickName,
			"user_name":  user.UserName,
			"avatar":     user.Avatar,
			"reason":     item.Reason,
			"created_at": item.CreatedAt,
		})
	}
	res.OkWithData(result, c)
}

// BlockView 拉黑用户。
func (SocialApi) BlockView(c *gin.Context) {
	var uri socialUserURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var cr socialBlockRequest
	_ = c.ShouldBindJSON(&cr)
	claims := getClaims(c)
	if uri.ID == 0 || uri.ID == claims.UserID {
		res.FailWithMessage("不能拉黑自己", c)
		return
	}
	var user models.UserModel
	if err := global.DB.Take(&user, uri.ID).Error; err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	reason := strings.TrimSpace(cr.Reason)
	if len([]rune(reason)) > 80 {
		reason = string([]rune(reason)[:80])
	}
	model := models.UserBlockModel{
		UserID:      claims.UserID,
		BlockUserID: uri.ID,
		Reason:      reason,
	}
	if err := global.DB.FirstOrCreate(&model, models.UserBlockModel{UserID: claims.UserID, BlockUserID: uri.ID}).Error; err != nil {
		res.FailWithMessage("拉黑失败", c)
		return
	}
	global.DB.Where("(user_id = ? AND follow_user_id = ?) OR (user_id = ? AND follow_user_id = ?)", claims.UserID, uri.ID, uri.ID, claims.UserID).Delete(&models.UserFollowModel{})
	sendSocketEvent(uri.ID, "friend_update", map[string]any{"user_id": claims.UserID, "type": "blocked"})
	res.OkWithData(buildRelation(claims.UserID, uri.ID), c)
}

// UnblockView 取消拉黑。
func (SocialApi) UnblockView(c *gin.Context) {
	var uri socialUserURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if err := global.DB.Where("user_id = ? AND block_user_id = ?", claims.UserID, uri.ID).Delete(&models.UserBlockModel{}).Error; err != nil {
		res.FailWithMessage("移出黑名单失败", c)
		return
	}
	sendSocketEvent(uri.ID, "friend_update", map[string]any{"user_id": claims.UserID, "type": "unblocked"})
	res.OkWithData(buildRelation(claims.UserID, uri.ID), c)
}
