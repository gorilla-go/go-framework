package request

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/validator"
)

// Bind 绑定请求数据并自动校验
// 支持 JSON/Form/Query，具体绑定方式由 Gin 根据 Content-Type 决定
func Bind(c *gin.Context, i any) error {
	if err := c.ShouldBind(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	if err := validator.Validate(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	return nil
}

// BindJSON 绑定 JSON 请求体并自动校验
func BindJSON(c *gin.Context, i any) error {
	if err := c.ShouldBindJSON(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	if err := validator.Validate(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	return nil
}

// BindQuery 绑定 Query 参数并自动校验
func BindQuery(c *gin.Context, i any) error {
	if err := c.ShouldBindQuery(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	if err := validator.Validate(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	return nil
}

// BindUri 绑定路径参数并自动校验
func BindUri(c *gin.Context, i any) error {
	if err := c.ShouldBindUri(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	if err := validator.Validate(i); err != nil {
		return errors.NewValidationError(err.Error(), err)
	}
	return nil
}
