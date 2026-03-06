package es_ser

import (
	"gvb-server/models"

	"github.com/olivere/elastic/v7"
)

// Option 描述文章 ES 查询的通用参数。
type Option struct {
	models.PageInfo
	Fields []string      // 需要参与搜索的字段
	Tag    string        // 标签筛选
	Query  elastic.Query // 自定义 ES Query
}

// GetForm 计算 ES 分页查询起始偏移量。
func (o *Option) GetForm() int {
	if o.Page == 0 {
		o.Page = 1
	}
	if o.Limit == 0 {
		o.Limit = 10
	}
	return (o.Page - 1) * o.Limit
}
