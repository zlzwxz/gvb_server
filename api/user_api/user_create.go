package user_api

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"
	"gvb-server/service/user_ser"
	"gvb-server/utils/jwts"
)

// UserCreateRequest 创建用户请求参数
type UserCreateRequest struct {
	NickName string     `json:"nick_name" binding:"required" msg:"请输入昵称" swag:"description:昵称"`
	UserName string     `json:"user_name" binding:"required" msg:"请输入用户名" swag:"description:用户名"`
	Email    string     `json:"email" binding:"omitempty,email" msg:"邮箱格式错误" swag:"description:邮箱"`
	Code     *string    `json:"code" swag:"description:注册验证码，普通用户注册时必填"`
	Password string     `json:"password" binding:"required" msg:"请输入密码" swag:"description:密码"`
	Role     ctype.Role `json:"role" binding:"required" msg:"请选择权限" swag:"description:权限 1 管理员 2 普通用户 3 游客"`
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

	cr.Email = strings.ToLower(strings.TrimSpace(cr.Email))
	if len(strings.TrimSpace(cr.Password)) < 8 {
		res.FailWithMessage("密码长度至少 8 位", c)
		return
	}
	isAdminRequest := isAdminCreator(c)
	if !isAdminRequest {
		cr.Role = ctype.PermissionUser
		if err := verifyRegisterEmailCode(cr.Email, cr.Code); err != nil {
			res.FailWithMessage(err.Error(), c)
			return
		}
	}

	err := user_ser.UserService{}.CreateUser(cr.UserName, cr.NickName, cr.Password, cr.Role, cr.Email, c.ClientIP())
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage(err.Error(), c)
		return
	}

	if !isAdminRequest {
		clearRegisterCodeEntry(cr.Email)
	}
	res.OkWithMessage(fmt.Sprintf("用户%s创建成功!", cr.UserName), c)
}

func verifyRegisterEmailCode(emailValue string, code *string) error {
	if strings.TrimSpace(emailValue) == "" {
		return fmt.Errorf("注册必须填写邮箱")
	}
	if code == nil || strings.TrimSpace(*code) == "" {
		return fmt.Errorf("请输入邮箱验证码")
	}

	entry, ok := getRegisterCodeEntry(emailValue)
	if !ok {
		return fmt.Errorf("验证码已失效，请重新获取")
	}
	if strings.TrimSpace(*code) != strings.TrimSpace(entry.Code) {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

func isAdminCreator(c *gin.Context) bool {
	token := strings.TrimSpace(c.GetHeader("token"))
	if token == "" {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}
	if token == "" {
		return false
	}

	claims, err := jwts.ParseToken(token)
	if err != nil || redis_ser.CheckLogout(token) {
		return false
	}
	return claims.Role == int(ctype.PermissionAdmin)
}
