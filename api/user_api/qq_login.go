package user_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"
	"gvb-server/plugins/qq"
	"gvb-server/utils"
	"gvb-server/utils/jwts"
	"gvb-server/utils/pwd"
	"gvb-server/utils/random"

	"github.com/gin-gonic/gin"
)

// QQLoginView QQ登录
// @Summary QQ登录
// @Description 通过QQ授权码进行登录，如果用户不存在则自动注册
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param code query string true "QQ授权码"
// @Success 200 {object} res.Response{data=string} "登录成功，返回token"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 500 {object} res.Response "登录失败"
// @Router /api/qq_login [post]
func (UserApi) QQLoginView(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		res.FailWithMessage("没有code", c)
		return
	}
	qqInfo, err := qq.NewQQLogin(code)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	openID := qqInfo.OpenID
	// 根据openID判断用户是否存在
	var user models.UserModel
	//拿到ip地址
	ip, addr := utils.GetAddrByGin(c)
	//添加日志记录
	log := log_stash.NewLogByGin(c)
	err = global.DB.Take(&user, "token = ?", openID).Error
	if err != nil {
		// 不存在，就注册
		hashPwd := pwd.HashPwd(random.RandString(16))
		user = models.UserModel{
			NickName:   qqInfo.Nickname,
			UserName:   openID,  // qq登录，邮箱+密码
			Password:   hashPwd, // 随机生成16位密码
			Avatar:     qqInfo.Avatar,
			Addr:       addr, // 根据ip算地址
			Token:      openID,
			IP:         ip,
			Role:       ctype.PermissionUser,
			SignStatus: ctype.SignQQ,
		}
		err = global.DB.Create(&user).Error
		if err != nil {
			global.Log.Error(err)

			log.Info(fmt.Sprintf("用户名%v 注册失败", openID))
			res.FailWithMessage("注册失败", c)
			return
		}

	}
	// 登录操作
	token, err := jwts.GenToken(jwts.JwtPayLoad{
		NickName: user.NickName,
		Role:     int(user.Role),
		UserID:   user.ID,
	})
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("token生成失败", c)
		return
	}

	//登录成功写入登录数据
	global.DB.Create(models.LoginDataModel{
		UserID:    user.ID,
		IP:        ip,
		NickName:  user.NickName,
		Token:     token,
		Device:    "",
		Addr:      addr,
		LoginType: int(ctype.SignQQ),
	})

	res.OkWithData(token, c)
}
