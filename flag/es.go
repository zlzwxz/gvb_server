package flag

import "gvb-server/models"

func EscreateIndex() {
	//调用es创建文章表结构
	models.ArticleModel{}.CreateIndex()
	//创建es文章全文搜索表结构
	models.FullTextModel{}.CreateIndex()
}
