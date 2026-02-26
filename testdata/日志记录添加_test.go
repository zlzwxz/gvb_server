package testdata

import (
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/plugins/log_stash"
	"testing"
)

func TestLogCreate(t *testing.T) {
	core.InitConf()
	global.Log = core.InitLogger()
	global.ESClient, _ = core.EsConnect()
	// 初始化数据库连接
	global.DB = core.InitGorm()
	if global.DB == nil {
		t.Fatal("数据库连接初始化失败")
	}
	// 使用空token避免JWT解析错误
	log := log_stash.New("192.168.1.1", "")
	log.Info("测试日志")
}
