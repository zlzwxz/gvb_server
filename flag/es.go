package flag

import "gvb-server/models"

func EscreateIndex() {
	//调用es创建表结构
	models.ArticleModel{}.CreateIndex()
}
