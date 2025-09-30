package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"github.com/gorilla-go/go-framework/pkg/response"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 打印堆栈信息
				stack := debug.Stack()

				errMsg := fmt.Sprintf("panic recovered: %v", r)
				logger.Errorf("%s\n%s", errMsg, string(stack))

				// 创建内部服务器错误并返回响应
				appErr := errors.NewInternalServerError("服务器内部错误", fmt.Errorf("%v", r))
				response.Fail(c, appErr)
			}
		}()

		c.Next()
	}
}
