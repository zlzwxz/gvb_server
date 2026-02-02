package global

import (
	"gvb-server/config"

	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 存放公共全局变量
var (
	//配置文件
	Config *config.Config
	//数据库连接
	DB *gorm.DB
	//日志打印
	Log *logrus.Logger
	//email
	Email *config.Email
	//jwt
	Jwt *config.Jwt
	//site_info
	SiteInfo *config.SiteInfo
	//qq
	QQ *config.QQ
	//qiniu
	QiNiu *config.QiNiu
	//mysql日志全局变量
	MysqlLog logger.Interface
	//redis初始话
	Redis *redis.Client
	//es全局变量
	ESClient *elastic.Client
)
