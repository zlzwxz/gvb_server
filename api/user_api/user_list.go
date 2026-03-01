package user_api

import (
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/common"
	"gvb-server/utils/desens"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
)

// UserListView 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页，非管理员会脱敏处理敏感信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认10"
// @Param sort query string false "排序方式"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.UserModel}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/users [get]
func (UserApi) UserListView(c *gin.Context) {
	// 如何判断是管理员
	claims, _ := c.MustGet("claims").(*jwts.CustomClaims)
	var page models.PageInfo
	if err := c.ShouldBindQuery(&page); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var users []models.UserModel
	list, count, _ := common.ComList(models.UserModel{}, common.Option{
		PageInfo: page,
	})
	for _, user := range list {
		if ctype.Role(claims.Role) != ctype.PermissionAdmin {
			// 管理员
			user.UserName = ""
		}
		user.Tel = desens.DesensitizationTel(user.Tel)
		user.Email = desens.DesensitizationEmail(user.Email)
		// 脱敏
		users = append(users, user)
	}

	res.OkWithList(users, count, c)
}
