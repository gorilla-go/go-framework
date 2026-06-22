package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/logger"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 打印堆栈信息
				stack := debug.Stack()
				cfg := config.MustFetch()

				// 始终记录 panic 与堆栈：debug 模式虽会渲染到页面，但日志同样需要留痕
				logger.Errorf("panic recovered: %v\n%s", r, string(stack))

				errors.RenderError(
					c.Writer,
					fmt.Errorf("%v", r),
					string(stack),
					cfg.IsDebug(),
				)
				c.Abort()
				return
			}
		}()

		c.Next()
	}
}
