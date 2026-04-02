package qq

import (
	"encoding/json"
	"errors"
	"fmt"
	"gvb-server/global"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type QQInfo struct {
	Ret      int    `json:"ret"` // ret=0 才是成功
	Msg      string `json:"msg"` // ret!=0 时可能携带错误信息
	Nickname string `json:"nickname"`     // 昵称
	Gender   string `json:"gender"`       // 性别
	Avatar   string `json:"-"`             // 头像（已归一化后的最佳选择）
	Avatar1  string `json:"figureurl_qq_1"` // 头像（40x40）
	Avatar2  string `json:"figureurl_qq_2"` // 头像（100x100）
	AvatarQQ string `json:"figureurl_qq"`   // 兼容字段（部分文档示例里存在）
	OpenID   string `json:"open_id"`
}

type QQLogin struct {
	appID     string
	appKey    string
	redirect  string
	code      string
	accessTok string
	openID    string
}

var qqHTTPClient = &http.Client{Timeout: 12 * time.Second}

func NewQQLogin(code string) (qqInfo QQInfo, err error) {
	qqLogin := &QQLogin{
		appID:    global.Config.QQ.AppID,
		appKey:   global.Config.QQ.Key,
		redirect: global.Config.QQ.Redirect,
		code:     code,
	}
	err = qqLogin.GetAccessToken()
	if err != nil {
		return qqInfo, err
	}
	err = qqLogin.GetOpenID()
	if err != nil {
		return qqInfo, err
	}
	qqInfo, err = qqLogin.GetUserInfo()
	if err != nil {
		return qqInfo, err
	}
	qqInfo.OpenID = qqLogin.openID
	qqInfo.Nickname = strings.TrimSpace(qqInfo.Nickname)
	avatar := strings.TrimSpace(qqInfo.Avatar2)
	if avatar == "" {
		avatar = strings.TrimSpace(qqInfo.Avatar1)
	}
	if avatar == "" {
		avatar = strings.TrimSpace(qqInfo.AvatarQQ)
	}
	qqInfo.Avatar = avatar
	return qqInfo, nil
}

// GetAccessToken 获取token
func (q *QQLogin) GetAccessToken() error {
	// 获取Access_token
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", q.appID)
	params.Add("client_secret", q.appKey)
	params.Add("code", q.code)
	params.Add("redirect_uri", q.redirect)
	u := url.URL{
		Scheme:   "https",
		Host:     "graph.qq.com",
		Path:     "/oauth2.0/token",
		RawQuery: params.Encode(),
	}

	res, err := qqHTTPClient.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return errors.New("qq access token 响应为空")
	}

	// 失败时 QQ 可能返回 `callback( {"error":...,"error_description":"..."} );`
	if strings.Contains(text, "{") && strings.Contains(text, "error") {
		var payload struct {
			Error            int    `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := parseCallbackJSON(text, &payload); err == nil && payload.Error != 0 {
			msg := strings.TrimSpace(payload.ErrorDescription)
			if msg == "" {
				msg = fmt.Sprintf("error=%d", payload.Error)
			}
			return fmt.Errorf("qq 获取 access_token 失败：%s", msg)
		}
	}

	qs, err := url.ParseQuery(text)
	if err != nil {
		return fmt.Errorf("解析 qq access_token 响应失败：%w", err)
	}
	token := strings.TrimSpace(qs.Get("access_token"))
	if token == "" {
		if msg := strings.TrimSpace(qs.Get("error_description")); msg != "" {
			return fmt.Errorf("qq 获取 access_token 失败：%s", msg)
		}
		return fmt.Errorf("qq access_token 不存在：%s", text)
	}
	q.accessTok = token
	return nil
}

// GetOpenID 获取openid
func (q *QQLogin) GetOpenID() error {
	// 获取openid
	u := url.URL{
		Scheme: "https",
		Host:   "graph.qq.com",
		Path:   "/oauth2.0/me",
	}
	values := url.Values{}
	values.Set("access_token", q.accessTok)
	u.RawQuery = values.Encode()

	res, err := qqHTTPClient.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return errors.New("qq openid 响应为空")
	}

	var payload struct {
		OpenID           string `json:"openid"`
		Error            int    `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := parseCallbackJSON(text, &payload); err != nil {
		return fmt.Errorf("解析 qq openid 失败：%w", err)
	}
	if payload.Error != 0 {
		msg := strings.TrimSpace(payload.ErrorDescription)
		if msg == "" {
			msg = fmt.Sprintf("error=%d", payload.Error)
		}
		return fmt.Errorf("qq 获取 openid 失败：%s", msg)
	}
	openID := strings.TrimSpace(payload.OpenID)
	if openID == "" {
		return errors.New("qq openid 不存在")
	}

	q.openID = openID
	return nil
}

// GetUserInfo 获取用户信息
func (q *QQLogin) GetUserInfo() (qqInfo QQInfo, err error) {
	params := url.Values{}
	params.Add("access_token", q.accessTok)
	params.Add("oauth_consumer_key", q.appID)
	params.Add("openid", q.openID)
	u := url.URL{
		Scheme:   "https",
		Host:     "graph.qq.com",
		Path:     "/user/get_user_info",
		RawQuery: params.Encode(),
	}

	res, err := qqHTTPClient.Get(u.String())
	if err != nil {
		return qqInfo, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return qqInfo, err
	}
	err = json.Unmarshal(data, &qqInfo)
	if err != nil {
		return qqInfo, err
	}
	if qqInfo.Ret != 0 {
		msg := strings.TrimSpace(qqInfo.Msg)
		if msg == "" {
			msg = fmt.Sprintf("ret=%d", qqInfo.Ret)
		}
		return qqInfo, fmt.Errorf("qq 获取用户信息失败：%s", msg)
	}
	return qqInfo, nil
}

// parseCallbackJSON 解析 QQ 常见的 `callback( {...} );` 或纯 JSON 响应。
func parseCallbackJSON(text string, dst any) error {
	trimmed := strings.TrimSpace(text)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end < 0 || end < start {
		return fmt.Errorf("invalid qq callback response: %s", trimmed)
	}
	if err := json.Unmarshal([]byte(trimmed[start:end+1]), dst); err != nil {
		return err
	}
	return nil
}

// readAll 读取所有数据并将其转换为字符串（保留此函数仅用于兼容历史调试输出）。
func readAll(r io.Reader) string {
	b, err := io.ReadAll(r)
	if err != nil {
		log.Println("[qq] readAll failed:", err)
		return ""
	}
	return string(b)
}
