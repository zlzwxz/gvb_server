package models

import (
	"context"
	"github.com/sirupsen/logrus"
	"gvb-server/global"
)

type FullTextModel struct {
	Title string `json:"title" structs:"title"` // 文章标题
	Body  string `json:"body" structs:"body"`   // 文章内容
	Slug  string `json:"slug" structs:"slug"`   // 文章分类
	Key   string `json:"key" structs:"key"`     //文章关联id

}

func (FullTextModel) Index() string {
	return "full_text_index"
}

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
      "key":  { 
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

// IndexExists 索引是否存在
func (a FullTextModel) IndexExists() bool {
	exists, err := global.ESClient.
		IndexExists(a.Index()).
		Do(context.Background())
	if err != nil {
		logrus.Error(err.Error())
		return exists
	}
	return exists
}

// CreateIndex 创建索引
func (a FullTextModel) CreateIndex() error {
	if a.IndexExists() {
		// 有索引
		a.RemoveIndex()
	}
	// 没有索引
	// 创建索引
	createIndex, err := global.ESClient.
		CreateIndex(a.Index()).
		BodyString(a.Mapping()).
		Do(context.Background())
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

// RemoveIndex 删除索引
func (a FullTextModel) RemoveIndex() error {
	global.Log.Info("索引存在，删除索引")
	// 删除索引
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
