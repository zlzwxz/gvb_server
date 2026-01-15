package config

type Config struct {
	Mysql    Mysql    `yaml:"mysql"`
	Logger   logger   `yaml:"logger"`
	System   System   `yaml:"system"`
	Email    Email    `yaml:"email"`
	Jwt      Jwt      `yaml:"jwt"`
	SiteInfo SiteInfo `yaml:"site_info"`
	QQ       QQ       `yaml:"qq"`
	QiNiu    QiNiu    `yaml:"qiniu"`
	Upload   Upload   `yaml:"upload"`
}
