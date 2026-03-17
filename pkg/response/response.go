package response

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/validator"
)

// ErrHandlerFunc 支持直接返回 error 的 handler 类型（参考 Echo 设计）
type ErrHandlerFunc func(*gin.Context) error

// H 将 ErrHandlerFunc 包装为标准 gin.HandlerFunc
// handler 返回 *errors.AppError 时自动调用 Fail()，其他 error 转为 500
func H(f ErrHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := f(c); err != nil {
			var appErr *errors.AppError
			if stderrors.As(err, &appErr) {
				Fail(c, appErr)
			} else {
				Fail(c, errors.NewInternalServerError(err.Error(), err))
			}
		}
	}
}

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 响应消息
	Data    any    `json:"data"`    // 响应数据
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	resp := Response{
		Code:    errors.Success,
		Message: "",
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SuccessWithDetail 带详细信息的成功响应
func SuccessD(c *gin.Context, detail string, data any) {
	resp := Response{
		Code:    errors.Success,
		Message: detail,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// Fail 失败响应
func Fail(c *gin.Context, err *errors.AppError) {
	// 构建响应
	resp := Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    err.Detail,
	}

	// 返回响应
	c.JSON(err.HTTPStatus(), resp)
	c.Abort()
}

func Redirect(c *gin.Context, url string, status ...int) {
	if len(status) > 0 && status[0] == 301 {
		c.Redirect(http.StatusMovedPermanently, url)
		c.Abort()
		return
	}

	c.Redirect(http.StatusFound, url)
	c.Abort()
}

// Bind 绑定请求数据并自动校验（参考 Echo/Fiber 设计）
// 成功返回 true，失败自动写入 400 响应并返回 false
// 支持 JSON/Form/Query，具体绑定方式由 Gin 根据 Content-Type 决定
func Bind(c *gin.Context, i any) bool {
	if err := c.ShouldBind(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	if err := validator.Validate(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	return true
}

// BindJSON 绑定 JSON 请求体并自动校验
func BindJSON(c *gin.Context, i any) bool {
	if err := c.ShouldBindJSON(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	if err := validator.Validate(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	return true
}

// BindQuery 绑定 Query 参数并自动校验
func BindQuery(c *gin.Context, i any) bool {
	if err := c.ShouldBindQuery(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	if err := validator.Validate(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	return true
}

// BindUri 绑定路径参数并自动校验
func BindUri(c *gin.Context, i any) bool {
	if err := c.ShouldBindUri(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	if err := validator.Validate(i); err != nil {
		Fail(c, errors.NewValidationError(err.Error(), err))
		return false
	}
	return true
}

func BadRequest(c *gin.Context) {
	Fail(c, errors.NewBadRequest("无效请求", nil))
}

func Forbidden(c *gin.Context) {
	Fail(c, errors.NewForbidden("无权访问", nil))
}
