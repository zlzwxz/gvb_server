package settings_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models/res"
)

// SettingsUri 统一接收路由里的配置名称。
type SettingsUri struct {
	Name string `uri:"name"`
}

// SettingsInfoView 获取某一项配置信息。
// @Tags 配置管理
// @Summary 获取配置信息
// @Description 根据配置名称获取对应的系统配置，支持 settings.yaml 的主要分组。
// @Accept json
// @Produce json
// @Param name path string true "配置名称" Enums(system,mysql,logger,email,jwt,qq,qiniu,upload,redis,es,site,site_info,news)
// @Success 200 {object} res.Response{data=interface{}} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 404 {object} res.Response "没有对应的配置信息"
// @Router /api/settings/{name} [get]
func (SettingsApi) SettingsInfoView(c *gin.Context) {
	var cr SettingsUri
	if err := c.ShouldBindUri(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	switch cr.Name {
	case "site", "site_info":
		res.OkWithData(global.Config.SiteInfo, c)
	case "system":
		res.OkWithData(global.Config.System, c)
	case "mysql":
		res.OkWithData(global.Config.Mysql, c)
	case "logger":
		res.OkWithData(global.Config.Logger, c)
	case "email":
		res.OkWithData(global.Config.Email, c)
	case "qq":
		res.OkWithData(global.Config.QQ, c)
	case "qiniu":
		res.OkWithData(global.Config.QiNiu, c)
	case "jwt":
		res.OkWithData(global.Config.Jwt, c)
	case "upload":
		res.OkWithData(global.Config.Upload, c)
	case "redis":
		res.OkWithData(global.Config.Redis, c)
	case "es":
		res.OkWithData(global.Config.ES, c)
	case "news":
		res.OkWithData(global.Config.News, c)
	default:
		res.FailWithMessage("没有对应的配置信息", c)
	}
}
