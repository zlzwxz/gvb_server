package settings_api

// SettingsApi 聚合系统配置读取与更新相关的 HTTP 处理函数。
type SettingsApi struct {
	// SettingInfo 预留给需要嵌套调用配置 API 的场景；当前主要保持兼容性。
	SettingInfo *SettingsApi
}
