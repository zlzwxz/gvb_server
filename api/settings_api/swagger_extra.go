package settings_api

import (
	"gvb-server/config"
	"gvb-server/models/res"
	"gvb-server/service/crawl_ser"
	"gvb-server/service/es_ser"
)

var (
	_ = res.Response{}
	_ = config.SiteInfo{}
	_ = crawl_ser.SyncResult{}
	_ = es_ser.ESIndexSummary{}
)

type settingsSyncFengfengArticlesRequest struct {
	ArticleIDs    []string `json:"article_ids" example:"article_demo_001,article_demo_002"`
	SyncAll       bool     `json:"sync_all" example:"false"`
	IncludeUpdate bool     `json:"include_update" example:"true"`
	Limit         int      `json:"limit" example:"20"`
}

type settingsSyncFengfengImagesRequest struct {
	ImageURLs []string `json:"image_urls" example:"https://static.example.com/demo-1.jpg,https://static.example.com/demo-2.jpg"`
	SyncAll   bool     `json:"sync_all" example:"false"`
}

// PublicSiteInfoViewDoc 获取前台可公开使用的站点信息。
// @Tags 配置管理
// @Summary 获取公开站点配置
// @Description 返回前台首页、备案、站点标题等公开展示配置。该接口无需登录，适合站点初始化时直接调用。
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=config.SiteInfo} "获取成功"
// @Router /api/settings/public/site_info [get]
func PublicSiteInfoViewDoc() {}

// PreviewFengfengArticlesViewDoc 预览枫枫文章同步结果。
// @Tags 配置管理
// @Summary 预览枫枫文章同步
// @Description 仅抓取候选文章和重复统计，不写入数据库。适合后台先看有哪些文章会被新增或识别为重复。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param limit query int false "扫描上限，-1=默认，0=全量，正数=指定条数"
// @Success 200 {object} res.Response{data=crawl_ser.PreviewResult} "预览成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/site_info/sync_fengfeng_preview [get]
func PreviewFengfengArticlesViewDoc() {}

// SyncFengfengArticlesViewDoc 执行枫枫文章同步。
// @Tags 配置管理
// @Summary 同步枫枫文章
// @Description 根据文章 ID、是否全量、是否允许更新等参数，把枫枫知道文章抓取并写入当前系统。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body settingsSyncFengfengArticlesRequest false "同步参数，不传时按默认策略执行"
// @Success 200 {object} res.Response{data=crawl_ser.SyncResult} "同步成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/site_info/sync_fengfeng [post]
func SyncFengfengArticlesViewDoc() {}

// PreviewFengfengImagesViewDoc 预览枫枫图片同步结果。
// @Tags 配置管理
// @Summary 预览枫枫图片同步
// @Description 只返回图片候选、重复数量和扫描统计，不真正落库存储。适合后台挑选需要抓取的图片。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=crawl_ser.PreviewImageResult} "预览成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/site_info/sync_fengfeng_images_preview [get]
func PreviewFengfengImagesViewDoc() {}

// SyncFengfengImagesViewDoc 执行枫枫图片同步。
// @Tags 配置管理
// @Summary 同步枫枫图片
// @Description 可以传指定图片 URL 列表，也可以直接一键同步预览结果中的全部新图片候选。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body settingsSyncFengfengImagesRequest false "同步参数，不传时按默认策略执行"
// @Success 200 {object} res.Response{data=crawl_ser.SyncImageResult} "同步成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/site_info/sync_fengfeng_images [post]
func SyncFengfengImagesViewDoc() {}

// ESIndexListViewDoc 获取 ES 索引列表。
// @Tags 配置管理
// @Summary 获取 ES 索引列表
// @Description 返回后台可见索引、文档数以及该索引是否支持页面导入导出。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=[]es_ser.ESIndexSummary} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/es/indices [get]
func ESIndexListViewDoc() {}

// ESIndexExportViewDoc 导出 ES 索引数据。
// @Tags 配置管理
// @Summary 导出 ES 索引
// @Description 导出指定索引为 JSON 文件。当前后台页面主要用于备份文章索引和排查全文索引内容。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param index query string true "索引名称，例如 article_index"
// @Success 200 {file} binary "导出的 JSON 文件"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/es/export [get]
func ESIndexExportViewDoc() {}

// ESIndexImportViewDoc 导入 ES 索引数据。
// @Tags 配置管理
// @Summary 导入 ES 索引
// @Description 上传导出的 JSON 文件并执行页面导入。当前页面只支持导入 article_index，并会按文章创建逻辑重新校验数据。
// @Accept multipart/form-data
// @Produce json
// @Param token header string true "token"
// @Param index formData string true "索引名称，当前仅支持 article_index"
// @Param file formData file true "要导入的 JSON 文件"
// @Success 200 {object} res.Response{data=es_ser.ArticleImportResult} "导入成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/settings/es/import [post]
func ESIndexImportViewDoc() {}
