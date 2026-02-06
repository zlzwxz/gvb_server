package testdata

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/service/es_ser"
	"testing"
)

// TestArticleSearch 测试文章全文搜索数据同步到Elasticsearch
// 该函数的主要功能是从Elasticsearch中获取所有文章数据，
// 然后生成对应的全文搜索索引数据并批量同步到全文搜索索引中
func TestArticleSearch(t *testing.T) {
	// 读取配置文件
	core.InitConf()
	// 初始化日志
	global.Log = core.InitLogger()
	// 连接es
	global.ESClient, _ = core.EsConnect()
	// 创建一个匹配所有文档的查询
	boolSearch := elastic.NewMatchAllQuery()
	// 从Elasticsearch中查询所有文章数据，最多获取1000条
	res, _ := global.ESClient.
		Search(models.ArticleModel{}.Index()). // 指定文章索引
		Query(boolSearch).                     // 设置查询条件
		Size(1000).                            // 设置查询结果数量限制
		Do(context.Background())               // 执行查询

	// 遍历查询结果中的每一篇文章
	for _, hit := range res.Hits.Hits {
		// 将查询结果转换为ArticleModel对象
		var article models.ArticleModel
		_ = json.Unmarshal(hit.Source, &article)

		//判断全文搜索索引数据是否存在
		//3. 执行查询 - 使用正确的模型和索引名称
		list, _ := global.ESClient.Search().
			Index(models.FullTextModel{}.Index()).             // 使用正确的模型和索引
			Query(elastic.NewMatchPhraseQuery("key", hit.Id)). // 使用match_phrase精确匹配短语
			Size(1).                                           // 获取更多结果便于调试
			Do(context.Background())

		if list.TotalHits() > 0 {
			logrus.Info("索引已存在")
			continue
		} // 根据文章ID、标题和内容生成全文搜索索引数据
		// GetSearchIndexDataByContent函数会将文章内容分词，生成多个搜索条目
		indexList := es_ser.GetSearchIndexDataByContent(hit.Id, article.Title, article.Content)

		// 创建批量操作请求
		bulk := global.ESClient.Bulk()
		// 遍历索引数据列表，为每个条目创建索引请求
		for _, indexData := range indexList {
			// 创建批量索引请求，指定全文搜索索引和文档数据
			req := elastic.NewBulkIndexRequest().Index(models.FullTextModel{}.Index()).Doc(indexData)
			// 将请求添加到批量操作中
			bulk.Add(req)
		}
		// 执行批量索引操作
		result, err := bulk.Do(context.Background())
		if err != nil {
			// 记录错误日志并继续处理下一篇文章
			logrus.Error(err)
			continue
		}
		// 输出同步结果
		fmt.Println(article.Title, "添加成功", "共", len(result.Succeeded()), " 条！")
	}
}
