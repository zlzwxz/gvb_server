package settings_api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gvb-server/global"
	"gvb-server/models/res"
	"gvb-server/service/crawl_ser"
)

// PublicSiteInfoView 公开返回站点展示信息（前台可匿名访问）。
func (SettingsApi) PublicSiteInfoView(c *gin.Context) {
	res.OkWithData(global.Config.SiteInfo, c)
}

// SyncFengfengArticlesView 手动触发一次“枫枫知道文章同步”。
func (SettingsApi) SyncFengfengArticlesView(c *gin.Context) {
	var cr struct {
		ArticleIDs    []string `json:"article_ids"`
		SyncAll       bool     `json:"sync_all"`
		IncludeUpdate bool     `json:"include_update"`
		Limit         *int     `json:"limit"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&cr); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
	}
	limit := -1
	if cr.Limit != nil {
		limit = *cr.Limit
	}
	result, err := crawl_ser.SyncFengfengArticlesWithOptions(crawl_ser.SyncArticleOptions{
		ArticleIDs:    cr.ArticleIDs,
		SyncAll:       cr.SyncAll,
		IncludeUpdate: cr.IncludeUpdate,
		Limit:         limit,
	})
	if err != nil {
		res.FailWithMessage("同步失败: "+err.Error(), c)
		return
	}
	res.OkWithData(result, c)
}

// PreviewFengfengArticlesView 仅检索最新文章数量与去重结果，不入库。
func (SettingsApi) PreviewFengfengArticlesView(c *gin.Context) {
	limit := -1
	if rawLimit, ok := c.GetQuery("limit"); ok {
		if parsed, err := strconv.Atoi(rawLimit); err == nil {
			limit = parsed
		}
	}
	result, err := crawl_ser.PreviewFengfengArticlesWithLimit(limit)
	if err != nil {
		res.FailWithMessage("检索失败: "+err.Error(), c)
		return
	}
	res.OkWithData(result, c)
}

// PreviewFengfengImagesView 仅检索最新图片候选，不入库。
func (SettingsApi) PreviewFengfengImagesView(c *gin.Context) {
	result, err := crawl_ser.PreviewFengfengImages()
	if err != nil {
		res.FailWithMessage("检索图片失败: "+err.Error(), c)
		return
	}
	res.OkWithData(result, c)
}

// SyncFengfengImagesView 手动触发一次图片素材抓取。
func (SettingsApi) SyncFengfengImagesView(c *gin.Context) {
	var cr struct {
		ImageURLs []string `json:"image_urls"`
		SyncAll   bool     `json:"sync_all"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&cr); err != nil {
			res.FailWithCode(res.ArgumentError, c)
			return
		}
	}
	result, err := crawl_ser.SyncFengfengImagesWithOptions(crawl_ser.SyncImageOptions{
		ImageURLs: cr.ImageURLs,
		SyncAll:   cr.SyncAll,
	})
	if err != nil {
		res.FailWithMessage("抓取图片失败: "+err.Error(), c)
		return
	}
	res.OkWithData(result, c)
}
