package errors

import (
	"fmt"
	"net/http"
)

// 定义错误码常量
const (
	// 成功
	Success   = 200
	NoContent = 204

	// 重定向
	MultipleChoices   = 300
	MovedPermanently  = 301
	Found             = 302
	SeeOther          = 303
	NotModified       = 304
	TemporaryRedirect = 307
	PermanentRedirect = 308

	// 客户端错误
	BadRequest       = 400
	Unauthorized     = 401
	Forbidden        = 403
	NotFound         = 404
	MethodNotAllowed = 405
	RequestTimeout   = 408
	Conflict         = 409
	TooManyRequests  = 429

	// 服务器错误
	InternalServerError = 500
	ServiceUnavailable  = 503
	GatewayTimeout      = 504

	// 自定义错误码
	ValidationError     = 10001
	DatabaseError       = 10002
	CacheError          = 10003
	ConfigError         = 10004
	AuthenticationError = 10005
	AuthorizationError  = 10006
)

// 错误码对应的消息
var ErrMsg = map[int]string{
	Success:             "成功",
	NoContent:           "无内容",
	MultipleChoices:     "多种选择",
	MovedPermanently:    "永久移动",
	Found:               "临时移动",
	SeeOther:            "查看其他位置",
	NotModified:         "未修改",
	TemporaryRedirect:   "临时重定向",
	PermanentRedirect:   "永久重定向",
	BadRequest:          "无效的请求",
	Unauthorized:        "未授权",
	Forbidden:           "拒绝访问",
	NotFound:            "资源不存在",
	MethodNotAllowed:    "方法不允许",
	RequestTimeout:      "请求超时",
	Conflict:            "资源冲突",
	TooManyRequests:     "请求过多",
	InternalServerError: "服务器内部错误",
	ServiceUnavailable:  "服务不可用",
	GatewayTimeout:      "网关超时",
	ValidationError:     "验证错误",
	DatabaseError:       "数据库错误",
	CacheError:          "缓存错误",
	ConfigError:         "配置错误",
	AuthenticationError: "认证错误",
	AuthorizationError:  "授权错误",
}

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误消息
	Detail  string `json:"detail"`  // 详细错误信息
	Err     error  `json:"-"`       // 原始错误
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("错误码: %d, 错误信息: %s, 详细信息: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("错误码: %d, 错误信息: %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus 根据错误码返回HTTP状态码
func (e *AppError) HTTPStatus() int {
	switch {
	case e.Code >= 400 && e.Code < 500:
		return e.Code
	case e.Code >= 500 && e.Code < 600:
		return e.Code
	case e.Code == ValidationError:
		return http.StatusBadRequest
	case e.Code == DatabaseError || e.Code == CacheError || e.Code == ConfigError:
		return http.StatusInternalServerError
	case e.Code == AuthenticationError:
		return http.StatusUnauthorized
	case e.Code == AuthorizationError:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新的错误
func New(code int, detail string, err error) *AppError {
	msg, ok := ErrMsg[code]
	if !ok {
		msg = "未知错误"
	}
	return &AppError{
		Code:    code,
		Message: msg,
		Detail:  detail,
		Err:     err,
	}
}

// NewBadRequest 创建无效请求错误
func NewBadRequest(detail string, err error) *AppError {
	return New(BadRequest, detail, err)
}

// NewUnauthorized 创建未授权错误
func NewUnauthorized(detail string, err error) *AppError {
	return New(Unauthorized, detail, err)
}

// NewForbidden 创建拒绝访问错误
func NewForbidden(detail string, err error) *AppError {
	return New(Forbidden, detail, err)
}

// NewNotFound 创建资源不存在错误
func NewNotFound(detail string, err error) *AppError {
	return New(NotFound, detail, err)
}

// NewInternalServerError 创建服务器内部错误
func NewInternalServerError(detail string, err error) *AppError {
	return New(InternalServerError, detail, err)
}

// NewValidationError 创建验证错误
func NewValidationError(detail string, err error) *AppError {
	return New(ValidationError, detail, err)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(detail string, err error) *AppError {
	return New(DatabaseError, detail, err)
}

// IsAppError 判断是否为AppError类型
func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}

	appErr, ok := err.(*AppError)
	return appErr, ok
}
