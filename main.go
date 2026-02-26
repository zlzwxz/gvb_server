package main

import (
	"gvb-server/core"
	_ "gvb-server/docs" // swag init生成后的docs路径
	"gvb-server/flag"
	"gvb-server/global"
	"gvb-server/routers"
)

// @title gvb_server API文档
// @version 1.0
// @description gvb_server API文档
// @host 127.0.0.01:8080
// @BasePath /
func main() {
	//初始化配置文件，读取里面的配置
	core.InitConf()
	//连接数据库之前初始化日志输出
	global.Log = core.InitLogger()
	//初始化数据库连接
	global.DB = core.InitGorm()
	//连接redis
	global.Redis = core.ConnectRedis()
	//连接es索引
	global.ESClient, _ = core.EsConnect()
	//绑定参数，创建表结构
	//go run main.go -db 创建数据库结构
	//go run main.go -user user 创建用户 admin为管理员 user为普通用户
	option := flag.Parse()
	if flag.IsWebStop(option) {
		flag.SwitchOption(option)
		return
	}
	//启动路由
	router := routers.InitRouter()
	addr := global.Config.System.Addr()
	global.Log.Infof("启动服务成功，端口：%s", addr)
	err := router.Run(addr)
	if err != nil {
		global.Log.Fatalf(err.Error())
	}
}
