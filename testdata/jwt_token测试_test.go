package testdata

import (
	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/utils/jwts"
	"testing"
)

func TestToken(t *testing.T) {
	//初始化配置文件，读取里面的配置
	core.InitConf()
	//连接数据库之前初始化日志输出
	global.Log = core.InitLogger()
	var jwtPayLoad jwts.JwtPayLoad = jwts.JwtPayLoad{
		Username: "admin",
		NickName: "管理员",
		Role:     1,
		UserID:   1,
	}
	// 添加实际的测试逻辑
	token, err := jwts.GenToken(jwtPayLoad)
	if err != nil {
		global.Log.Infof("生成 Token 失败: %v", err) // 使用 Infof
	}
	global.Log.Infof("生成的 Token: %s", token) // 使用 Infof
	tokens, err := jwts.ParseToken(token)
	if err != nil {
		global.Log.Infof("解析 Token 失败: %v", err) // 使用 Infof
	}
	global.Log.Infof("解析的 Token: %v", tokens)
}
