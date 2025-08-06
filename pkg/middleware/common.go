package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
)

// HandleError 处理错误并发送统一的错误响应
// 这个函数可以替代中间件中的重复错误处理代码
func HandleError(c *gin.Context, code int, message string, err error) {
	// 创建应用错误
	appErr := errors.New(code, message, err)

	// 构建统一响应
	resp := response.Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	}

	// 发送响应并终止请求处理
	c.AbortWithStatusJSON(code, resp)
}

// HandleAppError 处理应用错误并发送统一的错误响应
func HandleAppError(c *gin.Context, appErr *errors.AppError) {
	resp := response.Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	}

	c.AbortWithStatusJSON(appErr.HTTPStatus(), resp)
}

// 以下是常用错误处理的简便函数

// HandleBadRequest 处理400错误
func HandleBadRequest(c *gin.Context, message string, err error) {
	HandleError(c, errors.BadRequest, message, err)
}

// HandleUnauthorized 处理401错误
func HandleUnauthorized(c *gin.Context, message string, err error) {
	HandleError(c, errors.Unauthorized, message, err)
}

// HandleForbidden 处理403错误
func HandleForbidden(c *gin.Context, message string, err error) {
	HandleError(c, errors.Forbidden, message, err)
}

// HandleNotFound 处理404错误
func HandleNotFound(c *gin.Context, message string, err error) {
	HandleError(c, errors.NotFound, message, err)
}

// HandleTooManyRequests 处理429错误
func HandleTooManyRequests(c *gin.Context, message string, err error) {
	HandleError(c, errors.TooManyRequests, message, err)
}

// HandleInternalServerError 处理500错误
func HandleInternalServerError(c *gin.Context, message string, err error) {
	HandleError(c, errors.InternalServerError, message, err)
}
