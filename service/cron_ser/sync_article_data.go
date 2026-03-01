package cron_ser

import (
	"context"
	"encoding/json"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/service/redis_ser"

	"github.com/olivere/elastic/v7"
)

// SyncArticleData redis同步文章数据到es
func SyncArticleData() {
	//1.查询es里面的所有文章数据
	result, err := global.ESClient.Search(models.ArticleModel{}.Index()).
		Query(elastic.NewMatchAllQuery()).
		Size(10000).
		Do(context.Background())
	if err != nil {
		global.Log.Errorf("查询es文章数据失败，%v", err)
		return
	}
	//2.查询redis里面的所有文章数据
	diggInfo := redis_ser.NewDigg().GetInfo()
	lookInfo := redis_ser.NewArticleLook().GetInfo()
	commentInfo := redis_ser.NewCommentCount().GetInfo()
	for _, hit := range result.Hits.Hits {
		var article models.ArticleModel
		//解析es文章数据json成对象
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			global.Log.Errorf("解析es文章数据失败，%v", err)
			continue
		}
		digg := diggInfo[hit.Id]
		look := lookInfo[hit.Id]
		comment := commentInfo[hit.Id]
		//3.计算新的点赞数、浏览数、评论数
		NewDigg := digg + article.DiggCount
		NewLook := look + article.LookCount
		NewComment := comment + article.CommentCount
		//4.需要判断是否有更新
		if digg == 0 && look == 0 && comment == 0 {
			//如果没有更新，直接跳过
			global.Log.Infof("文章:%s 没有更新", article.Title)
			continue
		}
		//5.更新文章数据
		_, err = global.ESClient.Update().Index(models.ArticleModel{}.Index()).
			Id(hit.Id).
			Doc(map[string]interface{}{
				"look_count":    NewLook,
				"comment_count": NewComment,
				"digg_count":    NewDigg,
			}).
			Do(context.Background())
		if err != nil {
			global.Log.Errorf("更新es文章数据失败，%v", err)
			continue
		}
		global.Log.Infof("文章:%s 更新成功 点赞数:%d 浏览数:%d 评论数:%d", article.Title, NewDigg, NewLook, NewComment)
	}
	//6.更新redis数据后清除缓存
	redis_ser.NewDigg().Clear()
	redis_ser.NewArticleLook().Clear()
	redis_ser.NewCommentCount().Clear()
}
