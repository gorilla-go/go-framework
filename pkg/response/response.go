package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
)

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

func BadRequest(c *gin.Context) {
	Fail(c, errors.NewBadRequest("无效请求", nil))
}

func Forbidden(c *gin.Context) {
	Fail(c, errors.NewForbidden("无权访问", nil))
}
