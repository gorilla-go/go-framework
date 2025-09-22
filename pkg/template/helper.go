package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/logger"
)

var Template *TemplateManager = nil

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
	Template = &TemplateManager{
		templatesDir:    cfg.Path,
		layoutsDir:      cfg.Layouts,
		extension:       cfg.Extension,
		templates:       make(map[string]*template.Template),
		funcMap:         FuncMap(),
		defaultLayout:   "main",
		developmentMode: isDevelopment,
	}
	return Template
}

// loadTemplate 加载模板
func loadTemplate(names ...string) (*template.Template, error) {
	var tmpl *template.Template
	var err error
	var ok bool

	// 生成缓存键，包含所有模板名称
	cacheKey := strings.Join(names, ":")

	// 开发模式下不使用缓存，每次都重新加载模板
	if !Template.developmentMode {
		// 尝试从缓存中获取模板
		Template.mutex.RLock()
		tmpl, ok = Template.templates[cacheKey]
		Template.mutex.RUnlock()

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
		allTemplateFiles = append(allTemplateFiles, filepath.Join(Template.templatesDir, name+Template.extension))
	}

	// 确定主模板名称（基础模板）- 使用第一个模板作为基础
	baseTemplateName := filepath.Base(allTemplateFiles[0])

	// 创建带函数的基础模板
	tmpl = template.New(baseTemplateName).Funcs(Template.funcMap)

	// 解析所有模板文件
	tmpl, err = tmpl.ParseFiles(allTemplateFiles...)
	if err != nil {
		return nil, fmt.Errorf("解析模板文件失败: %w", err)
	}

	// 非开发模式下缓存模板
	if !Template.developmentMode {
		Template.mutex.Lock()
		Template.templates[cacheKey] = tmpl
		Template.mutex.Unlock()
	}

	return tmpl, nil
}

// Render 渲染模板
func Render(w io.Writer, name string, data any, layout string) error {
	var templateNames []string

	// 如果指定了布局，添加布局模板
	if layout != "" {
		layoutName := filepath.Join("layouts", layout)
		templateNames = append(templateNames, layoutName)
	}

	// 添加内容模板
	templateNames = append(templateNames, name)

	// 加载并渲染模板
	tmpl, err := loadTemplate(templateNames...)
	if err != nil {
		return handleError(w, err)
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		return handleError(w, err)
	}
	return nil
}

// handleError 处理模板错误
func handleError(w io.Writer, err error) error {
	// 如果需要显示错误，则渲染错误信息
	if Template.developmentMode {
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
func RenderWithDefaultLayout(w io.Writer, name string, data any) error {
	// 如果有默认布局，则同时加载布局和内容模板
	if Template.defaultLayout != "" {
		return Render(w, name, data, Template.defaultLayout)
	}
	// 否则只加载内容模板
	logger.Fatal("没有默认布局")
	return nil
}

// RenderPartial 渲染部分模板（不使用布局）
func RenderPartial(w io.Writer, name string, data any) error {
	// 使用loadTemplate加载模板
	tmpl, err := loadTemplate(name)
	if err != nil {
		return handleError(w, err)
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		return handleError(w, err)
	}
	return nil
}

// RenderWithoutLayout 渲染不带布局的模板
func RenderWithoutLayout(w io.Writer, name string, data any) error {
	return RenderPartial(w, name, data)
}

// ClearCache 清除模板缓存
func ClearCache() {
	Template.mutex.Lock()
	defer Template.mutex.Unlock()

	Template.templates = make(map[string]*template.Template)
}

// RenderMultiple 渲染多个模板
func RenderMultiple(w io.Writer, data any, names ...string) error {
	// 加载多个模板
	tmpl, err := loadTemplate(names...)
	if err != nil {
		return handleError(w, err)
	}

	// 渲染模板
	err = tmpl.Execute(w, data)
	if err != nil {
		return handleError(w, err)
	}
	return nil
}

// RenderBlock 动态加载指定模板文件中的特定块(block)并渲染
//
// 模板使用示例:
// {{ render "components/card" "content" .CardData }} <!-- 渲染 components/card.html 中的 "content" 块 -->
func RenderBlock(templatePath, blockName string, data any) template.HTML {
	// 创建一个缓冲区用于存放渲染结果
	var buf strings.Builder

	// 使用一个独立的函数来处理模板渲染，这样可以更好地捕获错误
	err := renderTemplateBlock(templatePath, blockName, data, &buf)
	if err != nil {
		return template.HTML(
			fmt.Sprintf(`<div class="error">渲染模板失败: %s</div>`,
				template.HTMLEscapeString(err.Error())),
		)
	}

	return template.HTML(buf.String())
}

// 辅助函数：用于渲染模板块
func renderTemplateBlock(templatePath, blockName string, data any, buf *strings.Builder) error {
	tmpl, err := loadTemplate(templatePath, "")
	if err != nil {
		return fmt.Errorf("无法解析模板 '%s': %v", templatePath, err)
	}

	// 查找指定的块
	blockTmpl := tmpl.Lookup(blockName)
	if blockTmpl == nil {
		return fmt.Errorf("模板 '%s' 中找不到块 '%s'", templatePath, blockName)
	}

	// 执行模板，写入缓冲区
	if err := blockTmpl.Execute(buf, data); err != nil {
		return fmt.Errorf("渲染块 '%s' 失败: %v", blockName, err)
	}

	return nil
}
