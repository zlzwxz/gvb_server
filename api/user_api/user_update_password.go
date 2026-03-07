package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"gvb-server/utils/pwd"
	"strings"

	"github.com/gin-gonic/gin"
)

// UpdatePasswordRequest 修改密码请求参数
type UpdatePasswordRequest struct {
	UserID uint   `json:"user_id" swag:"description:重置目标用户ID，仅管理员可用"`
	OldPwd string `json:"old_pwd" swag:"description:旧密码，普通用户修改自己密码时必填"`               // 旧密码
	Pwd    string `json:"pwd" binding:"required" msg:"请输入新密码" swag:"description:新密码"` // 新密码
}

// UserUpdatePassword 修改用户密码
// @Summary 修改用户密码
// @Description 修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body UpdatePasswordRequest true "密码信息"
// @Success 200 {object} res.Response{msg=string} "修改成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "用户不存在"
// @Router /api/user_password [put]
func (UserApi) UserUpdatePassword(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	var cr UpdatePasswordRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	targetUserID := claims.UserID
	isAdminReset := claims.Role == int(ctype.PermissionAdmin) && cr.UserID != 0 && cr.UserID != claims.UserID
	if isAdminReset {
		targetUserID = cr.UserID
	}

	var user models.UserModel
	err := global.DB.Take(&user, targetUserID).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	if !isAdminReset {
		if strings.TrimSpace(cr.OldPwd) == "" {
			res.FailWithMessage("请输入旧密码", c)
			return
		}
		// 判断密码是否一致
		if !pwd.ComparePasswords(user.Password, cr.OldPwd) {
			res.FailWithMessage("密码错误", c)
			return
		}
	}
	if len(strings.TrimSpace(cr.Pwd)) < 8 {
		res.FailWithMessage("新密码长度至少 8 位", c)
		return
	}
	hashPwd := pwd.HashPwd(cr.Pwd)
	if hashPwd == "" {
		res.FailWithMessage("密码加密失败", c)
		return
	}
	err = global.DB.Model(&user).Update("password", hashPwd).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("密码修改失败", c)
		return
	}
	if isAdminReset {
		res.OkWithMessage("密码重置成功", c)
		return
	}
	res.OkWithMessage("密码修改成功", c)
	return
}
