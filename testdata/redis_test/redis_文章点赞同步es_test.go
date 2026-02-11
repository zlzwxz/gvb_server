package redis

import (
	"context"
	"encoding/json"
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/service/redis_ser"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

func TestRedisDiggTest(t *testing.T) {
	// 读取配置文件
	core.InitConf()
	// 初始化日志
	global.Log = core.InitLogger()
	global.Redis = core.ConnectRedis()
	global.ESClient, _ = core.EsConnect()

	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(elastic.NewMatchAllQuery()).
		Size(10000).
		Do(context.Background())
	if err != nil {
		logrus.Error(err)
		return
	}

	diggInfo := redis_ser.NewDigg().GetInfo()
	lookInfo := redis_ser.NewArticleLook().GetInfo()
	for _, hit := range result.Hits.Hits {
		var article models.ArticleModel
		err = json.Unmarshal(hit.Source, &article)

		digg := diggInfo[hit.Id]
		look := lookInfo[hit.Id]
		newDigg := article.DiggCount + digg
		newLook := article.LookCount + look
		if article.DiggCount == newDigg && article.LookCount == newLook {
			logrus.Info(article.Title, "点赞数和浏览量无变化")
			continue
		}
		_, err := global.ESClient.
			Update().
			Index(models.ArticleModel{}.Index()).
			Id(hit.Id).
			Doc(map[string]int{
				"digg_count": newDigg,
				"look_count": newLook,
			}).
			Do(context.Background())
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		logrus.Info(article.Title, "点赞数和浏览量同步成功， 点赞数", newDigg, "浏览量", newLook)
	}
	redis_ser.NewDigg().Clear()
	redis_ser.NewArticleLook().Clear()
}
