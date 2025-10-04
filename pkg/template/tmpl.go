package template

import (
	"html/template"
	"net/http"
	"runtime/debug"

	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"go.uber.org/zap"
)

// 向后兼容的全局变量
var tmplManager Manager

// InitTemplateManager 初始化模板管理器(向后兼容)
func InitTemplateManager(cfg config.TemplateConfig, isDevelopment bool) Manager {
	tmplManager = NewTemplateManager(cfg, isDevelopment)
	return tmplManager
}

// 获取管理器实例(内部使用)
func getManager() Manager {
	if tmplManager == nil {
		panic("模板管理器未初始化，请先调用 InitTemplateManager")
	}
	return tmplManager
}

// ==================== 渲染 API ====================

// Render 渲染模板，支持可选布局参数
// 不传 layout 参数则不使用布局，传入布局名称则使用指定布局
// 自动处理错误：开发模式显示详细堆栈，生产模式显示通用错误页
//
// 示例：
//
//	template.Render(w, "index", data)              // 不使用布局
//	template.Render(w, "index", data, "main")      // 使用 main 布局
//	template.Render(w, "index", data, "admin")     // 使用 admin 布局
func Render(w http.ResponseWriter, name string, data any, layout ...string) {
	err := getManager().Render(w, name, data, layout...)
	if err != nil {
		handleHTTPError(w, err)
	}
}

// RenderL 使用默认布局渲染模板
// 这是最常用的函数，推荐在 Controller 中使用
// L = Layout (使用默认布局)
func RenderL(w http.ResponseWriter, name string, data any) {
	err := getManager().RenderWithDefaultLayout(w, name, data)
	if err != nil {
		handleHTTPError(w, err)
	}
}

// RenderBlock 动态加载指定模板文件中的特定块并渲染
// 用于在模板中嵌入其他模板的特定块
func RenderBlock(templatePath, blockName string, data any) template.HTML {
	return getManager().RenderBlock(templatePath, blockName, data)
}

// ==================== 工具函数 ====================

// ClearCache 清除模板缓存
func ClearCache() {
	getManager().ClearCache()
}

// IsDevelopmentMode 检查当前是否为开发模式
func IsDevelopmentMode() bool {
	tm := GetTemplateManager()
	if tm != nil {
		return tm.developmentMode
	}
	return false
}

// GetTemplateManager 获取底层的 TemplateManager 实例(用于高级操作)
func GetTemplateManager() *TemplateManager {
	manager := getManager()
	if tm, ok := manager.(*TemplateManager); ok {
		return tm
	}
	return nil
}

// ==================== HTTP 错误处理(内部函数) ====================

// handleHTTPError 处理 HTTP 渲染错误
// 开发模式：显示详细错误堆栈到浏览器 + 控制台
// 生产模式：显示通用错误页面 + 详细日志到控制台/文件
func handleHTTPError(w http.ResponseWriter, err error) {
	manager := getManager()

	// 获取类型断言后的 TemplateManager，以访问 developmentMode
	tm, ok := manager.(*TemplateManager)
	if !ok {
		// 降级处理：无法获取 developmentMode，使用保守策略
		logger.Error("模板渲染错误", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	isDev := tm.developmentMode
	if !isDev {
		logger.Error("模板渲染错误", zap.Error(err))
	}

	// 使用统一的错误渲染页面
	errors.RenderError(w, err, string(debug.Stack()), isDev)
}
