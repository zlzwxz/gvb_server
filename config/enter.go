package config

type Config struct {
	Mysql    Mysql    `yaml:"mysql"`
	Logger   Logger   `yaml:"logger"`
	System   System   `yaml:"system"`
	Email    Email    `yaml:"email"`
	Jwt      Jwt      `yaml:"jwt"`
	SiteInfo SiteInfo `yaml:"site_info"`
	QQ       QQ       `yaml:"qq"`
	QiNiu    QiNiu    `yaml:"qiniu"`
	News     News     `yaml:"news"`
	Upload   Upload   `yaml:"upload"`
	Redis    Redis    `yaml:"redis"`
	ES       ES       `yaml:"es"`
}
