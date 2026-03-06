package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gvb-server/plugins/log_stash"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w auditResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

type auditResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// OperationAudit 自动记录写操作审计日志，避免业务接口遗漏日志埋点。
func OperationAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipAudit(c.Request.Method) {
			c.Next()
			return
		}

		writer := &auditResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		startAt := time.Now()
		c.Next()
		costMS := time.Since(startAt).Milliseconds()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		statusCode := c.Writer.Status()
		respCode, respMsg := parseAuditResponse(writer.body.Bytes())
		if respMsg == "" {
			respMsg = http.StatusText(statusCode)
		}

		level := decideAuditLevel(statusCode, respCode)
		result := "OK"
		if level != log_stash.InfoLevel {
			result = "FAIL"
		}

		content := fmt.Sprintf("%s %s %s c:%d m:%s t:%dms",
			c.Request.Method,
			path,
			result,
			respCode,
			cleanAuditMsg(respMsg),
			costMS,
		)
		content = truncateAuditContent(content, 126)

		log := log_stash.NewLogByGin(c)
		switch level {
		case log_stash.ErrorLevel:
			log.Error(content)
		case log_stash.WarnLevel:
			log.Warn(content)
		default:
			log.Info(content)
		}
	}
}

func shouldSkipAudit(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func parseAuditResponse(raw []byte) (int, string) {
	var resp auditResponse
	if len(raw) == 0 {
		return 0, ""
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, ""
	}
	return resp.Code, resp.Msg
}

func decideAuditLevel(statusCode int, respCode int) log_stash.Level {
	if statusCode >= 500 {
		return log_stash.ErrorLevel
	}
	if statusCode >= 400 {
		return log_stash.WarnLevel
	}
	if respCode != 0 {
		return log_stash.WarnLevel
	}
	return log_stash.InfoLevel
}

func cleanAuditMsg(msg string) string {
	msg = strings.TrimSpace(msg)
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.ReplaceAll(msg, "\r", " ")
	return msg
}

func truncateAuditContent(content string, max int) string {
	runes := []rune(content)
	if len(runes) <= max {
		return content
	}
	return string(runes[:max])
}
