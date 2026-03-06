package config

// News 资讯榜单展示配置。
// enabled_source_names 为空时表示前台默认展示全部来源。
type News struct {
	EnabledSourceNames []string `yaml:"enabled_source_names" json:"enabled_source_names"`
	EnabledSourceIDs   []string `yaml:"enabled_source_ids" json:"enabled_source_ids"`
}
