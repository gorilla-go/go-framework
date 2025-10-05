package errors

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla-go/go-framework/pkg/config"
)

// 模板错误类型定义
var (
	ErrTemplateNotFound      = errors.New("模板文件未找到")
	ErrTemplateParseError    = errors.New("模板解析错误")
	ErrTemplateRenderError   = errors.New("模板渲染错误")
	ErrManagerNotInitialized = errors.New("模板管理器未初始化")
	ErrInvalidTemplateName   = errors.New("无效的模板名称")
	ErrInvalidLayoutName     = errors.New("无效的布局名称")
	ErrBlockNotFound         = errors.New("模板块未找到")
)

// 正则表达式缓存（延迟初始化）
var (
	templateErrorRe     *regexp.Regexp
	templateErrorReOnce sync.Once
)

// getTemplateErrorRegex 获取模板错误正则表达式（延迟初始化 + 缓存）
func getTemplateErrorRegex() *regexp.Regexp {
	templateErrorReOnce.Do(func() {
		cfg := config.MustFetch()
		// 匹配所有可能的模板路径格式，包括子目录（如 layouts/main.html）
		// [\w\-/.]+ 可以匹配字母、数字、下划线、连字符、斜杠和点
		templateErrorRe = regexp.MustCompile(`template:\s+([\w\-/]+\.` + cfg.Template.Extension + `):(\d+)`)
	})
	return templateErrorRe
}

// TemplateError 自定义模板错误类型
type TemplateError struct {
	Type         string
	Message      string
	TemplateName string
	FileName     string
	LineNumber   int
	Cause        error
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
		Type:         errorType,
		Message:      message,
		TemplateName: templateName,
		Cause:        cause,
	}
}

// NewParseError 创建解析错误
func NewParseError(templateName string, cause error) *TemplateError {
	return NewTemplateError("PARSE_ERROR", "模板解析失败", templateName, cause)
}

// NewRenderError 创建渲染错误
func NewRenderError(templateName string, cause error) *TemplateError {
	renderErr := NewTemplateError("RENDER_ERROR", "模板渲染失败", templateName, cause)

	// 尝试从错误信息中提取文件名和行号
	// Go template 渲染错误格式: "template: filename.html:10: error message"
	if fileName, lineNum := extractTemplateErrorInfo(cause.Error()); fileName != "" {
		renderErr.FileName = fileName
		renderErr.LineNumber = lineNum
	}

	return renderErr
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

// IsTemplateErrorType 检查是否为特定类型的模板错误
// errorType 可以是: "NOT_FOUND", "PARSE_ERROR", "RENDER_ERROR", "BLOCK_NOT_FOUND", "VALIDATION_ERROR"
func IsTemplateErrorType(err error, errorType string) bool {
	if te, ok := err.(*TemplateError); ok {
		return te.Type == errorType
	}

	// 兼容标准错误类型
	switch errorType {
	case "NOT_FOUND":
		return errors.Is(err, ErrTemplateNotFound)
	case "PARSE_ERROR":
		return errors.Is(err, ErrTemplateParseError)
	case "RENDER_ERROR":
		return errors.Is(err, ErrTemplateRenderError)
	case "BLOCK_NOT_FOUND":
		return errors.Is(err, ErrBlockNotFound)
	}

	return false
}

// IsTemplateNotFoundError 检查是否为模板未找到错误
func IsTemplateNotFoundError(err error) bool {
	return IsTemplateErrorType(err, "NOT_FOUND")
}

// IsTemplateParseError 检查是否为模板解析错误
func IsTemplateParseError(err error) bool {
	return IsTemplateErrorType(err, "PARSE_ERROR")
}

// IsTemplateRenderError 检查是否为模板渲染错误
func IsTemplateRenderError(err error) bool {
	return IsTemplateErrorType(err, "RENDER_ERROR")
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

// extractTemplateErrorInfo 从模板错误信息中提取文件名和行号并解析为完整路径
// Go template 错误格式: "template: filename.html:10: error message"
func extractTemplateErrorInfo(errMsg string) (fullPath string, lineNum int) {
	// 匹配 "template: filename.html:10:" 格式
	// 使用缓存的正则表达式
	re := getTemplateErrorRegex()
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) >= 3 {
		fileName := matches[1]
		lineNum, _ = strconv.Atoi(matches[2])

		// 解析为完整路径
		fullPath = resolveTemplateFilePath(fileName)
		return fullPath, lineNum
	}
	return "", 0
}

// resolveTemplateFilePath 将模板文件名解析为完整路径
func resolveTemplateFilePath(fileName string) string {
	if fileName == "" {
		return ""
	}

	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(fileName) {
		return fileName
	}

	cfg := config.MustFetch()
	pwd, _ := os.Getwd()

	// 尝试从模板目录构建路径（支持子目录，如 layouts/main.html）
	templatePath := filepath.Join(pwd, cfg.Template.Path, fileName)
	if _, err := os.Stat(templatePath); err == nil {
		return templatePath
	}

	// 如果文件名不包含路径分隔符，尝试从 layouts 目录查找
	if !strings.Contains(fileName, "/") && !strings.Contains(fileName, string(filepath.Separator)) {
		layoutPath := filepath.Join(pwd, cfg.Template.Path, cfg.Template.LayoutDir, fileName)
		if _, err := os.Stat(layoutPath); err == nil {
			return layoutPath
		}
	}

	// 尝试从当前工作目录构建路径
	fullPath := filepath.Join(pwd, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	}

	return fileName
}
