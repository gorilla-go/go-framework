package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultMinLength 默认最小压缩长度（字节）
	// 小于此长度的响应不进行压缩，因为压缩开销可能大于收益
	DefaultMinLength = 1024
)

// gzipWriter 实现了 gin.ResponseWriter 接口，用于 gzip 压缩
type gzipWriter struct {
	gin.ResponseWriter
	writer       *gzip.Writer
	minLength    int
	written      bool
	size         int
	shouldCompr  bool // 是否应该压缩
	compressing  bool // 是否正在压缩
}

// Write 实现 http.ResponseWriter
func (g *gzipWriter) Write(data []byte) (int, error) {
	// 首先累计大小
	g.size += len(data)

	// 如果尚未调用 WriteHeader，自动调用（模拟 http.ResponseWriter 行为）
	if !g.written {
		// 自动检测 Content-Type
		if g.Header().Get("Content-Type") == "" {
			g.Header().Set("Content-Type", http.DetectContentType(data))
		}
		g.WriteHeader(http.StatusOK)
	}

	// 如果正在压缩，使用 gzip writer
	if g.compressing {
		return g.writer.Write(data)
	}

	// 否则直接写入原始响应
	return g.ResponseWriter.Write(data)
}

// WriteString 实现 gin.ResponseWriter
func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

// WriteHeader 实现 http.ResponseWriter
func (g *gzipWriter) WriteHeader(code int) {
	// 在 WriteHeader 时检查 Content-Type 并决定是否压缩
	if !g.written {
		g.written = true

		// 检查是否应该压缩此内容类型
		contentType := g.Header().Get("Content-Type")
		g.shouldCompr = shouldCompress(contentType)

		// 如果应该压缩，设置响应头
		if g.shouldCompr && code != http.StatusNoContent && code != http.StatusNotModified {
			g.Header().Del("Content-Length")
			g.Header().Set("Content-Encoding", "gzip")
			g.Header().Add("Vary", "Accept-Encoding")
			g.compressing = true
		}
	}

	// 对于 204、304 等不包含响应体的状态码，删除 Content-Encoding
	if code == http.StatusNoContent || code == http.StatusNotModified {
		g.Header().Del("Content-Encoding")
		g.compressing = false
	}

	g.ResponseWriter.WriteHeader(code)
}

// Flush 实现 http.Flusher
func (g *gzipWriter) Flush() {
	if g.writer != nil {
		_ = g.writer.Flush()
	}
	if flusher, ok := g.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现 http.Hijacker
func (g *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := g.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// CloseNotify 实现 http.CloseNotifier（已废弃但保留兼容性）
func (g *gzipWriter) CloseNotify() <-chan bool {
	if notifier, ok := g.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	// 返回一个永远不会关闭的通道
	return make(chan bool)
}

// Size 返回已写入的字节数
func (g *gzipWriter) Size() int {
	return g.size
}

// Written 返回是否已写入数据
func (g *gzipWriter) Written() bool {
	return g.written
}

// gzipWriterPool gzip writer 对象池
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return &gzipWriter{
			writer:    gz,
			minLength: DefaultMinLength,
		}
	},
}

// getGzipWriter 从池中获取 gzipWriter
func getGzipWriter(w gin.ResponseWriter, level int, minLength int) *gzipWriter {
	gz := gzipWriterPool.Get().(*gzipWriter)

	// 重置 writer 的压缩级别（如果需要）
	if level != gzip.DefaultCompression {
		gz.writer.Reset(io.Discard)
		newWriter, _ := gzip.NewWriterLevel(w, level)
		gz.writer = newWriter
	} else {
		gz.writer.Reset(w)
	}

	gz.ResponseWriter = w
	gz.minLength = minLength
	gz.written = false
	gz.size = 0
	gz.shouldCompr = false
	gz.compressing = false

	return gz
}

// putGzipWriter 将 gzipWriter 放回池中
func putGzipWriter(gz *gzipWriter) {
	// 只有在实际使用了压缩时才关闭 writer
	if gz.writer != nil && gz.compressing {
		_ = gz.writer.Close()
	}
	gzipWriterPool.Put(gz)
}

// GzipConfig gzip 配置
type GzipConfig struct {
	// Level 压缩级别 (-2 到 9)
	// -2 = HuffmanOnly, -1 = DefaultCompression, 0 = NoCompression
	// 1 = BestSpeed, 9 = BestCompression
	Level int

	// MinLength 最小压缩长度（字节）
	MinLength int

	// ExcludedExtensions 排除的文件扩展名
	ExcludedExtensions []string

	// ExcludedPaths 排除的路径
	ExcludedPaths []string

	// ExcludedPathPrefixes 排除的路径前缀
	ExcludedPathPrefixes []string
}

// DefaultGzipConfig 默认配置
var DefaultGzipConfig = GzipConfig{
	Level:     gzip.DefaultCompression,
	MinLength: DefaultMinLength,
	ExcludedExtensions: []string{
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico",
		".mp4", ".mp3", ".avi", ".mov",
		".zip", ".tar", ".gz", ".rar", ".7z",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
	},
	ExcludedPaths:        []string{},
	ExcludedPathPrefixes: []string{},
}

// GzipMiddleware 返回一个 gzip 压缩中间件（使用默认配置）
func GzipMiddleware() gin.HandlerFunc {
	return GzipWithConfig(DefaultGzipConfig)
}

// GzipWithLevelMiddleware 返回一个指定压缩级别的 gzip 压缩中间件
func GzipWithLevelMiddleware(level int) gin.HandlerFunc {
	config := DefaultGzipConfig
	config.Level = level
	return GzipWithConfig(config)
}

// GzipWithConfig 返回一个使用自定义配置的 gzip 压缩中间件
func GzipWithConfig(config GzipConfig) gin.HandlerFunc {
	// 验证压缩级别
	if config.Level < gzip.HuffmanOnly || config.Level > gzip.BestCompression {
		config.Level = gzip.DefaultCompression
	}

	// 验证最小长度
	if config.MinLength < 0 {
		config.MinLength = DefaultMinLength
	}

	return func(c *gin.Context) {
		// 检查是否应该跳过此请求
		if shouldSkipRequest(c, config) {
			c.Next()
			return
		}

		// 检查客户端是否支持 gzip
		if !clientAcceptsGzip(c.Request) {
			c.Next()
			return
		}

		// 获取 gzipWriter
		gz := getGzipWriter(c.Writer, config.Level, config.MinLength)
		c.Writer = gz

		// 处理请求
		c.Next()

		// 关闭并回收 writer
		putGzipWriter(gz)
	}
}

// shouldSkipRequest 检查是否应该跳过此请求
func shouldSkipRequest(c *gin.Context, config GzipConfig) bool {
	// HEAD 和 OPTIONS 请求不压缩
	if c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
		return true
	}

	// WebSocket 连接不压缩
	if isWebSocketRequest(c.Request) {
		return true
	}

	// 检查路径是否被排除
	path := c.Request.URL.Path
	for _, excludedPath := range config.ExcludedPaths {
		if path == excludedPath {
			return true
		}
	}

	// 检查路径前缀是否被排除
	for _, prefix := range config.ExcludedPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// 检查文件扩展名是否被排除
	for _, ext := range config.ExcludedExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// clientAcceptsGzip 检查客户端是否支持 gzip 压缩
func clientAcceptsGzip(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "gzip")
}

// isWebSocketRequest 检查是否为 WebSocket 请求
func isWebSocketRequest(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket"
}

// shouldCompress 检查内容类型是否应该被压缩
func shouldCompress(contentType string) bool {
	// 如果内容类型为空，则假定可以压缩
	if contentType == "" {
		return true
	}

	// 提取主要内容类型（忽略参数）
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))

	// 可压缩的内容类型
	compressibleTypes := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/x-javascript",
		"application/xhtml+xml",
		"application/rss+xml",
		"application/atom+xml",
		"application/x-font-ttf",
		"application/x-font-opentype",
		"application/vnd.ms-fontobject",
		"image/svg+xml",
		"font/",
	}

	for _, t := range compressibleTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}

	// 已经压缩的内容类型
	compressedTypes := []string{
		"image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp",
		"video/", "audio/",
		"application/zip", "application/gzip", "application/x-gzip",
		"application/x-rar", "application/x-rar-compressed",
		"application/x-7z-compressed",
		"application/octet-stream",
		"application/pdf",
	}

	for _, t := range compressedTypes {
		if strings.HasPrefix(contentType, t) {
			return false
		}
	}

	// 默认不压缩未知类型
	return false
}