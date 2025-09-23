package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla-go/go-framework/pkg/config"
)

var tmplManager *TemplateManager = nil

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
func InitTemplateManager(cfg config.TemplateConfig, isDevelopment bool) *TemplateManager {
	tmplManager = &TemplateManager{
		templatesDir:    cfg.Path,
		layoutsDir:      cfg.Layouts,
		extension:       cfg.Extension,
		templates:       make(map[string]*template.Template),
		funcMap:         FuncMap(),
		defaultLayout:   "main",
		developmentMode: isDevelopment,
	}
	return tmplManager
}

// loadTemplate 加载模板
func loadTemplate(names ...string) (*template.Template, error) {
	var tmpl *template.Template
	var err error
	var ok bool

	// 生成缓存键，包含所有模板名称
	cacheKey := strings.Join(names, ":")

	// 开发模式下不使用缓存，每次都重新加载模板
	if !tmplManager.developmentMode {
		// 尝试从缓存中获取模板
		tmplManager.mutex.RLock()
		tmpl, ok = tmplManager.templates[cacheKey]
		tmplManager.mutex.RUnlock()

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
		allTemplateFiles = append(allTemplateFiles, filepath.Join(tmplManager.templatesDir, name+tmplManager.extension))
	}

	// 确定主模板名称（基础模板）- 使用第一个模板作为基础
	baseTemplateName := filepath.Base(allTemplateFiles[0])

	// 创建带函数的基础模板
	tmpl = template.New(baseTemplateName).Funcs(tmplManager.funcMap)

	// 解析所有模板文件
	tmpl, err = tmpl.ParseFiles(allTemplateFiles...)
	if err != nil {
		return nil, fmt.Errorf("解析模板文件失败: %w", err)
	}

	// 非开发模式下缓存模板
	if !tmplManager.developmentMode {
		tmplManager.mutex.Lock()
		tmplManager.templates[cacheKey] = tmpl
		tmplManager.mutex.Unlock()
	}

	return tmpl, nil
}

// Render 渲染模板，支持可选布局参数
func Render(w io.Writer, name string, data any, layout ...string) error {
	var templateNames []string

	// 处理布局参数
	if len(layout) > 0 && layout[0] != "" {
		templateNames = append(templateNames, filepath.Join("layouts", layout[0]))
	}

	// 添加内容模板
	templateNames = append(templateNames, name)

	// 加载并渲染模板
	tmpl, err := loadTemplate(templateNames...)
	if err != nil {
		return renderError(w, err)
	}
	return tmpl.Execute(w, data)
}

// RenderWithDefaultLayout 使用默认布局渲染模板
func RenderWithDefaultLayout(w io.Writer, name string, data any) error {
	return Render(w, name, data, tmplManager.defaultLayout)
}

// RenderPartial 渲染部分模板（不使用布局）
func RenderPartial(w io.Writer, name string, data any) error {
	return Render(w, name, data)
}

// RenderWithoutLayout 渲染不带布局的模板
func RenderWithoutLayout(w io.Writer, name string, data any) error {
	return Render(w, name, data)
}

// 内部错误处理函数
func renderError(w io.Writer, err error) error {
	if !tmplManager.developmentMode {
		return err
	}
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

// ClearCache 清除模板缓存
func ClearCache() {
	tmplManager.mutex.Lock()
	defer tmplManager.mutex.Unlock()

	tmplManager.templates = make(map[string]*template.Template)
}

// RenderMultiple 渲染多个模板
func RenderMultiple(w io.Writer, data any, names ...string) error {
	tmpl, err := loadTemplate(names...)
	if err != nil {
		return renderError(w, err)
	}
	return tmpl.Execute(w, data)
}

// RenderBlock 动态加载指定模板文件中的特定块并渲染
func RenderBlock(templatePath, blockName string, data any) template.HTML {
	var buf strings.Builder
	tmpl, err := loadTemplate(templatePath)
	if err != nil {
		return template.HTML(
			fmt.Sprintf(
				`<div class="error">%s</div>`,
				template.HTMLEscapeString(err.Error()),
			),
		)
	}

	if block := tmpl.Lookup(blockName); block != nil {
		if err := block.Execute(&buf, data); err != nil {
			return template.HTML(
				fmt.Sprintf(
					`<div class="error">%s</div>`,
					template.HTMLEscapeString(err.Error()),
				),
			)
		}
		return template.HTML(buf.String())
	}
	return template.HTML(fmt.Sprintf(`<div class="error">找不到块 '%s'</div>`, blockName))
}
