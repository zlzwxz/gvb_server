package settings_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/config"
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/models/res"
	"gvb-server/service/crawl_ser"
)

// SettingsInfoUpdateView 更新某一项配置信息。
// @Tags 配置管理
// @Summary 更新配置信息
// @Description 根据配置名称更新对应的配置信息，支持 settings.yaml 的主要分组。
// @Accept json
// @Produce json
// @Param name path string true "配置名称" Enums(system,mysql,logger,email,jwt,qq,qiniu,upload,redis,es,site,site_info,news)
// @Param data body object true "配置数据"
// @Success 200 {object} res.Response "更新成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 404 {object} res.Response "没有对应的配置信息"
// @Router /api/settings/{name} [put]
func (SettingsApi) SettingsInfoUpdateView(c *gin.Context) {
	var cr SettingsUri
	if err := c.ShouldBindUri(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	switch cr.Name {
	case "site", "site_info":
		var info config.SiteInfo
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		if info.AutoCrawl {
			if _, err := crawl_ser.EnsureCrawlerAccount(); err != nil {
				res.FailWithMessage("系统员账号创建失败: "+err.Error(), c)
				return
			}
		}
		global.Config.SiteInfo = info
	case "system":
		var info config.System
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.System = info
	case "mysql":
		var info config.Mysql
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Mysql = info
	case "logger":
		var info config.Logger
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Logger = info
	case "email":
		var info config.Email
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Email = info
	case "qq":
		var info config.QQ
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.QQ = info
	case "qiniu":
		var info config.QiNiu
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.QiNiu = info
	case "jwt":
		var info config.Jwt
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Jwt = info
	case "upload":
		var info config.Upload
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Upload = info
	case "redis":
		var info config.Redis
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.Redis = info
	case "es":
		var info config.ES
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.ES = info
	case "news":
		var info config.News
		if err := c.ShouldBindJSON(&info); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
		global.Config.News = info
	default:
		res.FailWithMessage("没有对应的配置信息", c)
		return
	}

	// 把内存中的最新配置重新写回 yaml，这样重启后仍然保留刚才的修改结果。
	core.SetYaml()
	res.OkWith(c)
}
