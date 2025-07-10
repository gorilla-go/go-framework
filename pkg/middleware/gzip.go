package middleware

import (
	"bufio"
	"compress/gzip"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
	status int
}

// Write implements http.ResponseWriter
func (g *gzipWriter) Write(data []byte) (int, error) {
	// 如果状态码是 304 Not Modified，则跳过写入
	// 因为 304 响应不应包含消息体
	if g.status == http.StatusNotModified {
		return 0, nil
	}

	if g.Header().Get("Content-Type") == "" {
		g.Header().Set("Content-Type", http.DetectContentType(data))
	}
	return g.writer.Write(data)
}

// WriteString implements gin.ResponseWriter
func (g *gzipWriter) WriteString(s string) (int, error) {
	// 如果状态码是 304 Not Modified，则跳过写入
	// 因为 304 响应不应包含消息体
	if g.status == http.StatusNotModified {
		return 0, nil
	}

	if g.Header().Get("Content-Type") == "" {
		g.Header().Set("Content-Type", http.DetectContentType([]byte(s)))
	}
	return g.writer.Write([]byte(s))
}

// Flush implements http.Flusher
func (g *gzipWriter) Flush() {
	g.writer.Flush()
	if flusher, ok := g.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker
func (g *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := g.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("ResponseWriter does not implement http.Hijacker")
}

// WriteHeader implements http.ResponseWriter
func (g *gzipWriter) WriteHeader(code int) {
	g.status = code
	g.ResponseWriter.WriteHeader(code)
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return &gzipWriter{
			writer: gzip.NewWriter(nil),
		}
	},
}

// GzipMiddleware 返回一个 gzip 压缩中间件
// 默认压缩级别为 DefaultCompression
func GzipMiddleware() gin.HandlerFunc {
	return GzipWithLevelMiddleware(gzip.DefaultCompression)
}

// GzipWithLevelMiddleware 返回一个指定压缩级别的 gzip 压缩中间件
// 压缩级别范围：gzip.NoCompression (-1) 到 gzip.BestCompression (9)
// gzip.DefaultCompression (0) 表示默认压缩级别
// gzip.BestSpeed (1) 表示最快速度
func GzipWithLevelMiddleware(level int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续中间件，以便可以捕获状态码
		c.Next()

		// 如果响应状态码是 304 Not Modified，则不进行压缩处理
		// 304 响应不应该包含响应体
		if c.Writer.Status() == http.StatusNotModified {
			return
		}

		// 已经写入了响应，不能再压缩
		if c.Writer.Written() {
			return
		}

		// 对于 HEAD 请求不进行压缩
		if c.Request.Method == "HEAD" {
			return
		}

		// 检查客户端是否支持 gzip 压缩
		if !clientAcceptsGzip(c.Request) {
			return
		}

		// 检查是否是二进制内容类型，如果是则不进行 gzip 压缩
		if !shouldCompress(c.Writer.Header().Get("Content-Type")) {
			return
		}

		gz := gzipWriterPool.Get().(*gzipWriter)
		defer gzipWriterPool.Put(gz)

		// 根据指定的压缩级别创建新的 writer
		writer, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			// 如果创建 writer 失败，记录错误并返回
			c.Error(err)
			return
		}
		gz.writer = writer
		gz.ResponseWriter = c.Writer
		gz.status = c.Writer.Status() // 保存当前状态码

		// 保留原始的 Content-Type（如果已设置）
		contentType := c.Writer.Header().Get("Content-Type")

		// 删除Content-Length头部，因为压缩后内容长度会发生变化
		c.Writer.Header().Del("Content-Length")

		// 设置响应头，标记内容使用 gzip 压缩
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// 如果原始响应已设置 Content-Type，则保留它
		if contentType != "" {
			c.Header("Content-Type", contentType)
		}

		// 替换响应写入器
		originalWriter := c.Writer
		c.Writer = gz

		// 确保 writer 关闭并恢复原始响应写入器
		defer func() {
			if err := gz.writer.Close(); err != nil {
				c.Error(err)
			}
			c.Writer = originalWriter
		}()
	}
}

// clientAcceptsGzip 检查客户端是否支持 gzip 压缩
func clientAcceptsGzip(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "gzip")
}

// shouldCompress 检查内容类型是否应该被压缩
func shouldCompress(contentType string) bool {
	// 如果内容类型为空，则假定可以压缩
	if contentType == "" {
		return true
	}

	// 获取主要内容类型
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))

	// 已经压缩的内容类型列表
	compressedTypes := []string{
		"image/", "video/", "audio/", "application/zip",
		"application/x-gzip", "application/x-rar-compressed",
		"application/octet-stream",
	}

	for _, t := range compressedTypes {
		if strings.HasPrefix(contentType, t) {
			return false
		}
	}

	return true
}
