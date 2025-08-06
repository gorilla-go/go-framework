package template

import (
	"fmt"
	"html/template"
	"io"
	"maps"
	"path/filepath"
	"strings"
	"sync"
)

// TemplateManager 模板管理器，负责模板的加载、缓存和渲染
type TemplateManager struct {
	templatesDir    string
	layoutsDir      string
	extension       string
	templates       map[string]*template.Template
	funcMap         template.FuncMap
	mutex           sync.RWMutex
	defaultLayout   string
	developmentMode bool
}

// NewTemplateManager 创建一个新的模板管理器
func NewTemplateManager(templatesDir, layoutsDir, extension string, isDevelopment bool) *TemplateManager {
	return &TemplateManager{
		templatesDir:    templatesDir,
		layoutsDir:      layoutsDir,
		extension:       extension,
		templates:       make(map[string]*template.Template),
		funcMap:         FuncMap(),
		defaultLayout:   "main",
		developmentMode: isDevelopment,
	}
}

// SetDevelopmentMode 设置开发模式
// 在开发模式下，每次渲染模板都会重新加载模板文件
func (tm *TemplateManager) SetDevelopmentMode(mode bool) {
	tm.developmentMode = mode
}

// SetDefaultLayout 设置默认布局
func (tm *TemplateManager) SetDefaultLayout(layout string) {
	tm.defaultLayout = layout
}

// AddFunc 添加自定义模板函数
func (tm *TemplateManager) AddFunc(name string, fn any) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.funcMap[name] = fn
}

// AddFuncs 添加多个自定义模板函数
func (tm *TemplateManager) AddFuncs(funcs template.FuncMap) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	maps.Copy(tm.funcMap, funcs)
}

// loadTemplate 加载模板
func (tm *TemplateManager) loadTemplate(names ...string) (*template.Template, error) {
	var tmpl *template.Template
	var err error
	var ok bool

	// 生成缓存键，包含所有模板名称
	cacheKey := strings.Join(names, ":")

	// 开发模式下不使用缓存，每次都重新加载模板
	if !tm.developmentMode {
		// 尝试从缓存中获取模板
		tm.mutex.RLock()
		tmpl, ok = tm.templates[cacheKey]
		tm.mutex.RUnlock()

		// 如果在缓存中找到，直接返回
		if ok {
			return tmpl, nil
		}
	}

	// 如果没有指定任何模板，返回错误
	if len(names) == 0 {
		return nil, fmt.Errorf("没有指定任何模板文件")
	}

	// 需要加载的所有模板文件路径
	var allTemplateFiles []string

	// 处理所有指定的模板
	for _, name := range names {
		if len(name) == 0 {
			continue
		}
		allTemplateFiles = append(allTemplateFiles, filepath.Join(tm.templatesDir, name+tm.extension))
	}

	// 确定主模板名称（基础模板）- 使用第一个模板作为基础
	baseTemplateName := filepath.Base(allTemplateFiles[0])

	// 创建带函数的基础模板
	tmpl = template.New(baseTemplateName).Funcs(tm.funcMap)

	// 解析所有模板文件
	tmpl, err = tmpl.ParseFiles(allTemplateFiles...)
	if err != nil {
		return nil, fmt.Errorf("解析模板文件失败: %w", err)
	}

	// 非开发模式下缓存模板
	if !tm.developmentMode {
		tm.mutex.Lock()
		tm.templates[cacheKey] = tmpl
		tm.mutex.Unlock()
	}

	return tmpl, nil
}

// Render 渲染模板
func (tm *TemplateManager) Render(w io.Writer, name string, data any, layout string) error {
	var templateNames []string

	// 如果指定了布局，添加布局模板
	if layout != "" {
		layoutName := filepath.Join("layouts", layout)
		templateNames = append(templateNames, layoutName)
	}

	// 添加内容模板
	templateNames = append(templateNames, name)

	// 加载并渲染模板
	tmpl, err := tm.loadTemplate(templateNames...)
	if err != nil {
		return tm.handleError(w, err)
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		return tm.handleError(w, err)
	}
	return nil
}

// handleError 处理模板错误
func (tm *TemplateManager) handleError(w io.Writer, err error) error {
	// 如果需要显示错误，则渲染错误信息
	if tm.developmentMode {
		errorHTML := fmt.Sprintf(`
		<div style="color:rgb(211, 50, 66); background-color: #f8d7da; border: 1px solid #f5c6cb; padding: 15px; margin: 15px; border-radius: 4px;">
			<h3 style="margin-top: 0;">模板渲染错误</h3>
			<pre style="background-color: #f8f9fa; padding: 10px; border-radius: 4px; overflow: auto;">%s</pre>
		</div>`, template.HTMLEscapeString(err.Error()))
		_, writeErr := w.Write([]byte(errorHTML))
		if writeErr != nil {
			return fmt.Errorf("原始错误: %v, 写入错误页面失败: %v", err, writeErr)
		}
		return nil
	}
	return err
}

// RenderWithDefaultLayout 使用默认布局渲染模板
func (tm *TemplateManager) RenderWithDefaultLayout(w io.Writer, name string, data any) error {
	// 如果有默认布局，则同时加载布局和内容模板
	if tm.defaultLayout != "" {
		return tm.Render(w, name, data, tm.defaultLayout)
	}
	// 否则只加载内容模板
	panic("没有默认布局")
}

// RenderPartial 渲染部分模板（不使用布局）
func (tm *TemplateManager) RenderPartial(w io.Writer, name string, data any) error {
	// 使用loadTemplate加载模板
	tmpl, err := tm.loadTemplate(name)
	if err != nil {
		return tm.handleError(w, err)
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		return tm.handleError(w, err)
	}
	return nil
}

// RenderWithoutLayout 渲染不带布局的模板
func (tm *TemplateManager) RenderWithoutLayout(w io.Writer, name string, data any) error {
	return tm.RenderPartial(w, name, data)
}

// ClearCache 清除模板缓存
func (tm *TemplateManager) ClearCache() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.templates = make(map[string]*template.Template)
}

// InitGlobalTemplateManager 初始化全局模板管理器
// 这确保在使用RenderBlock等函数时，全局templateManager已被正确设置
func InitGlobalTemplateManager(tm *TemplateManager) {
	templateManager = tm
}

// RenderMultiple 渲染多个模板
func (tm *TemplateManager) RenderMultiple(w io.Writer, data any, names ...string) error {
	// 加载多个模板
	tmpl, err := tm.loadTemplate(names...)
	if err != nil {
		return tm.handleError(w, err)
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		return tm.handleError(w, err)
	}
	return nil
}
