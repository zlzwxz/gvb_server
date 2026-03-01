package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// UserRole 用户权限请求参数
type UserRole struct {
	Role     ctype.Role `json:"role" binding:"required,oneof=1 2 3 4" msg:"权限参数错误" swag:"description:权限 1 管理员 2 普通用户 3 游客 4 禁用用户"`
	NickName string     `json:"nick_name" swag:"description:用户昵称"` // 防止用户昵称非法，管理员有能力修改
	UserID   uint       `json:"user_id" binding:"required" msg:"用户id错误" swag:"description:用户ID"`
}

// UserUpdateRoleView 用户权限变更
// @Summary 用户权限变更
// @Description 修改用户的权限和昵称
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body UserRole true "权限信息"
// @Success 200 {object} res.Response{msg=string} "修改成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "用户不存在"
// @Router /api/user_role [put]
func (UserApi) UserUpdateRoleView(c *gin.Context) {
	var cr UserRole
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	var user models.UserModel
	err := global.DB.Take(&user, cr.UserID).Error
	if err != nil {
		res.FailWithMessage("用户id错误，用户不存在", c)
		return
	}
	err = global.DB.Model(&user).Updates(map[string]any{
		"role":      cr.Role,
		"nick_name": cr.NickName,
	}).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("修改权限失败", c)
		return
	}
	res.OkWithMessage("修改权限成功", c)
}
