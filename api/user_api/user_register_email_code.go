package user_api

import (
	"fmt"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/plugins/email"
	"gvb-server/utils/random"

	"github.com/gin-gonic/gin"
)

const (
	registerCodeExpireMinutes = 10
	registerCodeCooldownSec   = 60
)

// RegisterEmailCodeRequest 注册邮箱验证码请求。
type RegisterEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email" msg:"邮箱非法" swag:"description:邮箱地址"`
}

// UserRegisterEmailCodeView 发送注册验证码。
// @Summary 发送注册邮箱验证码
// @Description 注册前发送邮箱验证码，验证码有效期 10 分钟
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body RegisterEmailCodeRequest true "邮箱信息"
// @Success 200 {object} res.Response{msg=string} "发送成功"
// @Failure 400 {object} res.Response "请求错误"
// @Router /api/user_register_email_code [post]
func (UserApi) UserRegisterEmailCodeView(c *gin.Context) {
	var cr RegisterEmailCodeRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	emailValue := strings.ToLower(strings.TrimSpace(cr.Email))
	var existed models.UserModel
	if err := global.DB.Take(&existed, "email = ?", emailValue).Error; err == nil {
		res.FailWithMessage("邮箱已被注册", c)
		return
	}

	if entry, ok := getRegisterCodeEntry(emailValue); ok {
		elapsed := int(time.Since(entry.SentAt).Seconds())
		if elapsed < registerCodeCooldownSec {
			remain := registerCodeCooldownSec - elapsed
			if remain < 1 {
				remain = 1
			}
			res.FailWithMessage(fmt.Sprintf("请求过于频繁，请 %d 秒后再试", remain), c)
			return
		}
	}

	code := random.Code(6)
	setRegisterCodeEntry(emailValue, code, registerCodeExpireMinutes*time.Minute)

	content := fmt.Sprintf("你的注册验证码是 %s，%d 分钟内有效。", code, registerCodeExpireMinutes)
	if err := email.NewCode().Send(emailValue, content); err != nil {
		clearRegisterCodeEntry(emailValue)
		global.Log.Error(err)
		res.FailWithMessage(readableEmailSendError(err), c)
		return
	}

	res.OkWithMessage("验证码已发送，请查收邮箱", c)
}
