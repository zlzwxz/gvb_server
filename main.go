package main

import (
	"gvb-server/core"
	"gvb-server/flag"
	"gvb-server/global"
	"gvb-server/routers"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	//初始化配置文件，读取里面的配置
	core.InitConf()
	//连接数据库之前初始化日志输出
	global.Log = core.InitLogger()
	//初始化数据库连接
	global.DB = core.InitGorm()
	//绑定参数，创建表结构
	//go run main.go -db
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
