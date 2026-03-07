package cron_ser

import (
	"time"

	"gvb-server/service/crawl_ser"

	"github.com/robfig/cron/v3"
)

func CronInit() {
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	// 创建一个支持秒级精度的cron调度器
	Cron := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	// 每秒钟执行一次文章数据同步
	Cron.AddFunc("0 0 0 * * *", SyncArticleData)
	Cron.AddFunc("*/10 * * * * *", SyncCommentData)
	// 每小时执行一次抓取任务（是否执行由配置开关控制）。
	Cron.AddFunc("0 0 * * * *", crawl_ser.SyncFengfengArticlesJob)
	Cron.Start()
}
