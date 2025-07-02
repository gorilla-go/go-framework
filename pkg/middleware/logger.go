package middleware

import (
	"bytes"
	"go-framework/pkg/logger"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 获取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
		}

		// 重置请求体，因为读取后会清空
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

		// 创建自定义响应写入器
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 请求信息
		requestInfo := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"status":     c.Writer.Status(),
			"latency":    latency.String(),
		}

		// 请求头
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		requestInfo["headers"] = headers

		// 请求体（可以根据需要添加）
		if len(requestBody) > 0 {
			// 限制请求体的大小，避免记录过大的请求体
			if len(requestBody) > 1024 {
				requestInfo["body"] = string(requestBody[:1024]) + "..."
			} else {
				requestInfo["body"] = string(requestBody)
			}
		}

		// 响应体（可以根据需要添加）
		if writer.body.Len() > 0 {
			// 限制响应体的大小，避免记录过大的响应体
			if writer.body.Len() > 1024 {
				responseBody := writer.body.Bytes()[:1024]
				requestInfo["response"] = string(responseBody) + "..."
			} else {
				requestInfo["response"] = writer.body.String()
			}
		}

		// 根据状态码记录不同级别的日志
		if c.Writer.Status() >= 500 {
			logger.Errorf("请求异常: %v", requestInfo)
		} else if c.Writer.Status() >= 400 {
			logger.Warnf("请求警告: %v", requestInfo)
		} else {
			logger.Infof("请求正常: %v", requestInfo)
		}
	}
}

// responseWriter 自定义响应写入器，用于捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，同时写入到原响应和缓冲区
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 重写WriteString方法，同时写入到原响应和缓冲区
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
