package user_api

import (
	"gvb-server/global"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// QQLoginLinkView 获取qq登录的跳转链接
// @Summary 获取QQ登录跳转链接
// @Description 获取QQ登录的跳转链接地址
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=string} "获取成功"
// @Failure 500 {object} res.Response "未配置QQ登录地址"
// @Router /api/qq_login_path [get]
func (UserApi) QQLoginLinkView(c *gin.Context) {
	path := global.Config.QQ.GetPath()
	if path == "" {
		res.FailWithMessage("未配置qq登录地址", c)
		return
	}
	res.OkWithData(path, c)
	return
}
