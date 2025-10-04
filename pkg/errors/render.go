package errors

import (
	"bufio"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/gorilla-go/go-framework/pkg/config"
)

// CodeLine 代码行
type CodeLine struct {
	Number  int
	Content string
	IsError bool
}

// RenderError 渲染 HTTP 错误到浏览器（用于 Recovery 中间件）
func RenderError(w http.ResponseWriter, err error, stack string, isDevelopment bool) {
	if !isDevelopment {
		// 生产模式：显示通用错误页面
		renderProductionError(w)
		return
	}

	// 开发模式：显示详细错误信息
	renderDevelopmentError(w, err, stack)
}

// ExtractFileAndLine 从错误中提取文件和行号
func ExtractFileAndLine(err error, stack string) (string, int) {
	if err == nil {
		return "", 0
	}

	// 优先尝试从 TemplateError 结构体中提取文件名和行号（使用反射避免循环依赖）
	if file, line := extractFromTemplateError(err); file != "" && line > 0 {
		return file, line
	}

	errMsg := err.Error()
	cfg := config.MustFetch()

	// 优先检查是否为模板错误（从错误消息中解析）
	// Go template 错误格式: "template: test-error.html:10: ..."
	// 使用共享的提取函数
	if file, line := extractTemplateErrorInfo(errMsg); file != "" && line > 0 {
		return file, line
	}

	// 匹配其他格式: /path/to/file.go:123 或 template.html:123
	// 支持 .go 文件（Go 代码错误）和 .html 文件（模板错误）
	re := regexp.MustCompile(`([/\w\-_.]+\.(?:go|` + cfg.Template.Extension + `)):(\d+)`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) >= 3 {
		file := matches[1]
		line := 0
		fmt.Sscanf(matches[2], "%d", &line)

		// 如果是模板文件，构建完整路径
		if strings.HasSuffix(file, "."+cfg.Template.Extension) && !filepath.IsAbs(file) {
			pwd, _ := os.Getwd()
			templatePath := filepath.Join(pwd, cfg.Template.Path, file)
			if _, err := os.Stat(templatePath); err == nil {
				return templatePath, line
			}
		}

		return resolveFilePath(file), line
	}

	// 从堆栈跟踪中提取用户代码位置
	lines := strings.Split(stack, "\n")
	for i := range len(lines) {
		line := strings.TrimSpace(lines[i])

		// 查找文件位置行 (通常以 / 开头或包含 .go)
		matches := re.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		file := matches[1]
		lineNum := 0
		fmt.Sscanf(matches[2], "%d", &lineNum)

		// 跳过 runtime 和框架内部文件
		if strings.Contains(file, "/runtime/") ||
			strings.Contains(file, "recovery.go") ||
			strings.Contains(file, "panic.go") {
			continue
		}

		// 或者返回第一个非框架文件
		return resolveFilePath(file), lineNum
	}

	return "", 0
}

// resolveFilePath 解析文件路径为绝对路径
func resolveFilePath(file string) string {
	if file == "" {
		return ""
	}

	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(file) {
		return file
	}

	// 尝试从当前工作目录构建完整路径
	pwd, _ := os.Getwd()
	fullPath := filepath.Join(pwd, file)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	}

	return file
}

// ReadCodeContext 读取代码上下文（错误行前后几行）
func ReadCodeContext(file string, errorLine int, contextLines int) []CodeLine {
	if file == "" || errorLine <= 0 {
		return nil
	}

	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []CodeLine
	scanner := bufio.NewScanner(f)
	lineNum := 1

	startLine := max(errorLine-contextLines, 1)
	endLine := errorLine + contextLines

	for scanner.Scan() {
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, CodeLine{
				Number:  lineNum,
				Content: scanner.Text(),
				IsError: lineNum == errorLine,
			})
		}
		lineNum++
		if lineNum > endLine {
			break
		}
	}

	return lines
}

// extractFromTemplateError 从 TemplateError 结构体中提取文件名和行号（使用反射避免循环依赖）
func extractFromTemplateError(err error) (string, int) {
	errType := fmt.Sprintf("%T", err)
	if !strings.Contains(errType, "TemplateError") {
		return "", 0
	}

	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", 0
	}

	fileNameField := v.FieldByName("FileName")
	lineNumberField := v.FieldByName("LineNumber")
	causeField := v.FieldByName("Cause")

	if !fileNameField.IsValid() || !lineNumberField.IsValid() {
		return "", 0
	}

	fileName := fileNameField.String()
	lineNumber := int(lineNumberField.Int())

	// 如果 FileName 和 LineNumber 已经设置，直接返回
	if fileName != "" && lineNumber > 0 {
		// 使用共享的路径解析函数
		fullPath := resolveTemplateFilePath(fileName)
		return fullPath, lineNumber
	}

	// 如果 FileName 为空，尝试从 Cause 错误中提取
	if causeField.IsValid() && !causeField.IsNil() {
		cause := causeField.Interface().(error)
		if cause != nil {
			return extractTemplateErrorInfo(cause.Error())
		}
	}

	return "", 0
}

// formatStackTrace 格式化堆栈跟踪，使其更易读
func formatStackTrace(stack string) string {
	lines := strings.Split(stack, "\n")
	var formatted strings.Builder

	for i := range len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// 函数调用行（不以/开头，不以+开头）
		if !strings.HasPrefix(trimmed, "/") && !strings.HasPrefix(trimmed, "+") && trimmed != "" {
			formatted.WriteString(`<div class="stack-function">`)
			formatted.WriteString(html.EscapeString(trimmed))
			formatted.WriteString("</div>\n")
		} else if strings.HasPrefix(trimmed, "/") || strings.Contains(trimmed, ".go:") {
			// 文件位置行
			formatted.WriteString(`<div class="stack-location">`)
			formatted.WriteString(`  `)
			formatted.WriteString(html.EscapeString(trimmed))
			formatted.WriteString("</div>\n")
		}
	}

	return formatted.String()
}

// renderDevelopmentError 渲染开发模式错误页面
func renderDevelopmentError(w http.ResponseWriter, err error, stack string) {
	// 解析错误信息
	errorType := "Runtime Error"
	errorMessage := err.Error()
	fileName := ""
	lineInfo := ""

	// 从堆栈中提取文件和行号
	file, line := ExtractFileAndLine(err, stack)
	if file != "" {
		fileName = file
		lineInfo = fmt.Sprintf("%s:%d", file, line)
	}

	// 读取代码上下文（错误行前后 5 行）
	var codeContext []CodeLine
	if fileName != "" && line > 0 {
		codeContext = ReadCodeContext(fileName, line, 5)
	}

	// 格式化堆栈跟踪
	formattedStack := formatStackTrace(stack)

	// 构建代码上下文的 HTML
	codeContextHTML := ""
	if len(codeContext) > 0 {
		codeContextHTML = `<div class="error-section">
			<div class="section-title">📝 代码上下文</div>
			<div class="code-context">`

		for _, codeLine := range codeContext {
			lineClass := "code-line"
			if codeLine.IsError {
				lineClass = "code-line error-line"
			}
			codeContextHTML += fmt.Sprintf(`
				<div class="%s">
					<span class="line-number">%d</span>
					<span class="line-content">%s</span>
				</div>`,
				lineClass,
				codeLine.Number,
				html.EscapeString(codeLine.Content),
			)
		}

		codeContextHTML += `</div></div>`
	}

	errorHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Application Error - Development Mode</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, 'Microsoft YaHei', sans-serif;
            background: #1e1e1e;
            color: #d4d4d4;
            padding: 20px;
            line-height: 1.6;
        }
        .error-container {
            max-width: 1200px;
            margin: 0 auto;
            background: #252526;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
        }
        .error-header {
            background: linear-gradient(135deg, #f14c4c 0%%, #c92a2a 100%%);
            padding: 30px;
            color: white;
        }
        .error-icon {
            font-size: 48px;
            margin-bottom: 15px;
        }
        .error-title {
            font-size: 28px;
            font-weight: 600;
            margin-bottom: 10px;
        }
        .error-subtitle {
            font-size: 14px;
            opacity: 0.9;
        }
        .error-content {
            padding: 30px;
        }
        .error-section {
            margin-bottom: 25px;
        }
        .section-title {
            color: #4ec9b0;
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 10px;
            padding-bottom: 8px;
            border-bottom: 2px solid #3e3e42;
        }
        .error-message {
            background: #2d2d30;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #f14c4c;
            color: #ce9178;
            font-family: 'Consolas', 'Monaco', monospace;
            font-size: 14px;
            word-break: break-word;
        }
        .stack-trace {
            background: #1e1e1e;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
            font-family: 'Consolas', 'Monaco', monospace;
            font-size: 12px;
            line-height: 1.8;
            max-height: 500px;
            overflow-y: auto;
        }
        .stack-trace::-webkit-scrollbar {
            width: 10px;
            height: 10px;
        }
        .stack-trace::-webkit-scrollbar-track {
            background: #2d2d30;
        }
        .stack-trace::-webkit-scrollbar-thumb {
            background: #3e3e42;
            border-radius: 5px;
        }
        .stack-trace::-webkit-scrollbar-thumb:hover {
            background: #4e4e52;
        }
        .stack-function {
            color: #dcdcaa;
            margin-top: 10px;
            padding: 4px 0;
            font-weight: 500;
        }
        .stack-location {
            color: #858585;
            padding-left: 20px;
            font-size: 11px;
            margin: 2px 0;
        }
        .stack-location:hover {
            color: #9cdcfe;
            background: rgba(255, 255, 255, 0.05);
        }
        .badge {
            display: inline-block;
            background: #f14c4c;
            color: white;
            padding: 4px 12px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
            margin-bottom: 15px;
        }
        .help-text {
            background: #2d2d30;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #4ec9b0;
            color: #9cdcfe;
            font-size: 13px;
        }
        .help-text strong {
            color: #4ec9b0;
        }
        .file-location {
            background: #2d2d30;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #ffa500;
            margin-bottom: 15px;
        }
        .file-location .label {
            color: #ffa500;
            font-weight: 600;
            font-size: 12px;
            text-transform: uppercase;
            margin-bottom: 8px;
        }
        .file-location .path {
            font-family: 'Consolas', 'Monaco', monospace;
            font-size: 14px;
            color: #dcdcaa;
            word-break: break-all;
        }
        .code-context {
            background: #1e1e1e;
            border-radius: 5px;
            overflow: hidden;
            font-family: 'Consolas', 'Monaco', monospace;
            font-size: 13px;
            line-height: 1.6;
        }
        .code-line {
            display: flex;
            padding: 4px 0;
            border-left: 3px solid transparent;
        }
        .code-line:hover {
            background: #2d2d30;
        }
        .code-line.error-line {
            background: rgba(255, 76, 76, 0.15);
            border-left-color: #f14c4c;
        }
        .code-line.error-line .line-number {
            color: #f14c4c;
            font-weight: bold;
        }
        .line-number {
            color: #858585;
            padding: 0 15px;
            text-align: right;
            user-select: none;
            min-width: 60px;
        }
        .line-content {
            flex: 1;
            padding-right: 15px;
            color: #d4d4d4;
            white-space: pre;
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-header">
            <div class="error-icon">⚠️</div>
            <div class="error-title">致命错误</div>
            <div class="error-subtitle">开发模式 - 详细错误信息</div>
        </div>

        <div class="error-content">
            <div class="badge">%s</div>

            <div class="error-section">
                <div class="section-title">💥 错误信息</div>
                <div class="error-message">%s</div>
            </div>

            %s

            %s

            <div class="error-section">
                <div class="section-title">🔍 完整堆栈跟踪</div>
                <div class="stack-trace">%s</div>
            </div>

            <div class="error-section">
                <div class="help-text">
                    <strong>💡 提示:</strong> 此错误页面仅在开发模式下显示。 生产模式不可见。
                </div>
            </div>
        </div>
    </div>
</body>
</html>`,
		html.EscapeString(errorType),
		html.EscapeString(errorMessage),
		func() string {
			if fileName != "" && lineInfo != "" {
				return fmt.Sprintf(`<div class="file-location">
					<div class="label">📂 错误位置</div>
					<div class="path">%s</div>
				</div>`, html.EscapeString(lineInfo))
			}
			return ""
		}(),
		codeContextHTML,
		formattedStack,
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(errorHTML))
}

// renderProductionError 渲染生产模式错误页面
func renderProductionError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
