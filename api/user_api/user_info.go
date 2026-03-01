package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
)

// UserInfoView 获取用户信息
// @Summary 获取用户信息
// @Description 根据token获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=models.UserModel} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "用户不存在"
// @Router /api/user_info [get]
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
