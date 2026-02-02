package core

import (
	"gvb-server/global"

	"github.com/olivere/elastic/v7"
)

func EsConnect() (*elastic.Client, error) {
	sniffOpt := elastic.SetSniff(false)
	host := global.Config.ES.URL()
	c, err := elastic.NewClient(
		elastic.SetURL(host),
		sniffOpt,
		elastic.SetBasicAuth(global.Config.ES.User, global.Config.ES.Password),
	)
	if err != nil {
		global.Log.Fatalf("连接es索引失败：%s", err.Error())
		return nil, err // 返回错误而不是致命退出
	}
	global.Log.Println("连接es索引成功")
	return c, nil
}
