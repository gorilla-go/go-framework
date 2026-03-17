package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapStd 将标准 net/http 中间件适配为 gin.HandlerFunc（参考 Chi 设计）
// 兼容 OpenTelemetry、OAuth2、CSRF 等遵循标准签名的 Go 中间件生态
//
// 示例：
//
//	import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
//	r.Use(middleware.WrapStd(otelhttp.NewMiddleware("my-service")))
func WrapStd(fn func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var nextCalled bool
		fn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			nextCalled = true
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)

		// 如果标准中间件没有调用 next（例如直接拦截请求），确保 gin 链也终止
		if !nextCalled {
			c.Abort()
		}
	}
}
