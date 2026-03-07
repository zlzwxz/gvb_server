package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gvb-server/plugins/log_stash"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *auditResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *auditResponseWriter) ReadFrom(reader io.Reader) (int64, error) {
	teeReader := io.TeeReader(reader, w.body)
	if readFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return readFrom.ReadFrom(teeReader)
	}
	return io.Copy(w.ResponseWriter, teeReader)
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
		requestBody := captureRequestSnapshot(c)

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
		content = truncateAuditContent(content, 255)
		requestPayload := truncateAuditContent(strings.TrimSpace(string(requestBody)), 6000)
		responsePayload := truncateAuditContent(describeAuditResponse(writer.body.Bytes(), statusCode, c.Writer.Header()), 6000)

		log := log_stash.NewLogByGin(c)
		log.Log(level, log_stash.Payload{
			Content:      content,
			Method:       c.Request.Method,
			Path:         path,
			StatusCode:   statusCode,
			RespCode:     respCode,
			RequestBody:  requestPayload,
			ResponseBody: responsePayload,
		})
	}
}

func shouldSkipAudit(method string) bool {
	return method == http.MethodHead || method == http.MethodOptions
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

func captureRequestSnapshot(c *gin.Context) []byte {
	snapshot := map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	}
	if len(c.Request.URL.RawQuery) > 0 {
		snapshot["raw_query"] = c.Request.URL.RawQuery
	}
	if query := c.Request.URL.Query(); len(query) > 0 {
		snapshot["query"] = query
	}
	if len(c.Params) > 0 {
		params := map[string]string{}
		for _, item := range c.Params {
			params[item.Key] = item.Value
		}
		if len(params) > 0 {
			snapshot["params"] = params
		}
	}

	if c.Request == nil || c.Request.Body == nil {
		return mustMarshalAuditPayload(snapshot)
	}
	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	if strings.Contains(contentType, "multipart/form-data") {
		if form, err := c.MultipartForm(); err == nil && form != nil {
			if len(form.Value) > 0 {
				snapshot["form"] = form.Value
			}
			if files := describeMultipartFiles(form.File); len(files) > 0 {
				snapshot["files"] = files
			}
		} else {
			snapshot["body"] = "[multipart/form-data]"
		}
		return mustMarshalAuditPayload(snapshot)
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return mustMarshalAuditPayload(snapshot)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(bytes.TrimSpace(body)) == 0 {
		return mustMarshalAuditPayload(snapshot)
	}
	if json.Valid(body) {
		snapshot["body"] = json.RawMessage(body)
		return mustMarshalAuditPayload(snapshot)
	}
	snapshot["body"] = string(body)
	return mustMarshalAuditPayload(snapshot)
}

func mustMarshalAuditPayload(payload map[string]interface{}) []byte {
	byteData, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return byteData
}

func describeAuditResponse(raw []byte, statusCode int, headers http.Header) string {
	contentType := strings.TrimSpace(headers.Get("Content-Type"))
	contentDisposition := strings.TrimSpace(headers.Get("Content-Disposition"))
	trimmed := bytes.TrimSpace(raw)
	if shouldRecordBinaryResponse(contentType, contentDisposition) {
		byteData, _ := json.Marshal(map[string]interface{}{
			"status_code":         statusCode,
			"content_type":        contentType,
			"content_disposition": contentDisposition,
			"size":                len(trimmed),
			"body":                "[binary response omitted]",
		})
		return string(byteData)
	}
	if len(trimmed) == 0 {
		byteData, _ := json.Marshal(map[string]interface{}{
			"status_code":  statusCode,
			"content_type": contentType,
			"body":         "",
		})
		return string(byteData)
	}
	if json.Valid(trimmed) {
		return string(trimmed)
	}
	byteData, _ := json.Marshal(map[string]interface{}{
		"status_code":  statusCode,
		"content_type": contentType,
		"size":         len(trimmed),
		"body":         string(trimmed),
	})
	return string(byteData)
}

func describeMultipartFiles(files map[string][]*multipart.FileHeader) map[string][]map[string]interface{} {
	result := make(map[string][]map[string]interface{}, len(files))
	for field, fileHeaders := range files {
		items := make([]map[string]interface{}, 0, len(fileHeaders))
		for _, fileHeader := range fileHeaders {
			if fileHeader == nil {
				continue
			}
			items = append(items, map[string]interface{}{
				"filename": fileHeader.Filename,
				"size":     fileHeader.Size,
				"header":   fileHeader.Header,
			})
		}
		if len(items) > 0 {
			result[field] = items
		}
	}
	return result
}

func shouldRecordBinaryResponse(contentType string, contentDisposition string) bool {
	if strings.Contains(strings.ToLower(contentDisposition), "attachment") {
		return true
	}
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if contentType == "" {
		return false
	}
	if strings.Contains(contentType, "json") || strings.HasPrefix(contentType, "text/") {
		return false
	}
	if strings.Contains(contentType, "xml") || strings.Contains(contentType, "javascript") {
		return false
	}
	return true
}
