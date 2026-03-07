package main

import (
	"gvb-server/core"
	_ "gvb-server/docs" // swag init 生成后的文档包，匿名导入后 Swagger 页面才能读取到注释元数据。
	"gvb-server/flag"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/plugins/log_stash"
	"gvb-server/routers"
	"gvb-server/service/board_ser"
	"gvb-server/service/cron_ser"
)

// @title gvb_server API 文档
// @version 1.0
// @description gvb_server 服务端接口文档，覆盖博客、用户、评论、消息、聊天、配置等核心能力。
// @host 127.0.0.1:8080
// @BasePath /
func main() {
	// `main` 是整个后端服务的启动总入口。
	// 你可以把它理解成后端的“总装配厂”：
	// 1. 先读配置；
	// 2. 再初始化日志、数据库、Redis、ES 等依赖；
	// 3. 如果本次是命令行任务模式，就执行完任务直接退出；
	// 4. 如果是 Web 服务模式，就初始化定时任务和路由，最后启动 HTTP 服务。

	// 初始化顺序尽量固定：先配置，再日志，再数据库和外部依赖。
	// 这样后续任一步失败时，日志系统已经可以正常输出错误信息。
	core.InitConf()
	global.Log = core.InitLogger()
	global.DB = core.InitGorm()

	// 如果数据库可用，就执行运行时迁移。
	// 运行时迁移的含义是：让程序在启动时自动补齐需要的数据表结构。
	if global.DB != nil {
		if err := global.DB.AutoMigrate(
			&models.BoardModel{},
			&models.BannerModel{},
			&models.AnnouncementModel{},
			&models.ArticleReportModel{},
			&models.UserFollowModel{},
			&models.UserBlockModel{},
			&models.UserPresenceModel{},
			&models.SocialFileModel{},
			&models.SocialGroupModel{},
			&models.SocialGroupMemberModel{},
			&models.SocialMessageModel{},
			&models.SocialConversationReadModel{},
			&models.SocialCallLogModel{},
			&models.UserSpacePostModel{},
			&models.UserSpaceMessageModel{},
			&log_stash.LogStashModel{},
		); err != nil {
			global.Log.Errorf("运行时迁移失败: %v", err)
		}

		// 启动时顺带确保默认板块存在，避免前台首页或发帖功能没有基础板块可用。
		if err := board_ser.EnsureDefaultBoards(); err != nil {
			global.Log.Errorf("初始化默认板块失败: %v", err)
		}
	}

	// 初始化其他外部依赖。
	global.Redis = core.ConnectRedis()
	global.ESClient, _ = core.EsConnect()
	core.InitAddrDB()
	defer global.AddrDB.Close()

	// 命令行模式用于建表、创建用户等一次性任务；如果命中命令模式，就不再启动 Web 服务。
	option := flag.Parse()
	if flag.IsWebStop(option) {
		flag.SwitchOption(option)
		return
	}

	// 启动定时任务，比如文章同步、统计同步等后台任务。
	cron_ser.CronInit()

	// 初始化 Gin 路由引擎。
	router := routers.InitRouter()
	addr := global.Config.System.Addr()
	global.Log.Infof("gvb_server 项目启动服务成功，端口：%s", addr)
	global.Log.Infof("gvb_server 项目 api 文档运行在: http://%s/swagger/index.html#/", addr)
	if err := router.Run(addr); err != nil {
		global.Log.Fatalf("%v", err)
	}
}
