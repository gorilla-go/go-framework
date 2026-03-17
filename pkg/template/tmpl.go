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

// 全局模板管理器
var tmplManager *TemplateManager

// InitTemplateManager 初始化全局模板管理器
func InitTemplateManager(cfg config.TemplateConfig, isDevelopment bool) Manager {
	tmplManager = NewTemplateManager(cfg, isDevelopment)
	return tmplManager
}

// getManager 获取全局管理器实例
func getManager() *TemplateManager {
	if tmplManager == nil {
		panic("模板管理器未初始化，请先调用 InitTemplateManager")
	}
	return tmplManager
}

// ==================== 渲染 API ====================

// Render 渲染模板，支持可选布局参数
// 不传 layout 参数则不使用布局，传入布局名称则使用指定布局
//
// 示例：
//
//	template.Render(w, "index", data)              // 不使用布局
//	template.Render(w, "index", data, "main")      // 使用 main 布局
func Render(w http.ResponseWriter, name string, data any, layout ...string) {
	err := getManager().Render(w, name, data, layout...)
	if err != nil {
		handleHTTPError(w, err)
	}
}

// RenderL 使用默认布局渲染模板（推荐在 Controller 中使用）
func RenderL(w http.ResponseWriter, name string, data any) {
	err := getManager().RenderWithDefaultLayout(w, name, data)
	if err != nil {
		handleHTTPError(w, err)
	}
}

// RenderBlock 动态加载指定模板文件中的特定块并渲染
func RenderBlock(templatePath, blockName string, data any) template.HTML {
	return getManager().RenderBlock(templatePath, blockName, data)
}

// ==================== 工具函数 ====================

// ClearCache 清除模板缓存
func ClearCache() {
	getManager().ClearCache()
}

// ==================== HTTP 错误处理（内部函数）====================

func handleHTTPError(w http.ResponseWriter, err error) {
	tm := getManager()
	isDev := tm.developmentMode
	if !isDev {
		logger.Error("模板渲染错误", zap.Error(err))
	}
	errors.RenderError(w, err, string(debug.Stack()), isDev)
}
