package elasticsearch

import (
	"context"
	"fmt"
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/models"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

func TestCommArticleList(t *testing.T) {
	// 读取配置文件
	core.InitConf()
	// 初始化日志
	global.Log = core.InitLogger()
	// 连接es
	global.ESClient, _ = core.EsConnect()
	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(elastic.NewMultiMatchQuery("node", "title", "abstract", "content")).
		Highlight(elastic.NewHighlight().Field("title")).
		Size(100).
		Do(context.Background())
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, hit := range result.Hits.Hits {
		fmt.Println(string(hit.Source))
		fmt.Println(hit.Highlight)
	}
}
