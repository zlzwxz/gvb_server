package user_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/user_ser"

	"github.com/gin-gonic/gin"
)

// UserCreateRequest 创建用户请求参数
type UserCreateRequest struct {
	NickName string     `json:"nick_name" binding:"required" msg:"请输入昵称" swag:"description:昵称"`              // 昵称
	UserName string     `json:"user_name" binding:"required" msg:"请输入用户名" swag:"description:用户名"`            // 用户名
	Password string     `json:"password" binding:"required" msg:"请输入密码" swag:"description:密码"`               // 密码
	Role     ctype.Role `json:"role" binding:"required" msg:"请选择权限" swag:"description:权限 1 管理员 2 普通用户 3 游客"` // 权限  1 管理员  2 普通用户  3 游客
}

// UserCreateView 创建用户
// @Summary 创建用户
// @Description 创建新用户，支持设置昵称、用户名、密码和权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body UserCreateRequest true "用户创建参数"
// @Success 200 {object} res.Response{msg=string} "创建成功"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/user_create [post]
func (UserApi) UserCreateView(c *gin.Context) {
	var cr UserCreateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	err := user_ser.UserService{}.CreateUser(cr.UserName, cr.NickName, cr.Password, cr.Role, "", c.ClientIP())
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage(err.Error(), c)
		return
	}
	res.OkWithMessage(fmt.Sprintf("用户%s创建成功!", cr.UserName), c)
	return
}
