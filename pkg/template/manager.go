package template

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla-go/go-framework/pkg/config"
)

// Manager 模板管理器接口
type Manager interface {
	Render(w io.Writer, name string, data any, layout ...string) error
	RenderWithDefaultLayout(w io.Writer, name string, data any) error
	RenderPartial(w io.Writer, name string, data any) error
	RenderMultiple(w io.Writer, data any, names ...string) error
	RenderBlock(templatePath, blockName string, data any) template.HTML
	ClearCache()
	SetDevelopmentMode(isDev bool)
	GetTemplateNames() []string
}

// TemplateManager 模板管理器实现
type TemplateManager struct {
	templatesDir    string
	layoutsDir      string
	extension       string
	templates       map[string]*template.Template
	funcMap         template.FuncMap
	mutex           sync.RWMutex
	defaultLayout   string
	developmentMode bool
	loadStats       map[string]int64 // 模板加载统计
	statsMutex      sync.RWMutex
}

// 全局实例
var (
	defaultManager Manager
	managerOnce    sync.Once
)

// NewTemplateManager 创建一个新的模板管理器
func NewTemplateManager(cfg config.TemplateConfig, isDevelopment bool) *TemplateManager {
	return &TemplateManager{
		templatesDir:    cfg.Path,
		layoutsDir:      cfg.Layouts,
		extension:       cfg.Extension,
		templates:       make(map[string]*template.Template),
		funcMap:         FuncMap(),
		defaultLayout:   cfg.DefaultLayout,
		developmentMode: isDevelopment,
		loadStats:       make(map[string]int64),
	}
}

// InitGlobalTemplateManager 初始化全局模板管理器（内部使用）
func InitGlobalTemplateManager(cfg config.TemplateConfig, isDevelopment bool) Manager {
	managerOnce.Do(func() {
		defaultManager = NewTemplateManager(cfg, isDevelopment)
	})
	return defaultManager
}

// GetManager 获取全局模板管理器
func GetManager() Manager {
	if defaultManager == nil {
		panic("模板管理器未初始化，请先调用 InitTemplateManager")
	}
	return defaultManager
}

// SetDevelopmentMode 设置开发模式
func (tm *TemplateManager) SetDevelopmentMode(isDev bool) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.developmentMode = isDev
}

// GetTemplateNames 获取所有已加载的模板名称
func (tm *TemplateManager) GetTemplateNames() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// loadTemplate 加载模板（内部方法）
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
			tm.updateLoadStats(cacheKey)
			return tmpl, nil
		}
	}

	// 如果没有指定任何模板，返回错误
	if len(names) == 0 {
		return nil, NewTemplateError("VALIDATION_ERROR", "没有指定任何模板文件", "", ErrInvalidTemplateName)
	}

	// 需要加载的所有模板文件路径
	var allTemplateFiles []string

	// 处理所有指定的模板
	for _, name := range names {
		if len(name) == 0 {
			continue
		}
		// 验证模板名称
		if err := ValidateTemplateName(name); err != nil {
			return nil, err
		}
		allTemplateFiles = append(allTemplateFiles, filepath.Join(tm.templatesDir, name+tm.extension))
	}

	if len(allTemplateFiles) == 0 {
		return nil, NewTemplateError("VALIDATION_ERROR", "没有找到有效的模板文件", "", ErrInvalidTemplateName)
	}

	// 确定主模板名称（基础模板）- 使用第一个模板作为基础
	baseTemplateName := filepath.Base(allTemplateFiles[0])

	// 创建带函数的基础模板
	tmpl = template.New(baseTemplateName).Funcs(tm.funcMap)

	// 解析所有模板文件
	tmpl, err = tmpl.ParseFiles(allTemplateFiles...)
	if err != nil {
		return nil, NewParseError(strings.Join(names, ":"), err)
	}

	// 非开发模式下缓存模板
	if !tm.developmentMode {
		tm.mutex.Lock()
		tm.templates[cacheKey] = tmpl
		tm.mutex.Unlock()
	}

	tm.updateLoadStats(cacheKey)
	return tmpl, nil
}

// updateLoadStats 更新加载统计
func (tm *TemplateManager) updateLoadStats(cacheKey string) {
	tm.statsMutex.Lock()
	defer tm.statsMutex.Unlock()
	tm.loadStats[cacheKey]++
}

// Render 渲染模板，支持可选布局参数
func (tm *TemplateManager) Render(w io.Writer, name string, data any, layout ...string) error {
	// 验证模板名称
	if err := ValidateTemplateName(name); err != nil {
		return tm.renderError(w, err)
	}

	var templateNames []string

	// 处理布局参数
	if len(layout) > 0 && layout[0] != "" {
		if err := ValidateLayoutName(layout[0]); err != nil {
			return tm.renderError(w, err)
		}
		templateNames = append(templateNames, filepath.Join("layouts", layout[0]))
	}

	// 添加内容模板
	templateNames = append(templateNames, name)

	// 加载并渲染模板
	tmpl, err := tm.loadTemplate(templateNames...)
	if err != nil {
		return tm.renderError(w, err)
	}

	// 在渲染前设置 Content-Type（如果 w 是 http.ResponseWriter 且未设置）
	tm.ensureContentType(w)

	// 执行模板渲染
	if err := tmpl.Execute(w, data); err != nil {
		return tm.renderError(w, NewRenderError(name, err))
	}
	return nil
}

// RenderWithDefaultLayout 使用默认布局渲染模板
func (tm *TemplateManager) RenderWithDefaultLayout(w io.Writer, name string, data any) error {
	return tm.Render(w, name, data, tm.defaultLayout)
}

// ensureContentType 确保设置了 Content-Type（仅对 http.ResponseWriter 有效）
func (tm *TemplateManager) ensureContentType(w io.Writer) {
	// 尝试将 w 转换为 http.ResponseWriter
	type headerWriter interface {
		Header() http.Header
		WriteHeader(int)
	}

	if hw, ok := w.(headerWriter); ok {
		// 检查是否已设置 Content-Type
		if hw.Header().Get("Content-Type") == "" {
			// 设置默认的 HTML Content-Type
			hw.Header().Set("Content-Type", "text/html; charset=utf-8")
			// 设置状态码（如果尚未设置）
			hw.WriteHeader(http.StatusOK)
		}
	}
}

// RenderPartial 渲染部分模板（不使用布局）
func (tm *TemplateManager) RenderPartial(w io.Writer, name string, data any) error {
	return tm.Render(w, name, data)
}

// RenderMultiple 渲染多个模板
func (tm *TemplateManager) RenderMultiple(w io.Writer, data any, names ...string) error {
	tmpl, err := tm.loadTemplate(names...)
	if err != nil {
		return tm.renderError(w, err)
	}
	return tmpl.Execute(w, data)
}

// RenderBlock 动态加载指定模板文件中的特定块并渲染
func (tm *TemplateManager) RenderBlock(templatePath, blockName string, data any) template.HTML {
	// 验证参数
	if err := ValidateTemplateName(templatePath); err != nil {
		return tm.renderBlockError(err)
	}
	if blockName == "" {
		return tm.renderBlockError(NewTemplateError("VALIDATION_ERROR", "块名称不能为空", templatePath, nil))
	}

	var buf strings.Builder
	tmpl, err := tm.loadTemplate(templatePath)
	if err != nil {
		return tm.renderBlockError(err)
	}

	if block := tmpl.Lookup(blockName); block != nil {
		if err := block.Execute(&buf, data); err != nil {
			return tm.renderBlockError(NewRenderError(templatePath, err))
		}
		return template.HTML(buf.String())
	}
	return tm.renderBlockError(NewBlockNotFoundError(templatePath, blockName))
}

// renderBlockError 渲染块错误信息
func (tm *TemplateManager) renderBlockError(err error) template.HTML {
	if !tm.developmentMode {
		// 生产模式下返回空内容或占位符
		return template.HTML(`<div class="template-error-placeholder"></div>`)
	}
	return template.HTML(
		fmt.Sprintf(
			`<div class="template-error" style="color: #d13212; background-color: #f8d7da; border: 1px solid #f5c6cb; padding: 10px; margin: 5px; border-radius: 3px;">%s</div>`,
			template.HTMLEscapeString(err.Error()),
		),
	)
}

// renderError 内部错误处理函数
func (tm *TemplateManager) renderError(w io.Writer, err error) error {
	if !tm.developmentMode {
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
func (tm *TemplateManager) ClearCache() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.templates = make(map[string]*template.Template)

	tm.statsMutex.Lock()
	tm.loadStats = make(map[string]int64)
	tm.statsMutex.Unlock()
}

// GetLoadStats 获取模板加载统计信息
func (tm *TemplateManager) GetLoadStats() map[string]int64 {
	tm.statsMutex.RLock()
	defer tm.statsMutex.RUnlock()

	stats := make(map[string]int64, len(tm.loadStats))
	for k, v := range tm.loadStats {
		stats[k] = v
	}
	return stats
}