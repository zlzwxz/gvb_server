package user_api

import "strings"

func readableEmailSendError(err error) string {
	if err == nil {
		return ""
	}
	raw := strings.ToLower(strings.TrimSpace(err.Error()))

	switch {
	case strings.Contains(raw, "550") && strings.Contains(raw, "recipient"):
		return "验证码发送失败：收件人邮箱不存在或拼写错误，请检查邮箱地址"
	case strings.Contains(raw, "550"):
		return "验证码发送失败：邮箱服务拒绝了投递，请检查收件邮箱是否可用"
	case strings.Contains(raw, "535") || strings.Contains(raw, "authentication failed"):
		return "验证码发送失败：发件邮箱认证失败，请检查 SMTP 授权码"
	case strings.Contains(raw, "connection refused"), strings.Contains(raw, "i/o timeout"), strings.Contains(raw, "no such host"):
		return "验证码发送失败：SMTP 连接失败，请检查邮箱服务器配置"
	default:
		return "验证码发送失败，请稍后重试"
	}
}
