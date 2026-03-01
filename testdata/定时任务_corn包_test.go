package testdata

import (
	"fmt"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestCorn(t *testing.T) {
	// 创建一个cron调度器，使用默认配置
	//c := cron.New()

	// 创建一个支持秒级精度的cron调度器
	c := cron.New(cron.WithSeconds())

	// 示例1: 每分钟执行一次
	// 表达式格式: 秒 分 时 日 月 星期
	_, err := c.AddFunc("0 * * * * *", func() {
		fmt.Println("每分钟执行一次:", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		t.Fatalf("添加每分钟任务失败: %v", err)
	}

	// 示例2: 每5秒执行一次
	_, err = c.AddFunc("*/5 * * * * *", func() {
		fmt.Println("每5秒执行一次:", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		t.Fatalf("添加每5秒任务失败: %v", err)
	}

	// 示例3: 每天的10:30执行
	_, err = c.AddFunc("0 30 10 * * *", func() {
		fmt.Println("每天10:30执行:", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		t.Fatalf("添加每天10:30任务失败: %v", err)
	}

	// 示例4: 每周一的14:00执行
	_, err = c.AddFunc("0 0 14 * * 1", func() {
		fmt.Println("每周一14:00执行:", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		t.Fatalf("添加每周一14:00任务失败: %v", err)
	}

	// 启动cron调度器
	c.Start()

	// 测试运行20秒后停止
	timer := time.NewTimer(20 * time.Second)
	<-timer.C

	// 停止cron调度器
	c.Stop()
	fmt.Println("定时任务测试结束")
}
