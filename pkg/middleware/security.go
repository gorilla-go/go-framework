package middleware

import "github.com/gin-gonic/gin"

// SecurityMiddleware 安全中间件
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止浏览器嗅探MIME类型
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// 防止点击劫持
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		// XSS保护
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// HSTS
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// 内容安全策略
		// 允许 style-src 'unsafe-inline' 用于错误页面的内联样式
		// 允许 script-src 'self' 用于本地脚本
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'")

		// 引用策略
		c.Writer.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")

		// 功能策略
		c.Writer.Header().Set("Feature-Policy", "camera 'none'; microphone 'none'; geolocation 'none'")

		c.Next()
	}
}
