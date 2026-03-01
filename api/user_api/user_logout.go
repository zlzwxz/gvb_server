package user_api

import (
	"gvb-server/global"
	"gvb-server/models/res"
	"gvb-server/service"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
)

// LogoutView 用户注销登录
// @Summary 用户注销登录
// @Description 用户注销登录，将token加入黑名单
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{msg=string} "注销成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/logout [post]
func (UserApi) LogoutView(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	token := c.Request.Header.Get("token")

	err := service.ServiceApp.UserService.Logout(claims, token)

	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("注销失败", c)
		return
	}

	res.OkWithMessage("注销成功", c)

}
