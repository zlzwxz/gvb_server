package config

type Upload struct {
	Path     string `yaml:"path" json:"path"`         //读取配置文件图片里面的地址
	Max_Size int64  `yaml:"max_size" json:"max_size"` //读取配置文件里面的图片大小
}
