package flag

import sys_flag "flag"

type Option struct {
	DB   bool
	User string
}

// Parse 解析命令行参数
func Parse() Option {
	db := sys_flag.Bool("db", false, "初始化数据库")
	user := sys_flag.String("user", "", "创建用户")
	// 解析命令行参数写入注册的flag里
	sys_flag.Parse()
	return Option{
		DB:   *db,
		User: *user,
	}
}

// IsWebStop 是否停止web项目
func IsWebStop(option Option) bool {
	if option.DB || option.User != "" {
		return true
	}
	return false
}

// SwitchOption 根据命令执行不同的函数
func SwitchOption(option Option) {
	if option.DB {
		Makemigrations()
	}
	if option.User == "admin" || option.User == "user" {
		CreateUser(option.User)
	}
}
