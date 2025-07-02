package middleware

import (
	"fmt"
	"go-framework/pkg/errors"
	"go-framework/pkg/logger"
	"go-framework/pkg/response"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
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

				// 创建内部服务器错误
				appErr := errors.NewInternalServerError("服务器内部错误", fmt.Errorf("%v", r))

				// 返回统一错误响应
				resp := response.Response{
					Code:    appErr.Code,
					Message: appErr.Message,
					Data:    nil,
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, resp)
			}
		}()

		c.Next()
	}
}
