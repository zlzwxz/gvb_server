package log_stash

import (
	"gvb-server/global"
	"gvb-server/utils"
	"gvb-server/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Log struct {
	ip     string `json:"ip"`
	addr   string `json:"addr"`
	userId uint   `json:"user_id"`
	method string `json:"method"`
	path   string `json:"path"`
}

type Payload struct {
	Content      string
	Method       string
	Path         string
	StatusCode   int
	RespCode     int
	RequestBody  string
	ResponseBody string
}

func New(ip string, token string) *Log {
	return NewWithRequest(ip, token, "", "")
}

func NewWithRequest(ip string, token string, method string, path string) *Log {
	var userID uint
	// 检查token是否有效再解析
	if token != "" {
		claims, err := jwts.ParseToken(token)
		if err == nil && claims != nil {
			userID = claims.UserID
		}
	}
	addr := utils.GetAddr(ip)
	// 拿到用户id
	return &Log{
		ip:     ip,
		addr:   addr,
		userId: userID,
		method: strings.ToUpper(strings.TrimSpace(method)),
		path:   strings.TrimSpace(path),
	}
}

func NewLogByGin(c *gin.Context) *Log {
	ip := c.ClientIP()
	token := strings.TrimSpace(c.Request.Header.Get("token"))
	if token == "" {
		authHeader := strings.TrimSpace(c.Request.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	return NewWithRequest(ip, token, c.Request.Method, path)
}

func (l Log) Debug(content string) {
	l.send(DebugLevel, Payload{Content: content})
}
func (l Log) Info(content string) {
	l.send(InfoLevel, Payload{Content: content})
}
func (l Log) Warn(content string) {
	l.send(WarnLevel, Payload{Content: content})
}
func (l Log) Error(content string) {
	l.send(ErrorLevel, Payload{Content: content})
}

func (l Log) Log(level Level, payload Payload) {
	l.send(level, payload)
}

func (l Log) send(level Level, payload Payload) {
	method := strings.ToUpper(strings.TrimSpace(payload.Method))
	if method == "" {
		method = l.method
	}
	path := strings.TrimSpace(payload.Path)
	if path == "" {
		path = l.path
	}
	err := global.DB.Create(&LogStashModel{
		IP:           l.ip,
		Addr:         l.addr,
		Level:        level,
		Content:      payload.Content,
		UserID:       l.userId,
		Method:       method,
		Path:         path,
		StatusCode:   payload.StatusCode,
		RespCode:     payload.RespCode,
		RequestBody:  payload.RequestBody,
		ResponseBody: payload.ResponseBody,
	}).Error
	if err != nil {
		logrus.Error(err)
	}
}
