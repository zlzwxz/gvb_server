package models

import (
	"context"
	"gvb-server/global"
	"gvb-server/models/ctype"

	"github.com/goccy/go-json"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type ArticleModel struct {
	ID        string `json:"id" structs:"id"`                 // es的id
	CreatedAt string `json:"created_at" structs:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at" structs:"updated_at"` // 更新时间

	Title    string `json:"title" structs:"title"`                 // 文章标题
	Keyword  string `json:"keyword, omit(list)" structs:"keyword"` // 关键字
	Abstract string `json:"abstract" structs:"abstract"`           // 文章简介
	Content  string `json:"content, omit(list)" structs:"content"` // 文章内容

	LookCount     int `json:"look_count" structs:"look_count"`         // 浏览量
	CommentCount  int `json:"comment_count" structs:"comment_count"`   // 评论量
	DiggCount     int `json:"digg_count" structs:"digg_count"`         // 点赞量
	CollectsCount int `json:"collects_count" structs:"collects_count"` // 收藏量

	UserID       uint   `json:"user_id" structs:"user_id"`               // 用户id
	UserNickName string `json:"user_nick_name" structs:"user_nick_name"` //用户昵称
	UserAvatar   string `json:"user_avatar" structs:"user_avatar"`       // 用户头像

	Category  string `json:"category" structs:"category"`         // 文章分类
	Source    string `json:"source, omit(list)" structs:"source"` // 文章来源
	Link      string `json:"link, omit(list)" structs:"link"`     // 原文链接
	BoardID   uint   `json:"board_id" structs:"board_id"`         // 板块ID
	BoardName string `json:"board_name" structs:"board_name"`     // 板块名称

	BannerID  uint   `json:"banner_id" structs:"banner_id"`   // 文章封面id
	BannerUrl string `json:"banner_url" structs:"banner_url"` // 文章封面

	Tags ctype.Array `json:"tags" structs:"tags"` // 文章标签

	Attachments []ArticleAttachment `json:"attachments" structs:"attachments"` // 附件列表

	IsPrivate bool `json:"is_private" structs:"is_private"` // 是否私密，私密文章仅作者和管理员可见

	ReviewStatus     ctype.ArticleReviewStatus `json:"review_status" structs:"review_status"`           // 审核状态
	ReviewReason     string                    `json:"review_reason" structs:"review_reason"`           // 审核备注
	ReviewedAt       string                    `json:"reviewed_at" structs:"reviewed_at"`               // 审核时间
	ReviewerID       uint                      `json:"reviewer_id" structs:"reviewer_id"`               // 审核人ID
	ReviewerNickName string                    `json:"reviewer_nick_name" structs:"reviewer_nick_name"` // 审核人昵称

	DuplicateRate        float64 `json:"duplicate_rate" structs:"duplicate_rate"`                       // 重复率（%）
	DuplicateTargetID    string  `json:"duplicate_target_id, omit(list)" structs:"duplicate_target_id"` // 最相似文章ID
	DuplicateTargetTitle string  `json:"duplicate_target_title" structs:"duplicate_target_title"`       // 最相似文章标题
}

type ArticleAttachment struct {
	FileID uint   `json:"file_id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Size   int64  `json:"size"`
}

func (ArticleModel) Index() string {
	return "article_index"
}

func (ArticleModel) Mapping() string {
	return `
{
  "settings": {
    "index":{
      "max_result_window": "100000"
    }
  }, 
  "mappings": {
    "properties": {
      "title": { 
        "type": "text"
      },
      "keyword": { 
        "type": "keyword"
      },
      "abstract": { 
        "type": "text"
      },
      "content": { 
        "type": "text"
      },
      "look_count": {
        "type": "integer"
      },
      "comment_count": {
        "type": "integer"
      },
      "digg_count": {
        "type": "integer"
      },
      "collects_count": {
        "type": "integer"
      },
      "user_id": {
        "type": "integer"
      },
      "user_nick_name": { 
        "type": "keyword"
      },
      "user_avatar": { 
        "type": "keyword"
      },
      "category": { 
        "type": "keyword"
      },
      "source": { 
        "type": "keyword"
      },
      "link": { 
        "type": "keyword"
      },
      "board_id": {
        "type": "integer"
      },
      "board_name": {
        "type": "keyword"
      },
      "banner_id": {
        "type": "integer"
      },
      "banner_url": { 
        "type": "keyword"
      },
      "tags": { 
        "type": "keyword"
      },
      "attachments": {
        "type": "nested",
        "properties": {
          "file_id": {
            "type": "integer"
          },
          "name": {
            "type": "keyword"
          },
          "url": {
            "type": "keyword"
          },
          "size": {
            "type": "long"
          }
        }
      },
      "is_private": {
        "type": "boolean"
      },
      "review_status": {
        "type": "integer"
      },
      "review_reason": {
        "type": "text"
      },
      "reviewer_id": {
        "type": "integer"
      },
      "reviewer_nick_name": {
        "type": "keyword"
      },
      "reviewed_at": {
        "type": "date",
        "null_value": "null",
        "format": "[yyyy-MM-dd HH:mm:ss]"
      },
      "duplicate_rate": {
        "type": "float"
      },
      "duplicate_target_id": {
        "type": "keyword"
      },
      "duplicate_target_title": {
        "type": "text"
      },
      "created_at":{
        "type": "date",
        "null_value": "null",
        "format": "[yyyy-MM-dd HH:mm:ss]"
      },
      "updated_at":{
        "type": "date",
        "null_value": "null",
        "format": "[yyyy-MM-dd HH:mm:ss]"
      }
    }
  }
}
`
}

// IndexExists 索引是否存在
func (a ArticleModel) IndexExists() bool {
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
func (a ArticleModel) CreateIndex() error {
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
func (a ArticleModel) RemoveIndex() error {
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

// Create 添加的方法
func (a *ArticleModel) Create() (err error) {
	indexResponse, err := global.ESClient.Index().
		Index(a.Index()).
		BodyJson(a).Do(context.Background())
	if err != nil {
		global.Log.Error(err.Error())
		return err
	}
	a.ID = indexResponse.Id
	return nil
}

// ISExistData 是否存在该文章
func (a ArticleModel) ISExistData() bool {
	res, err := global.ESClient.
		Search(a.Index()).
		Query(elastic.NewTermQuery("keyword", a.Title)).
		Size(1).
		Do(context.Background())
	if err != nil {
		logrus.Error(err.Error())
		return false
	}
	if res.Hits.TotalHits.Value > 0 {
		return true
	}
	return false
}

// DataGetByID 是否存在该文章
func (a *ArticleModel) DataGetByID(id string) error {
	res, err := global.ESClient.
		Get().
		Id(id).
		Index(a.Index()).
		Do(context.Background())
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	err = json.Unmarshal(res.Source, a)
	return err
}
