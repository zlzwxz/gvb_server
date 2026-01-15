package config

import "strconv"

// 对应setting里面的mysql配置文件
type Mysql struct {
	Host         string `yaml: "host"`      //数据库地址
	Port         int    `yaml: "port"`      //端口
	User         string `yaml: "user"`      //用户名
	Password     string `yaml: "password"`  //密码
	DB           string `yaml: "db"`        //数据库名
	Log_level    string `yaml: "log_level"` //日志级别,debug就是全部sql，info就是info级别以上的日志，dev,release
	Config       string `yaml: "config"`    //数据库高级配置
	MaxIdleConns int    `yaml: "maxIdleConns"`
	MaxOpenConns int    `yaml: "maxOpenConns"`
	Timeout      int    `yaml: "timeout"`
}

func (m *Mysql) Dsn() string {
	return m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + strconv.Itoa(m.Port) + ")/" + m.DB + "?" + m.Config
}
