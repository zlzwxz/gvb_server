package settings_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models/res"
)

type SettingsUri struct {
	Name string `uri:"name"`
}

// SettingsInfoView 显示某一项的配置信息
// @Tags 配置管理
// @Summary 获取配置信息
// @Description 根据配置名称获取对应的配置信息
// @Accept json
// @Produce json
// @Param name path string true "配置名称" Enums(site, email, qq, qiniu, jwt)
// @Success 200 {object} res.Response{data=interface{}}
// @Router /api/settings/{name} [get]
// SettingsInfoView 显示某一项的配置信息
func (SettingsApi) SettingsInfoView(c *gin.Context) {

	var cr SettingsUri
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	switch cr.Name {
	case "site":
		res.OkWithData(global.Config.Logger, c)
	case "email":
		res.OkWithData(global.Config.Email, c)
	case "qq":
		res.OkWithData(global.Config.QQ, c)
	case "qiniu":
		res.OkWithData(global.Config.QiNiu, c)
	case "jwt":
		res.OkWithData(global.Config.Jwt, c)
	default:
		res.FailWithMessage("没有对应的配置信息", c)
	}
}
