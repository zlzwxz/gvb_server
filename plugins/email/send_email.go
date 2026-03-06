package email

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"strings"

	"gvb-server/global"

	"gopkg.in/gomail.v2"
)

type Subject string

const (
	Code  Subject = "平台验证码"
	Note  Subject = "操作通知"
	Alarm Subject = "告警通知"
)

type Api struct {
	Subject Subject
}

func (a Api) Send(name, body string) error {
	return send(name, string(a.Subject), body)
}

func NewCode() Api {
	return Api{
		Subject: Code,
	}
}
func NewNote() Api {
	return Api{
		Subject: Note,
	}
}
func NewAlarm() Api {
	return Api{
		Subject: Alarm,
	}
}

// send 邮件发送  发给谁，主题，正文
func send(name, subject, body string) error {
	e := global.Config.Email
	to := strings.TrimSpace(name)
	if to == "" {
		return fmt.Errorf("收件邮箱不能为空")
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("收件邮箱格式不正确")
	}
	return sendMail(
		e.User,
		e.Password,
		e.Host,
		e.Port,
		to,
		e.DefaultFromEmail,
		subject,
		body,
		e.UseSSL,
		e.UserTls,
	)
}

func sendMail(userName, authCode, host string, port int, mailTo, sendName string, subject, body string, useSSL bool, useTLS bool) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(strings.TrimSpace(userName), strings.TrimSpace(sendName)))
	m.SetHeader("To", strings.TrimSpace(mailTo))
	m.SetHeader("Subject", strings.TrimSpace(subject))
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, userName, authCode)
	if useSSL || port == 465 {
		d.SSL = true
	}
	if useTLS || d.SSL {
		d.TLSConfig = &tls.Config{
			ServerName: strings.TrimSpace(host),
		}
	}

	return d.DialAndSend(m)
}
