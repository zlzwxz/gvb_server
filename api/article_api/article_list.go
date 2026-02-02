package article_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"

	"github.com/gin-gonic/gin"
	"github.com/liu-cn/json-filter/filter"
)

func (ArticleApi) ArticleListView(c *gin.Context) {
	var cr models.PageInfo
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	option := es_ser.Option{
		PageInfo: cr,
		Fields:   []string{"title", "content"},
		Tag:      cr.Tag,
	}
	list, count, err := es_ser.CommList(option)
	if err != nil {
		global.Log.Error(err)
		res.OkWithMessage("查询失败", c)
		return
	}

	data := filter.Omit("list", list)
	if data == nil {
		list = make([]models.ArticleModel, 0)
		res.OkWithList(list, int64(count), c)
		return
	}

	// 安全地进行类型断言
	_list, ok := data.(map[string]interface{})
	if !ok {
		// 如果不是期望的类型，使用原始数据
		res.OkWithList(filter.Omit("list", list), int64(count), c)
		return
	}

	// 检查 map 是否为空
	if len(_list) == 0 {
		list = make([]models.ArticleModel, 0)
		res.OkWithList(list, int64(count), c)
		return
	}

	res.OkWithList(data, int64(count), c)

	//res.OkWithList(filter.Omit("list", list), int64(count), c)
}
