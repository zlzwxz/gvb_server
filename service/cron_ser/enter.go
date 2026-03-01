package cron_ser

import (
	"time"

	"github.com/robfig/cron/v3"
)

func CronInit() {
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	// 创建一个支持秒级精度的cron调度器
	Cron := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	// 每秒钟执行一次文章数据同步
	Cron.AddFunc("0 0 0 * * *", SyncArticleData)
	Cron.AddFunc("*/10 * * * * *", SyncCommentData)
	Cron.Start()
}
