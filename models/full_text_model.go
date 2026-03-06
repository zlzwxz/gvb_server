package models

import (
	"context"

	"github.com/sirupsen/logrus"
	"gvb-server/global"
)

// FullTextModel 定义全文检索索引里的文档结构。
// 这里不是数据库表，而是写入 Elasticsearch 的文档模型。
type FullTextModel struct {
	Title string `json:"title" structs:"title"` // 文章标题
	Body  string `json:"body" structs:"body"`   // 文章正文分词片段
	Slug  string `json:"slug" structs:"slug"`   // 标题 slug，用于匹配目录和搜索结果
	Key   string `json:"key" structs:"key"`     // 关联的文章 ID
}

// Index 返回全文搜索索引名。
func (FullTextModel) Index() string {
	return "full_text_index"
}

// Mapping 返回全文搜索索引的 mapping 定义。
func (FullTextModel) Mapping() string {
	return `
{
  "settings": {
    "index":{
      "max_result_window": "100000"
    }
  },
  "mappings": {
    "properties": {
      "key": {
        "type": "keyword"
      },
      "title": {
        "type": "text"
      },
      "body": {
        "type": "keyword"
      },
      "slug": {
        "type": "text"
      }
    }
  }
}
`
}

// IndexExists 判断全文索引是否已经存在。
func (a FullTextModel) IndexExists() bool {
	exists, err := global.ESClient.IndexExists(a.Index()).Do(context.Background())
	if err != nil {
		logrus.Error(err.Error())
		return exists
	}
	return exists
}

// CreateIndex 重新创建全文索引。
func (a FullTextModel) CreateIndex() error {
	if a.IndexExists() {
		a.RemoveIndex()
	}
	createIndex, err := global.ESClient.CreateIndex(a.Index()).BodyString(a.Mapping()).Do(context.Background())
	if err != nil {
		global.Log.Error("创建索引失败")
		global.Log.Error(err.Error())
		return err
	}
	if !createIndex.Acknowledged {
		global.Log.Error("创建失败")
		return err
	}
	global.Log.Infof("索引 %s 创建成功", a.Index())
	return nil
}

// RemoveIndex 删除全文索引。
func (a FullTextModel) RemoveIndex() error {
	global.Log.Info("索引存在，删除索引")
	indexDelete, err := global.ESClient.DeleteIndex(a.Index()).Do(context.Background())
	if err != nil {
		global.Log.Error("删除索引失败")
		global.Log.Error(err.Error())
		return err
	}
	if !indexDelete.Acknowledged {
		global.Log.Error("删除索引失败")
		return err
	}
	global.Log.Info("索引删除成功")
	return nil
}
