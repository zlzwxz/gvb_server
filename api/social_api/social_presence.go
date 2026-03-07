package social_api

import (
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// PresenceUpdateView 更新当前用户在线状态与个性签名。
func (SocialApi) PresenceUpdateView(c *gin.Context) {
	var cr socialPresenceRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	presence := savePresence(claims.UserID, cr)
	sendSocketEvent(claims.UserID, "self_presence", presence)
	notifyPresenceChange(claims.UserID)
	res.OkWithData(presence, c)
}
