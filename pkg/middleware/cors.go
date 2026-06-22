package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const corsAllowHeaders = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
const corsAllowMethods = "GET, POST, PUT, DELETE, PATCH, OPTIONS"

// CORSMiddleware 跨域中间件。
//
// 不传参数时为通配模式：返回 Access-Control-Allow-Origin: *（不携带凭证，符合规范，
// 因为 "*" 不能与 Allow-Credentials 共用）。
//
// 传入允许的来源时为白名单模式：仅对命中白名单的 Origin 回显该来源并允许携带凭证（Cookie 等）。
//
//	r.Use(middleware.CORSMiddleware())                              // 允许任意来源（无凭证）
//	r.Use(middleware.CORSMiddleware("https://a.com", "https://b.com")) // 仅白名单来源（带凭证）
func CORSMiddleware(allowOrigins ...string) gin.HandlerFunc {
	allowAll := len(allowOrigins) == 0
	allowed := make(map[string]bool, len(allowOrigins))
	for _, o := range allowOrigins {
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 在所有响应（含非预检请求）上设置 CORS 头，否则真实请求会被浏览器拦截
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowed[origin] {
			// 回显具体来源 + Vary，避免缓存串源；此模式下才允许携带凭证
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Headers", corsAllowHeaders)
			c.Header("Access-Control-Allow-Methods", corsAllowMethods)
			c.Header("Access-Control-Max-Age", "86400") // 预检结果缓存 24 小时
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
