package main

import (
	"gvb-server/core"
	_ "gvb-server/docs" // swag init 生成后的文档包，匿名导入后 Swagger 页面才能读取到注释元数据。
	"gvb-server/flag"
	"gvb-server/global"
	"gvb-server/routers"
	"gvb-server/service/cron_ser"
)

// @title gvb_server API 文档
// @version 1.0
// @description gvb_server 服务端接口文档，覆盖博客、用户、评论、消息、聊天、配置等核心能力。
// @host 127.0.0.1:8080
// @BasePath /
func main() {
	// 初始化顺序尽量固定：先配置，再日志，再数据库和外部依赖。
	// 这样后续任一步失败时，日志系统已经可以正常输出错误信息。
	core.InitConf()
	global.Log = core.InitLogger()
	global.DB = core.InitGorm()
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

	cron_ser.CronInit()
	router := routers.InitRouter()
	addr := global.Config.System.Addr()
	global.Log.Infof("gvb_server 项目启动服务成功，端口：%s", addr)
	global.Log.Infof("gvb_server 项目 api 文档运行在: http://%s/swagger/index.html#/", addr)
	if err := router.Run(addr); err != nil {
		global.Log.Fatalf(err.Error())
	}
}
