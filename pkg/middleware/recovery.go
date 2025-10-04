package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/logger"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 打印堆栈信息
				stack := debug.Stack()
				cfg := config.MustFetch()

				if !cfg.IsDebug() {
					logger.Errorf(
						"%s\n%s",
						fmt.Sprintf("panic recovered: %v", r),
						string(stack),
					)
				}

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
