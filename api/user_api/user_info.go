package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
)

// UserInfoView 用户信息
func (UserApi) UserInfoView(c *gin.Context) {
	//根据token获取用户信息
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	var userInfo models.UserModel
	err := global.DB.Take(&userInfo, claims.UserID).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	res.OkWithData(filter.Select("info", userInfo), c)

}
