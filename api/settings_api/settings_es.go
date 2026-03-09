package settings_api

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
)

type esIndexQuery struct {
	Index string `form:"index"`
}

// ESIndexListView 返回可供后台选择的 ES 索引列表。
func (SettingsApi) ESIndexListView(c *gin.Context) {
	list, err := es_ser.ListVisibleIndices()
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	res.OkWithData(list, c)
}

// ESIndexExportView 导出后台选中的 ES 索引数据。
func (SettingsApi) ESIndexExportView(c *gin.Context) {
	var query esIndexQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	indexName := strings.TrimSpace(query.Index)
	if indexName == "" {
		res.FailWithMessage("请选择要导出的索引", c)
		return
	}

	data, err := es_ser.ExportIndexPayloadJSON(indexName)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	fileName := buildESExportFileName(indexName)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fileName, url.QueryEscape(fileName)))
	c.Data(200, "application/json; charset=utf-8", data)
}

// ESIndexImportView 导入 ES 导出文件。
// 当前页面导入只支持 article_index，并且会按文章创建规则逐条校验。
func (SettingsApi) ESIndexImportView(c *gin.Context) {
	indexName := strings.TrimSpace(c.PostForm("index"))
	fileHeader, err := c.FormFile("file")
	if err != nil {
		res.FailWithMessage("请先选择要导入的 JSON 文件", c)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		res.FailWithMessage("读取上传文件失败", c)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		res.FailWithMessage("读取上传文件内容失败", c)
		return
	}

	var claims *jwts.CustomClaims
	if _claims, ok := c.Get("claims"); ok {
		claims, _ = _claims.(*jwts.CustomClaims)
	}

	result, err := es_ser.ImportArticleIndexPayload(data, indexName, claims)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	msg := fmt.Sprintf("导入完成：共 %d 条，成功 %d 条，失败 %d 条", result.Total, result.Success, result.Failed)
	res.Ok(result, msg, c)
}

func buildESExportFileName(indexName string) string {
	safeName := strings.ReplaceAll(strings.TrimSpace(indexName), " ", "_")
	if safeName == "" {
		safeName = "es_index"
	}
	return fmt.Sprintf("%s_%s.json", safeName, time.Now().Format("20060102_150405"))
}
