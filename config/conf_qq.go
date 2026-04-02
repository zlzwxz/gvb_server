package config

import (
	"net/url"
	"strings"
)

type QQ struct {
	AppID    string `json:"app_id" yaml:"app_id"`
	Key      string `json:"key" yaml:"key"`
	Redirect string `json:"redirect" yaml:"redirect"` // 登录之后的回调地址
}

func (q QQ) GetPath() string {
	return q.GetPathWith("pc", "")
}

// GetPathWith 生成 QQ OAuth 登录地址。
// display: "pc" 或 "mobile"（留空则不传）
// state:   可选，会被原样透传到回调地址
func (q QQ) GetPathWith(display string, state string) string {
	if strings.TrimSpace(q.Key) == "" || strings.TrimSpace(q.AppID) == "" || strings.TrimSpace(q.Redirect) == "" {
		return ""
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", strings.TrimSpace(q.AppID))
	params.Set("redirect_uri", strings.TrimSpace(q.Redirect))
	// `get_user_info` 是后端拉取 QQ 昵称/头像的基础授权范围。
	params.Set("scope", "get_user_info")

	if text := strings.TrimSpace(display); text != "" {
		params.Set("display", text)
	}
	if text := strings.TrimSpace(state); text != "" {
		params.Set("state", text)
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "graph.qq.com",
		Path:     "/oauth2.0/authorize",
		RawQuery: params.Encode(),
	}
	return u.String()
}
