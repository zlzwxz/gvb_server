package config

import "fmt"

// 对应setting.yml,里面的system配置
type System struct {
	Host string `yaml:"host` //地址
	Port int    `yaml:"port` //端口
	Env  string `yaml:"env`  //日志级别
}

func (s System) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
