package user_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"
	"gvb-server/utils"
	"gvb-server/utils/jwts"
	"gvb-server/utils/pwd"

	"github.com/gin-gonic/gin"
)

// EmailLoginRequest 邮箱登录请求参数
type EmailLoginRequest struct {
	UserName string `json:"user_name" binding:"required" msg:"请输入用户名" swag:"description:用户名或邮箱"`
	Password string `json:"password" binding:"required" msg:"请输入密码" swag:"description:密码"`
}

// EmailLoginView 邮箱登录
// @Summary 邮箱登录
// @Description 通过用户名/邮箱和密码进行登录，返回token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body EmailLoginRequest true "登录参数"
// @Success 200 {object} res.Response{data=string} "登录成功，返回token"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "用户名或密码错误"
// @Router /api/email_login [post]
func (UserApi) EmailLoginView(c *gin.Context) {
	var cr EmailLoginRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	var userModel models.UserModel

	//添加日志记录
	log := log_stash.NewLogByGin(c)
	ip, addr := utils.GetAddrByGin(c)
	err = global.DB.Take(&userModel, "user_name = ? or email = ?", cr.UserName, cr.UserName).Error
	if err != nil {
		// 没找到
		global.Log.Warn("用户名不存在")
		log.Info(fmt.Sprintf("用户名%v不存在", cr.UserName))
		res.FailWithMessage("用户名或密码错误", c)
		return
	}
	// 校验密码
	isCheck := pwd.ComparePasswords(userModel.Password, cr.Password)
	if !isCheck {
		global.Log.Warn("用户名密码错误")
		log.Info(fmt.Sprintf("用户名%v密码错误", cr.Password))
		res.FailWithMessage("用户名或密码错误", c)
		return
	}
	// 登录成功，生成token
	token, err := jwts.GenToken(jwts.JwtPayLoad{
		NickName: userModel.NickName,
		Role:     int(userModel.Role),
		UserID:   userModel.ID,
	})
	if err != nil {
		global.Log.Error(err)
		log.Info(fmt.Sprintf("token生成失败"))
		res.FailWithMessage("token生成失败", c)
		return
	}
	log.Info(fmt.Sprintf("用户名:%v  登录成功", cr.UserName))

	// 创建结构体变量，然后传递地址给Create方法
	loginData := models.LoginDataModel{
		UserID:    userModel.ID,
		IP:        ip,
		NickName:  userModel.NickName,
		Token:     token,
		Device:    "",
		Addr:      addr,
		LoginType: int(ctype.SignEmail), // 显式类型转换
	}
	global.DB.Create(&loginData) // 传递地址

	res.OkWithData(token, c)

}
