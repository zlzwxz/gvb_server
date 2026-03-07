package config

type SiteInfo struct {
	CreatedAt   string `yaml:"created_at" json:"created_at"`
	BeiAn       string `yaml:"bei_an" json:"bei_an"`
	Title       string `yaml:"title" json:"title"`
	QQImage     string `yaml:"qq_image" json:"qq_image"`
	Version     string `yaml:"version" json:"version"`
	Email       string `yaml:"email" json:"email"`
	WechatImage string `yaml:"wechat_image" json:"wechat_image"`
	Name        string `yaml:"name" json:"name"`
	Job         string `yaml:"job" json:"job"`
	Addr        string `yaml:"addr" json:"addr"`
	Slogan      string `yaml:"slogan" json:"slogan"`
	SloganEn    string `yaml:"slogan_en" json:"slogan_en"`
	Web         string `yaml:"web" json:"web"`
	BiliBiliUrl string `yaml:"bilibili_url" json:"bilibili_url"`
	GiteeUrl    string `yaml:"gitee_url" json:"gitee_url"`
	GithubUrl   string `yaml:"github_url" json:"github_url"`
	Profile     string `yaml:"profile" json:"profile"`                                           // 个人介绍文案
	Contact     string `yaml:"contact" json:"contact"`                                           // 联系方式（手机号/微信/QQ文本）
	ServiceURL  string `yaml:"service_url" json:"service_url"`                                   // 客服地址
	AutoCrawl   bool   `yaml:"auto_crawl_fengfeng_articles" json:"auto_crawl_fengfeng_articles"` // 是否自动抓取枫枫知道文章
	CrawlerUser string `yaml:"crawler_user_name" json:"crawler_user_name"`                       // 自动抓取写入文章时使用的系统账号
	CrawlerNick string `yaml:"crawler_nick_name" json:"crawler_nick_name"`                       // 自动抓取写入文章时使用的系统昵称
}
