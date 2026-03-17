package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"go.uber.org/zap"
)

const (
	maxBodyLogSize = 1024
	// LogEntryKey 是存储在 gin.Context 中的 LogEntry 键名
	LogEntryKey = "log_entry"
)

// LogEntry 请求日志条目，支持在 handler 中追加自定义字段（参考 Chi LogEntry 设计）
// 用法：middleware.GetLogEntry(c).AddField("user_id", userID)
type LogEntry struct {
	fields []zap.Field
}

// AddField 追加一个结构化日志字段，在请求结束时一并写入日志
func (e *LogEntry) AddField(key string, value any) {
	e.fields = append(e.fields, zap.Any(key, value))
}

// GetLogEntry 从 gin.Context 获取当前请求的 LogEntry
// 可在任意 handler 或下游中间件中调用以追加字段
func GetLogEntry(c *gin.Context) *LogEntry {
	if v, exists := c.Get(LogEntryKey); exists {
		if entry, ok := v.(*LogEntry); ok {
			return entry
		}
	}
	return &LogEntry{} // 返回空对象防止 nil panic
}

// Logger 日志中间件（基于 Zap 结构化日志）
// isDev=true 时，对 4xx/5xx 请求额外记录请求体和响应体（便于调试）
func Logger(isDev bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 创建 LogEntry 并注入 context，供下游 handler 追加字段
		entry := &LogEntry{}
		c.Set(LogEntryKey, entry)

		// dev 模式下读取请求体（读后需还原）
		var reqBody string
		if isDev && c.Request.Body != nil {
			raw, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
			if len(raw) > maxBodyLogSize {
				reqBody = string(raw[:maxBodyLogSize]) + "..."
			} else if len(raw) > 0 {
				reqBody = string(raw)
			}
		}

		// dev 模式下捕获响应体
		var rw *responseWriter
		if isDev {
			rw = &responseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
			c.Writer = rw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		}
		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if ua := c.Request.UserAgent(); ua != "" {
			fields = append(fields, zap.String("user_agent", ua))
		}

		// 仅在 dev 模式且请求出错时附加 body 信息
		if isDev && status >= 400 {
			if reqBody != "" {
				fields = append(fields, zap.String("req_body", reqBody))
			}
			if rw != nil && rw.body.Len() > 0 {
				resp := rw.body.String()
				if len(resp) > maxBodyLogSize {
					resp = resp[:maxBodyLogSize] + "..."
				}
				fields = append(fields, zap.String("resp_body", resp))
			}
		}

		// 合并 handler 通过 LogEntry 追加的自定义字段
		fields = append(fields, entry.fields...)

		msg := c.Request.Method + " " + path
		log := logger.ZapLogger

		switch {
		case status >= 500:
			log.Error(msg, fields...)
		case status >= 400:
			log.Warn(msg, fields...)
		default:
			log.Info(msg, fields...)
		}
	}
}

// responseWriter 捕获响应体（仅 dev 模式使用）
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
