package template

import (
	"errors"
	"fmt"
	"strings"
)

// 模板错误类型定义
var (
	ErrTemplateNotFound     = errors.New("模板文件未找到")
	ErrTemplateParseError   = errors.New("模板解析错误")
	ErrTemplateRenderError  = errors.New("模板渲染错误")
	ErrManagerNotInitialized = errors.New("模板管理器未初始化")
	ErrInvalidTemplateName  = errors.New("无效的模板名称")
	ErrInvalidLayoutName    = errors.New("无效的布局名称")
	ErrBlockNotFound        = errors.New("模板块未找到")
)

// TemplateError 自定义模板错误类型
type TemplateError struct {
	Type        string
	Message     string
	TemplateName string
	FileName    string
	LineNumber  int
	Cause       error
}

// Error 实现 error 接口
func (e *TemplateError) Error() string {
	if e.TemplateName != "" {
		return fmt.Sprintf("[%s] %s (模板: %s)", e.Type, e.Message, e.TemplateName)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap 支持错误链
func (e *TemplateError) Unwrap() error {
	return e.Cause
}

// NewTemplateError 创建新的模板错误
func NewTemplateError(errorType, message, templateName string, cause error) *TemplateError {
	return &TemplateError{
		Type:        errorType,
		Message:     message,
		TemplateName: templateName,
		Cause:       cause,
	}
}

// NewParseError 创建解析错误
func NewParseError(templateName string, cause error) *TemplateError {
	return NewTemplateError("PARSE_ERROR", "模板解析失败", templateName, cause)
}

// NewRenderError 创建渲染错误
func NewRenderError(templateName string, cause error) *TemplateError {
	return NewTemplateError("RENDER_ERROR", "模板渲染失败", templateName, cause)
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(templateName string) *TemplateError {
	return NewTemplateError("NOT_FOUND", "模板文件未找到", templateName, ErrTemplateNotFound)
}

// NewBlockNotFoundError 创建块未找到错误
func NewBlockNotFoundError(templateName, blockName string) *TemplateError {
	return NewTemplateError("BLOCK_NOT_FOUND",
		fmt.Sprintf("在模板 '%s' 中未找到块 '%s'", templateName, blockName),
		templateName, ErrBlockNotFound)
}

// IsTemplateError 检查错误是否为模板错误
func IsTemplateError(err error) bool {
	_, ok := err.(*TemplateError)
	return ok
}

// IsTemplateNotFoundError 检查是否为模板未找到错误
func IsTemplateNotFoundError(err error) bool {
	if te, ok := err.(*TemplateError); ok {
		return te.Type == "NOT_FOUND"
	}
	return errors.Is(err, ErrTemplateNotFound)
}

// IsTemplateParseError 检查是否为模板解析错误
func IsTemplateParseError(err error) bool {
	if te, ok := err.(*TemplateError); ok {
		return te.Type == "PARSE_ERROR"
	}
	return errors.Is(err, ErrTemplateParseError)
}

// IsTemplateRenderError 检查是否为模板渲染错误
func IsTemplateRenderError(err error) bool {
	if te, ok := err.(*TemplateError); ok {
		return te.Type == "RENDER_ERROR"
	}
	return errors.Is(err, ErrTemplateRenderError)
}

// ValidateTemplateName 验证模板名称
func ValidateTemplateName(name string) error {
	if name == "" {
		return NewTemplateError("VALIDATION_ERROR", "模板名称不能为空", name, ErrInvalidTemplateName)
	}

	// 检查是否包含非法字符
	if strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return NewTemplateError("VALIDATION_ERROR", "模板名称包含非法字符", name, ErrInvalidTemplateName)
	}

	return nil
}

// ValidateLayoutName 验证布局名称
func ValidateLayoutName(name string) error {
	if name == "" {
		return nil // 布局名称可以为空
	}

	// 检查是否包含非法字符
	if strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return NewTemplateError("VALIDATION_ERROR", "布局名称包含非法字符", name, ErrInvalidLayoutName)
	}

	return nil
}