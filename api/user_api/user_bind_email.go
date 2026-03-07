package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/plugins/email"
	"gvb-server/utils/jwts"
	"gvb-server/utils/pwd"
	"gvb-server/utils/random"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// BindEmailRequest 绑定邮箱请求参数
type BindEmailRequest struct {
	Email    string  `json:"email" binding:"required,email" msg:"邮箱非法" swag:"description:邮箱地址"`
	Code     *string `json:"code" swag:"description:验证码，第二次请求时必填"`
	Password string  `json:"password" swag:"description:可选，新密码"`
}

// UserBindEmailView 用户绑定邮箱
// @Summary 用户绑定邮箱
// @Description 用户绑定邮箱，第一次请求发送验证码，第二次请求验证并绑定
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body BindEmailRequest true "邮箱信息"
// @Success 200 {object} res.Response{msg=string} "操作成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "用户不存在"
// @Router /api/user_bind_email [post]
func (UserApi) UserBindEmailView(c *gin.Context) {
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	// 用户绑定邮箱， 第一次输入是 邮箱
	// 后台会给这个邮箱发验证码
	var cr BindEmailRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	emailValue := strings.ToLower(strings.TrimSpace(cr.Email))

	session := sessions.Default(c)
	if cr.Code == nil {
		// 第一次，后台发验证码
		// 生成4位验证码， 将生成的验证码存入session
		code := random.Code(4)
		// 写入session
		session.Set("valid_code", code)
		session.Set("valid_email", emailValue)

		err = session.Save()
		if err != nil {
			global.Log.Error(err)
			res.FailWithMessage("session错误", c)
			return
		}
		err = email.NewCode().Send(emailValue, "你的验证码是 "+code)
		if err != nil {
			global.Log.Error(err)
			res.FailWithMessage(readableEmailSendError(err), c)
			return
		}
		res.OkWithMessage("验证码已发送，请查收", c)
		return
	}
	// 第二次，用户输入邮箱，验证码，密码
	codeValue := strings.TrimSpace(*cr.Code)
	code, _ := session.Get("valid_code").(string)
	savedEmail, _ := session.Get("valid_email").(string)
	if strings.TrimSpace(code) == "" || strings.TrimSpace(savedEmail) == "" {
		res.FailWithMessage("验证码已失效，请重新获取", c)
		return
	}
	// 校验验证码
	if strings.TrimSpace(code) != codeValue {
		res.FailWithMessage("验证码错误", c)
		return
	}
	if strings.TrimSpace(savedEmail) != emailValue {
		res.FailWithMessage("两次填写的邮箱不一致，请重新获取验证码", c)
		return
	}
	// 修改用户的邮箱
	var user models.UserModel
	err = global.DB.Take(&user, claims.UserID).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	// 第一次的邮箱，和第二次的邮箱也要做一致性校验
	updateMap := map[string]any{
		"email": emailValue,
	}
	if strings.TrimSpace(cr.Password) != "" {
		if len(strings.TrimSpace(cr.Password)) < 8 {
			res.FailWithMessage("密码强度太低", c)
			return
		}
		updateMap["password"] = pwd.HashPwd(cr.Password)
	}
	err = global.DB.Model(&user).Updates(updateMap).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("绑定邮箱失败", c)
		return
	}
	session.Delete("valid_code")
	session.Delete("valid_email")
	_ = session.Save()
	// 完成绑定
	res.OkWithMessage("邮箱绑定成功", c)
	return
}
