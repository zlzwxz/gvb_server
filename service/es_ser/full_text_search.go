package es_ser

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/olivere/elastic/v7"
	"github.com/russross/blackfriday"
	"github.com/sirupsen/logrus"
	"gvb-server/global"
	"gvb-server/models"
	"strings"
)

type SearchData struct {
	Body  string `json:"body"`  // 正文
	Slug  string `json:"slug"`  // 包含文章的id 的跳转地址
	Title string `json:"title"` // 标题
	Key   string `json:"key"`   //文章关联id
}

// GetSearchIndexDataByContent 全文文章搜索索引创建
func GetSearchIndexDataByContent(id, title, content string) (searchDataList []SearchData) {
	dataList := strings.Split(content, "\n")
	var isCode bool = false
	var headList, bodyList []string
	var body string
	headList = append(headList, getHeader(title))
	for _, s := range dataList {
		// #{1,6}
		// 判断一下是否是代码块
		if strings.HasPrefix(s, "```") {
			isCode = !isCode
		}
		if strings.HasPrefix(s, "#") && !isCode {
			headList = append(headList, getHeader(s))
			//if strings.TrimSpace(body) != "" {
			bodyList = append(bodyList, getBody(body))
			//}
			body = ""
			continue
		}
		body += s
	}
	bodyList = append(bodyList, getBody(body))
	ln := len(headList)
	for i := 0; i < ln; i++ {
		searchDataList = append(searchDataList, SearchData{
			Title: headList[i],
			Body:  bodyList[i],
			Slug:  id + GetSlug(headList[i]),
			Key:   id,
		})
	}
	_, _ = json.Marshal(searchDataList)
	return searchDataList
}

// AsyncArticleByFullText 同步文章数据到全文搜索
func AsyncArticleByFullText(index SearchData) {
	fmt.Println(index, "11")
	//es默认有一秒钟刷新间隔，需要强制刷新，否则无法立即生效
	global.ESClient.Refresh().Index(models.FullTextModel{}.Index()).Do(context.Background())

	list, err := global.ESClient.Search().
		Index(models.FullTextModel{}.Index()). // 指定索引
		// 核心修改：用TermQuery替代MatchPhraseQuery，适配keyword类型
		Query(elastic.NewTermQuery("key", index.Key)).
		Size(1). // 只查1条结果即可判断是否存在
		Do(context.Background())

	// 增加错误处理（原代码缺失，建议补充）
	if err != nil {
		logrus.Error("ES查询失败:", err)
		return // 或根据业务逻辑处理错误
	}

	if list.TotalHits() > 0 {
		logrus.Info("索引已存在，key值为:", index.Key)
		return
	}
	logrus.Info("索引不存在，key值为:", index.Key)
	fmt.Println(err)
	// 根据文章ID、标题和内容生成全文搜索索引数据
	// GetSearchIndexDataByContent函数会将文章内容分词，生成多个搜索条目
	indexList := GetSearchIndexDataByContent(index.Key, index.Title, index.Body)
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
		global.Log.Error(err)
	}
	global.Log.Info("全文文章索引创建成功!", len(result.Succeeded()))
}

// DeleteFullTextByArticleID 删除文章数据
func DeleteFullTextByArticleID(id string) {
	boolSearch := elastic.NewTermQuery("key", id)
	res, _ := global.ESClient.
		DeleteByQuery().
		Index(models.FullTextModel{}.Index()).
		Query(boolSearch).
		Do(context.Background())
	logrus.Infof("成功删除 %d 条记录", res.Deleted)
}

func getHeader(head string) string {
	head = strings.ReplaceAll(head, "#", "")
	head = strings.ReplaceAll(head, " ", "")
	return head
}

func getBody(body string) string {
	unsafe := blackfriday.MarkdownCommon([]byte(body))
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(unsafe)))
	return doc.Text()
}

func GetSlug(slug string) string {
	return "#" + slug
}
