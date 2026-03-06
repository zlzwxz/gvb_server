package config

import "strconv"

// 对应setting里面的mysql配置文件
type Mysql struct {
	Host         string `yaml:"host" json:"host"`                 //数据库地址
	Port         int    `yaml:"port" json:"port"`                 //端口
	User         string `yaml:"user" json:"user"`                 //用户名
	Password     string `yaml:"password" json:"password"`         //密码
	DB           string `yaml:"db" json:"db"`                     //数据库名
	Log_level    string `yaml:"log_level" json:"log_level"`       //日志级别,debug就是全部sql，info就是info级别以上的日志，dev,release
	Config       string `yaml:"config" json:"config"`             //数据库高级配置
	MaxIdleConns int    `yaml:"maxidleconns" json:"maxidleconns"` //最大空闲连接数
	MaxOpenConns int    `yaml:"maxopenconns" json:"maxopenconns"` //最大打开连接数
	Timeout      int    `yaml:"timeout" json:"timeout"`           //连接超时时间
}

func (m *Mysql) Dsn() string {
	return m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + strconv.Itoa(m.Port) + ")/" + m.DB + "?" + m.Config
}
