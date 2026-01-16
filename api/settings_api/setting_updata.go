package settings_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/config"
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/models/res"
)

// SettingsInfoUpdateView 修改某一项的配置信息
// @Tags 配置管理
// @Summary 更新配置信息
// @Description 根据配置名称更新对应的配置信息。site: {title,subtitle,keywords...}; email: {smtp_host,smtp_port...}; qq: {app_id,app_key...}; qiniu: {access_key,secret_key...}; jwt: {sign_key,expires_time...}
// @Accept json
// @Produce json
// @Param name path string true "配置名称" Enums(site,email,qq,qiniu,jwt)
// @Success 200 {object} res.Response{}
// @Router /api/settings/{name} [put]
// @Param data body object true "配置数据"
// SettingsInfoUpdateView 修改某一项的配置信息
func (SettingsApi) SettingsInfoUpdateView(c *gin.Context) {
	var cr SettingsUri
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	switch cr.Name {
	case "site":
		var info config.SiteInfo
		err = c.ShouldBindJSON(&info)
		if err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.SiteInfo = info

	case "email":
		var info config.Email
		err = c.ShouldBindJSON(&info)
		if err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Email = info
	case "qq":
		var info config.QQ
		err = c.ShouldBindJSON(&info)
		if err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.QQ = info
	case "qiniu":
		var info config.QiNiu
		err = c.ShouldBindJSON(&info)
		if err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.QiNiu = info
	case "jwt":
		var info config.Jwt
		err = c.ShouldBindJSON(&info)
		if err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Jwt = info
	default:
		res.FailWithMessage("没有对应的配置信息", c)
		return
	}

	core.SetYaml()
	res.OkWith(c)
}
