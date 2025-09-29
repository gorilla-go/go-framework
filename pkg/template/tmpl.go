package template

import (
	"html/template"
	"io"

	"github.com/gorilla-go/go-framework/pkg/config"
)

// 向后兼容的全局变量
var tmplManager Manager

// InitTemplateManager 初始化模板管理器（向后兼容）
func InitTemplateManager(cfg config.TemplateConfig, isDevelopment bool) Manager {
	tmplManager = NewTemplateManager(cfg, isDevelopment)
	return tmplManager
}

// 获取管理器实例（内部使用）
func getManager() Manager {
	if tmplManager == nil {
		panic("模板管理器未初始化，请先调用 InitTemplateManager")
	}
	return tmplManager
}

// Render 渲染模板，支持可选布局参数
func Render(w io.Writer, name string, data any, layout ...string) error {
	return getManager().Render(w, name, data, layout...)
}

// RenderWithDefaultLayout 使用默认布局渲染模板
func RenderWithDefaultLayout(w io.Writer, name string, data any) error {
	return getManager().RenderWithDefaultLayout(w, name, data)
}

// RenderPartial 渲染部分模板（不使用布局）
func RenderPartial(w io.Writer, name string, data any) error {
	return getManager().RenderPartial(w, name, data)
}

// RenderWithoutLayout 渲染不带布局的模板（向后兼容）
func RenderWithoutLayout(w io.Writer, name string, data any) error {
	return getManager().RenderPartial(w, name, data)
}

// ClearCache 清除模板缓存
func ClearCache() {
	getManager().ClearCache()
}

// RenderMultiple 渲染多个模板
func RenderMultiple(w io.Writer, data any, names ...string) error {
	return getManager().RenderMultiple(w, data, names...)
}

// RenderBlock 动态加载指定模板文件中的特定块并渲染
func RenderBlock(templatePath, blockName string, data any) template.HTML {
	return getManager().RenderBlock(templatePath, blockName, data)
}
