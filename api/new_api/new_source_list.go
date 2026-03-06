package new_api

import (
	"strings"

	"gvb-server/global"
	"gvb-server/models/res"
	new2 "gvb-server/utils/news"

	"github.com/gin-gonic/gin"
)

// NewsSource 表示一个可以切换的资讯来源。
// 前端先拿这个列表，再把选中的 ID 传给 /api/news。
type NewsSource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Icon     string `json:"icon"`
	Category string `json:"category"`
	Enabled  bool   `json:"enabled"`
}

// NewSourceListView 获取新闻来源列表
// @Summary 获取新闻来源列表
// @Description 返回可切换的热搜榜来源，供前端构建资讯来源面板
// @Tags 新闻管理
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=[]NewsSource} "获取成功"
// @Router /api/news/sources [get]
func (NewApi) NewSourceListView(c *gin.Context) {
	sources := getNewsSources()
	if len(sources) == 0 {
		res.FailWithMessage("暂无可用的新闻来源", c)
		return
	}
	res.OkWithData(sources, c)
}

func getNewsSources() []NewsSource {
	categoryResp := new2.GetNewsId()
	seen := map[string]struct{}{}
	sources := make([]NewsSource, 0, len(categoryResp.Data))
	for _, item := range categoryResp.Data {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		sources = append(sources, NewsSource{
			ID:       item.ID,
			Name:     item.Name,
			Type:     item.Type,
			Icon:     item.Icon,
			Category: item.Category,
			Enabled:  isNewsSourceEnabled(item.ID, item.Name),
		})
	}
	return sources
}

func filterEnabledNewsSources(sources []NewsSource) []NewsSource {
	idSet := getEnabledNewsSourceIDSet()
	nameSet := getEnabledNewsSourceNameSet()
	if len(idSet) == 0 && len(nameSet) == 0 {
		return sources
	}
	list := make([]NewsSource, 0, len(sources))
	for _, source := range sources {
		if isNewsSourceEnabled(source.ID, source.Name) {
			list = append(list, source)
		}
	}
	return list
}

func isNewsSourceEnabled(id string, name string) bool {
	idSet := getEnabledNewsSourceIDSet()
	if len(idSet) > 0 {
		_, ok := idSet[strings.TrimSpace(id)]
		return ok
	}

	nameSet := getEnabledNewsSourceNameSet()
	if len(nameSet) == 0 {
		return true
	}
	_, ok := nameSet[normalizeNewsSourceName(name)]
	return ok
}

func getEnabledNewsSourceIDSet() map[string]struct{} {
	sourceIDs := global.Config.News.EnabledSourceIDs
	if len(sourceIDs) == 0 {
		return nil
	}
	result := make(map[string]struct{}, len(sourceIDs))
	for _, id := range sourceIDs {
		normalized := strings.TrimSpace(id)
		if normalized == "" {
			continue
		}
		result[normalized] = struct{}{}
	}
	return result
}

func getEnabledNewsSourceNameSet() map[string]struct{} {
	sourceNames := global.Config.News.EnabledSourceNames
	if len(sourceNames) == 0 {
		return nil
	}
	result := make(map[string]struct{}, len(sourceNames))
	for _, name := range sourceNames {
		normalized := normalizeNewsSourceName(name)
		if normalized == "" {
			continue
		}
		result[normalized] = struct{}{}
	}
	return result
}

func normalizeNewsSourceName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func pickNewsSourceID(requestID string, sources []NewsSource) string {
	if requestID != "" {
		for _, source := range sources {
			if source.ID == requestID {
				return requestID
			}
		}
	}
	if len(sources) == 0 {
		return ""
	}
	return sources[0].ID
}
