package config

type Config struct {
	Mysql  Mysql  `yaml:"mysql"`
	Logger logger `yaml:"logger"`
	System System `yaml:"system"`
}

// 对应setting里面的mysql配置文件
type Mysql struct {
	Host      string `yaml: "host"`      //数据库地址
	Port      int    `yaml: "port"`      //端口
	User      string `yaml: "user"`      //用户名
	Password  string `yaml: "password"`  //密码
	Db        string `yaml: "db"`        //数据库名
	Log_level string `yaml:" log_level"` //日志级别,debug就是全部sql，info就是info级别以上的日志，dev,release
}

// 对应setting.yml,里面的日志配置
type logger struct {
	Level          string `yaml:"level`          //日志级别
	Prefix         string `yaml:"prefix`         //日志前缀
	Director       string `yaml:"director`       //日志保存目录
	Show_line      string `yaml:"show_line`      //显示行号
	Log_in_console string `yaml:"log_in_console` //是否显示打印路径
}

// 对应setting.yml,里面的system配置
type System struct {
	Host string `yaml:"host` //地址
	Port string `yaml:"port` //端口
	Env  string `yaml:"env`  //日志级别
}
