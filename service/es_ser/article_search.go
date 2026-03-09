package es_ser

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/service/redis_ser"
	"strings"
)

// CommList 文章搜索
func CommList(option Option) (list []models.ArticleModel, count int, err error) {

	boolSearch := elastic.NewBoolQuery()
	if option.Query != nil {
		boolSearch.Must(option.Query)
	}

	if option.Key != "" {
		keywordQuery := elastic.NewBoolQuery().MinimumNumberShouldMatch(1)
		if len(option.Fields) > 0 {
			keywordQuery.Should(
				elastic.NewMultiMatchQuery(option.Key, option.Fields...),
			)
		}
		text := strings.TrimSpace(option.Key)
		if text != "" {
			like := "*" + text + "*"
			keywordQuery.Should(
				elastic.NewWildcardQuery("tags", like),
				elastic.NewWildcardQuery("board_name", like),
				elastic.NewWildcardQuery("category", like),
				elastic.NewTermQuery("keyword", text),
			)
		}
		boolSearch.Must(keywordQuery)
	}
	if option.Tag != "" {
		boolSearch.Must(elastic.NewTermQuery("tags", option.Tag))
	}
	type SortField struct {
		Field     string
		Ascending bool
	}
	sortField := SortField{
		Field:     "created_at",
		Ascending: false, // 从小到大  从大到小
	}
	if option.Sort != "" {
		_list := strings.Split(option.Sort, " ")
		if len(_list) == 2 && (_list[1] == "desc" || _list[1] == "asc") {
			sortField.Field = _list[0]
			if _list[1] == "desc" {
				sortField.Ascending = false
			}
			if _list[1] == "asc" {
				sortField.Ascending = true
			}
		}
	}
	//redis查询是否有点赞数量
	diggInfo := redis_ser.NewDigg().GetInfo()
	//redis查看浏览量
	lookInfo := redis_ser.NewArticleLook().GetInfo()
	//redis查看评论数
	commentInfo := redis_ser.NewCommentCount().GetInfo()
	searchService := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(boolSearch).
		From(option.GetForm()).
		Sort(sortField.Field, sortField.Ascending).
		Size(option.Limit)

	if strings.TrimSpace(option.Key) != "" {
		searchService = searchService.Highlight(
			elastic.NewHighlight().
				RequireFieldMatch(false).
				Fields(
					elastic.NewHighlighterField("title").NumOfFragments(0),
					elastic.NewHighlighterField("abstract").FragmentSize(160).NumOfFragments(1).NoMatchSize(120),
					elastic.NewHighlighterField("content").FragmentSize(200).NumOfFragments(1).NoMatchSize(160),
				),
		)
	}

	res, err := searchService.Do(context.Background())
	if err != nil {
		return
	}
	count = int(res.Hits.TotalHits.Value) //搜索到结果总条数
	demoList := []models.ArticleModel{}
	for _, hit := range res.Hits.Hits {
		var model models.ArticleModel
		data, err := hit.Source.MarshalJSON()
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		err = json.Unmarshal(data, &model)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if strings.TrimSpace(option.Key) != "" {
			title, ok := hit.Highlight["title"]
			if ok && len(title) > 0 {
				model.Title = title[0]
			}
			abstract, ok := hit.Highlight["abstract"]
			if ok && len(abstract) > 0 {
				model.Abstract = abstract[0]
			} else {
				content, ok := hit.Highlight["content"]
				if ok && len(content) > 0 {
					model.Abstract = content[0]
				}
			}
		}
		//添加点赞数
		model.ID = hit.Id
		diggCount := diggInfo[hit.Id]
		model.DiggCount += diggCount
		//添加浏览量
		lookCount := lookInfo[hit.Id]
		model.LookCount += lookCount
		//添加评论数
		model.CommentCount += commentInfo[hit.Id]

		demoList = append(demoList, model)
	}

	return demoList, count, err
}

func CommDetail(id string) (model models.ArticleModel, err error) {
	res, err := global.ESClient.Get().
		Index(models.ArticleModel{}.Index()).
		Id(id).
		Do(context.Background())
	if err != nil {
		return
	}
	err = json.Unmarshal(res.Source, &model)
	if err != nil {
		return
	}
	model.ID = res.Id
	//文章详情需要添加浏览量
	model.LookCount += redis_ser.NewArticleLook().Get(res.Id)
	return
}

func CommDetailByKeyword(key string) (model models.ArticleModel, err error) {
	res, err := global.ESClient.Search().
		Index(models.ArticleModel{}.Index()).
		Query(elastic.NewTermQuery("keyword", key)).
		Size(1).
		Do(context.Background())
	if err != nil {
		return
	}
	if res.Hits.TotalHits.Value == 0 {
		return model, errors.New("文章不存在")
	}
	hit := res.Hits.Hits[0]

	err = json.Unmarshal(hit.Source, &model)
	if err != nil {
		return
	}
	model.ID = hit.Id
	return
}

// ArticleUpdate 更新文章收藏数
func ArticleUpdate(id string, data map[string]any) error {
	_, err := global.ESClient.
		Update().
		Index(models.ArticleModel{}.Index()).
		Id(id).
		Doc(data).
		Do(context.Background())
	logrus.Info("更新文章收藏数成功")
	return err
}
