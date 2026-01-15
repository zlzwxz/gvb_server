package config

// 对应setting.yml,里面的日志配置
type logger struct {
	Level          string `yaml:"level`          //日志级别
	Prefix         string `yaml:"prefix`         //日志前缀
	Director       string `yaml:"director`       //日志保存目录
	Show_line      bool   `yaml:"show_line`      //显示行号
	Log_in_console bool   `yaml:"log_in_console` //是否显示打印路径
}
