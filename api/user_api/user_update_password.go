package user_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"gvb-server/utils/pwd"
)

type UpdatePasswordRequest struct {
	OldPwd string `json:"old_pwd"` // 旧密码
	Pwd    string `json:"pwd"`     // 新密码
}

// UserUpdatePassword 修改登录人的id
func (UserApi) UserUpdatePassword(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	var cr UpdatePasswordRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	var user models.UserModel
	err := global.DB.Take(&user, claims.UserID).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	// 判断密码是否一致
	if !pwd.ComparePasswords(user.Password, cr.OldPwd) {
		res.FailWithMessage("密码错误", c)
		return
	}
	fmt.Println(user.Password, "123")
	hashPwd := pwd.HashPwd(cr.Pwd)
	err = global.DB.Model(&user).Update("password", hashPwd).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("密码修改失败", c)
		return
	}
	res.OkWithMessage("密码修改成功", c)
	return
}
