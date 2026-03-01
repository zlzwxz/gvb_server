package user_api

import (
	"gvb-server/global"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// QQLoginLinkView 获取qq登录的跳转链接
func (UserApi) QQLoginLinkView(c *gin.Context) {
	path := global.Config.QQ.GetPath()
	if path == "" {
		res.FailWithMessage("未配置qq登录地址", c)
		return
	}
	res.OkWithData(path, c)
	return
}
