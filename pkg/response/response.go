package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/logger"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 错误码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	resp := Response{
		Code:    errors.Success,
		Message: "成功",
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SuccessWithDetail 带详细信息的成功响应
func SuccessWithDetail(c *gin.Context, detail string, data interface{}) {
	resp := Response{
		Code:    errors.Success,
		Message: detail,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// Fail 失败响应
func Fail(c *gin.Context, err error) {
	// 尝试将错误转换为AppError
	appErr, ok := errors.IsAppError(err)
	if !ok {
		// 如果不是AppError，则创建内部服务器错误
		appErr = errors.NewInternalServerError("系统错误", err)
	}

	// 记录错误日志
	logger.Errorf("请求失败: %s, 路径: %s, 错误: %v", c.Request.Method, c.Request.URL.Path, err)

	// 构建响应
	resp := Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    appErr.Detail,
	}

	// 返回响应
	c.JSON(appErr.HTTPStatus(), resp)
	c.Abort()
}

// BadRequest 无效请求响应
func BadRequest(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewBadRequest(detail, err))
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewUnauthorized(detail, err))
}

// Forbidden 拒绝访问响应
func Forbidden(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewForbidden(detail, err))
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewNotFound(detail, err))
}

// InternalServerError 服务器内部错误响应
func InternalServerError(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewInternalServerError(detail, err))
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewValidationError(detail, err))
}

// DatabaseError 数据库错误响应
func DatabaseError(c *gin.Context, detail string, err error) {
	Fail(c, errors.NewDatabaseError(detail, err))
}
