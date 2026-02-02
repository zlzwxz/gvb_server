package flag

import sys_flag "flag"

type Option struct {
	DB   bool
	User string //-user admin 创建管理员用户 -user user 创建普通用户
	ES   string //-es article 创建文章索引
}

// Parse 解析命令行参数
func Parse() Option {
	db := sys_flag.Bool("db", false, "初始化数据库")
	user := sys_flag.String("user", "", "创建用户")
	es := sys_flag.String("es", "", "创建es索引")
	// 解析命令行参数写入注册的flag里
	sys_flag.Parse()
	return Option{
		DB:   *db,
		User: *user,
		ES:   *es,
	}
}

// IsWebStop 是否停止web项目
func IsWebStop(option Option) bool {
	if option.DB || option.User != "" || option.ES != "" {
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
	if option.ES == "create" {
		EscreateIndex()
	}
}
