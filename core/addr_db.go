package core

import (
	"gvb-server/global"
	"log"

	geoip2db "github.com/cc14514/go-geoip2-db"
)

func InitAddrDB() {
	db, err := geoip2db.NewGeoipDbByStatik()
	if err != nil {
		log.Fatal(err, "addr数据库连接失败")
	} else {
		global.Log.Info("addr数据库连接成功")
	}
	global.AddrDB = db
}
