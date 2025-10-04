// Package template 提供用于HTML模板的辅助函数
package template

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla-go/go-framework/pkg/router"
)

// 预编译的正则表达式，避免重复编译
var (
	htmlTagRegex *regexp.Regexp
	regexOnce    sync.Once
)

// 初始化预编译的正则表达式
func initRegex() {
	regexOnce.Do(func() {
		htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	})
}

// 最常用的模板函数集合
// FuncMap 返回可用于HTML模板的函数映射
func FuncMap() template.FuncMap {
	return template.FuncMap{
		// 字符串处理（最常用）
		"trim":      strings.TrimSpace,
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"title":     strings.Title,
		"replace":   strings.Replace,
		"split":     strings.Split,
		"join":      strings.Join,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"substr":    Substr,
		"truncate":  Truncate,
		"nl2br":     Nl2br,
		"stripTags": StripTags,

		// 数值处理（最常用）
		"add":      Add,
		"subtract": Subtract,
		"multiply": Multiply,
		"divide":   Divide,
		"mod":      Mod,
		"round":    Round,

		// 日期时间处理（最常用）
		"now":            Now,
		"formatDateTime": FormatDateTime,
		"formatDate":     FormatDate,
		"dateFormat":     DateFormat,
		"humanizeTime":   HumanizeTime,

		// 集合处理（最常用）
		"first":    First,
		"last":     Last,
		"empty":    Empty,
		"notEmpty": NotEmpty,
		"length":   Length,
		"inArray":  InArray,

		// Map处理函数
		"map":     NewMap,
		"mapGet":  MapGet,
		"mapHas":  MapHas,
		"mapKeys": MapKeys,
		"mapSet":  MapSet,

		// 条件处理（最常用）
		"default": Default,
		"ternary": Ternary,
		"eq":      Eq,
		"ne":      Ne,
		"lt":      Lt,
		"lte":     Lte,
		"gt":      Gt,
		"gte":     Gte,

		// 安全处理（最常用）
		"safeHTML": SafeHTML,
		"safeJS":   SafeJS,
		"safeCSS":  SafeCSS,
		"safeURL":  SafeURL,

		// URL处理
		"url": Route, // 简单URL生成函数

		// 块处理
		"render": func(templatePath, blockName string, data any) template.HTML {
			return RenderBlock(templatePath, blockName, data)
		},

		// 错误处理
		"panic": Panic,

		// 调试函数
		"dump": Dump,
	}
}

// ========== 字符串处理函数 ==========

// Substr 返回字符串的子串
//
// 模板使用示例:
// {{ substr "Hello World" 0 5 }} <!-- 输出: "Hello" -->
// {{ substr "你好世界" 0 2 }} <!-- 输出: "你好" -->
func Substr(s string, start, length int) string {
	if start < 0 {
		start = 0
	}

	if length <= 0 {
		return ""
	}

	runes := []rune(s)

	if start >= len(runes) {
		return ""
	}

	end := start + length
	if end > len(runes) {
		end = len(runes)
	}

	return string(runes[start:end])
}

// Truncate 截断字符串并添加省略号
//
// 模板使用示例:
// {{ truncate "这是一段很长的文本，需要被截断" 9 }} <!-- 输出: "这是一段很长的..." -->
func Truncate(s string, length int) string {
	if length <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) <= length {
		return s
	}

	return string(runes[:length]) + "..."
}

// Nl2br 将换行符转换为HTML的<br>标签
//
// 模板使用示例:
// {{ nl2br "第一行\n第二行" }} <!-- 输出: "第一行<br>第二行" -->
func Nl2br(s string) template.HTML {
	return template.HTML(strings.Replace(
		template.HTMLEscapeString(s),
		"\n",
		"<br>",
		-1,
	))
}

// StripTags 移除HTML标签
//
// 模板使用示例:
// {{ stripTags "<p>这是<b>HTML</b>内容</p>" }} <!-- 输出: "这是HTML内容" -->
func StripTags(s string) string {
	initRegex()
	return htmlTagRegex.ReplaceAllString(s, "")
}

// ========== 数值处理函数 ==========

// Add 加法（优化版本，优先处理常见类型）
//
// 模板使用示例:
// {{ add 5 3 }} <!-- 输出: 8 -->
// {{ add 5.5 3.2 }} <!-- 输出: 8.7 -->
func Add(a, b any) any {
	// 优先处理最常见的类型，避免反射开销
	switch aVal := a.(type) {
	case int:
		switch bVal := b.(type) {
		case int:
			return aVal + bVal
		case float64:
			return float64(aVal) + bVal
		case int64:
			return int64(aVal) + bVal
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			return aVal + bVal
		case int:
			return aVal + float64(bVal)
		case int64:
			return aVal + float64(bVal)
		}
	case int64:
		switch bVal := b.(type) {
		case int64:
			return aVal + bVal
		case int:
			return aVal + int64(bVal)
		case float64:
			return float64(aVal) + bVal
		}
	}

	// 回退到反射方式处理其他类型
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int()
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int())
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float()
		}
	}

	return 0
}

// Subtract 减法（优化版本）
//
// 模板使用示例:
// {{ subtract 10 3 }} <!-- 输出: 7 -->
// {{ subtract 10.5 3.2 }} <!-- 输出: 7.3 -->
func Subtract(a, b any) any {
	// 优先处理常见类型
	switch aVal := a.(type) {
	case int:
		switch bVal := b.(type) {
		case int:
			return aVal - bVal
		case float64:
			return float64(aVal) - bVal
		case int64:
			return int64(aVal) - bVal
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			return aVal - bVal
		case int:
			return aVal - float64(bVal)
		case int64:
			return aVal - float64(bVal)
		}
	case int64:
		switch bVal := b.(type) {
		case int64:
			return aVal - bVal
		case int:
			return aVal - int64(bVal)
		case float64:
			return float64(aVal) - bVal
		}
	}

	// 回退到反射方式
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int()
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int())
		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float()
		}
	}

	return 0
}

// Multiply 乘法（优化版本）
//
// 模板使用示例:
// {{ multiply 5 3 }} <!-- 输出: 15 -->
// {{ multiply 5.5 3 }} <!-- 输出: 16.5 -->
func Multiply(a, b any) any {
	// 优先处理常见类型
	switch aVal := a.(type) {
	case int:
		switch bVal := b.(type) {
		case int:
			return aVal * bVal
		case float64:
			return float64(aVal) * bVal
		case int64:
			return int64(aVal) * bVal
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			return aVal * bVal
		case int:
			return aVal * float64(bVal)
		case int64:
			return aVal * float64(bVal)
		}
	case int64:
		switch bVal := b.(type) {
		case int64:
			return aVal * bVal
		case int:
			return aVal * int64(bVal)
		case float64:
			return float64(aVal) * bVal
		}
	}

	// 回退到反射方式
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int()
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int())
		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float()
		}
	}

	return 0
}

// Divide 除法（优化版本）
//
// 模板使用示例:
// {{ divide 10 2 }} <!-- 输出: 5 -->
// {{ divide 10 3 }} <!-- 输出: 3.3333333333333335 -->
// {{ divide 10 0 }} <!-- 输出: "除数不能为零" -->
func Divide(a, b any) any {
	// 优先处理常见类型
	switch aVal := a.(type) {
	case int:
		switch bVal := b.(type) {
		case int:
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / float64(bVal)
		case float64:
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / bVal
		case int64:
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / float64(bVal)
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			if bVal == 0 {
				return "除数不能为零"
			}
			return aVal / bVal
		case int:
			if bVal == 0 {
				return "除数不能为零"
			}
			return aVal / float64(bVal)
		case int64:
			if bVal == 0 {
				return "除数不能为零"
			}
			return aVal / float64(bVal)
		}
	case int64:
		switch bVal := b.(type) {
		case int64:
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / float64(bVal)
		case int:
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / float64(bVal)
		case float64:
			if bVal == 0 {
				return "除数不能为零"
			}
			return float64(aVal) / bVal
		}
	}

	// 回退到反射方式
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if bv.Int() == 0 {
				return "除数不能为零"
			}
			return float64(av.Int()) / float64(bv.Int())
		case reflect.Float32, reflect.Float64:
			if bv.Float() == 0 {
				return "除数不能为零"
			}
			return float64(av.Int()) / bv.Float()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if bv.Int() == 0 {
				return "除数不能为零"
			}
			return av.Float() / float64(bv.Int())
		case reflect.Float32, reflect.Float64:
			if bv.Float() == 0 {
				return "除数不能为零"
			}
			return av.Float() / bv.Float()
		}
	}

	return 0
}

// Mod 取模
//
// 模板使用示例:
// {{ mod 10 3 }} <!-- 输出: 1 -->
// {{ mod 10 2 }} <!-- 输出: 0 -->
func Mod(a, b any) any {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if bv.Int() == 0 {
				return 0
			}
			return av.Int() % bv.Int()
		}
	}

	return 0
}

// Round 四舍五入（优化版本）
//
// 模板使用示例:
// {{ round 3.1415926 2 }} <!-- 输出: 3.14 -->
// {{ round 3.1415926 4 }} <!-- 输出: 3.1416 -->
func Round(a any, precision int) float64 {
	var f float64

	// 优先处理常见类型
	switch v := a.(type) {
	case float64:
		f = v
	case float32:
		f = float64(v)
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	case int32:
		f = float64(v)
	default:
		f, _ = toFloat64(a)
	}

	p := math.Pow10(precision)
	return math.Round(f*p) / p
}

// ========== 日期时间处理函数 ==========

// Now 返回当前时间
//
// 模板使用示例:
// {{ now }} <!-- 输出: 当前时间对象 -->
func Now() time.Time {
	return time.Now()
}

// FormatDateTime 格式化时间
//
// 模板使用示例:
// {{ formatDateTime now }} <!-- 输出: "2023-05-20 14:30:00" -->
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
//
// 模板使用示例:
// {{ formatDate now }} <!-- 输出: "2023-05-20" -->
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// DateFormat 格式化日期时间
// 支持以下格式符号：
// Y - 四位数年份 (2006)
// y - 两位数年份 (06)
// m - 月份，有前导零 (01-12)
// n - 月份，无前导零 (1-12)
// d - 日期，有前导零 (01-31)
// j - 日期，无前导零 (1-31)
// H - 小时，24小时制，有前导零 (00-23)
// G - 小时，24小时制，无前导零 (0-23)
// h - 小时，12小时制，有前导零 (01-12)
// g - 小时，12小时制，无前导零 (1-12)
// i - 分钟，有前导零 (00-59)
// s - 秒数，有前导零 (00-59)
// A - 上午/下午 (AM/PM)
// a - 上午/下午 (am/pm)
// D - 星期几的缩写 (Mon-Sun)
// l - 星期几的全称 (Monday-Sunday)
// M - 月份的缩写 (Jan-Dec)
// F - 月份的全称 (January-December)
//
// 模板使用示例:
// {{ dateFormat now "Y-m-d" }} <!-- 输出: "2023-05-20" -->
// {{ dateFormat .UpdateTime "Y-m-d H:i:s" }} <!-- 输出: "2023-05-20 14:30:00" -->
// {{ dateFormat now "l, F j, Y" }} <!-- 输出: "Saturday, May 20, 2023" -->
func DateFormat(t time.Time, format string) string {
	patterns := map[string]string{
		// 年
		"Y": "2006", // 四位数年份
		"y": "06",   // 两位数年份
		// 月
		"m": "01",      // 有前导零 (01-12)
		"n": "1",       // 无前导零 (1-12)
		"M": "Jan",     // 月份的缩写 (Jan-Dec)
		"F": "January", // 月份的全称 (January-December)
		// 日
		"d": "02", // 有前导零 (01-31)
		"j": "2",  // 无前导零 (1-31)
		// 星期
		"D": "Mon",    // 星期几的缩写 (Mon-Sun)
		"l": "Monday", // 星期几的全称 (Monday-Sunday)
		// 时间
		"H": "15", // 小时，24小时制，有前导零 (00-23)
		"G": "15", // 小时，24小时制，无前导零 (0-23)
		"h": "03", // 小时，12小时制，有前导零 (01-12)
		"g": "3",  // 小时，12小时制，无前导零 (1-12)
		"i": "04", // 分钟，有前导零 (00-59)
		"s": "05", // 秒数，有前导零 (00-59)
		"A": "PM", // 上午/下午 (AM/PM)
		"a": "pm", // 上午/下午 (am/pm)
	}

	layout := format
	for p, l := range patterns {
		layout = strings.ReplaceAll(layout, p, l)
	}

	return t.Format(layout)
}

// HumanizeTime 人性化时间显示
//
// 模板使用示例:
// {{ humanizeTime .CreateTime }} <!-- 根据与当前时间的差距输出，如 "3小时前"、"昨天"、"2个月前" -->
func HumanizeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	} else if diff < 48*time.Hour {
		return "昨天"
	} else if diff < 72*time.Hour {
		return "前天"
	} else if diff < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	} else if diff < 365*24*time.Hour {
		return fmt.Sprintf("%d个月前", int(diff.Hours()/(24*30)))
	}

	return fmt.Sprintf("%d年前", int(diff.Hours()/(24*365)))
}

// ========== 集合处理函数 ==========

// First 返回切片的第一个元素
//
// 模板使用示例:
// {{ first .Items }} <!-- 输出: 切片的第一个元素 -->
func First(a any) any {
	v := reflect.ValueOf(a)

	if v.Kind() == reflect.Slice && v.Len() > 0 {
		return v.Index(0).Interface()
	}

	return nil
}

// Last 返回切片的最后一个元素
//
// 模板使用示例:
// {{ last .Items }} <!-- 输出: 切片的最后一个元素 -->
func Last(a any) any {
	v := reflect.ValueOf(a)

	if v.Kind() == reflect.Slice && v.Len() > 0 {
		return v.Index(v.Len() - 1).Interface()
	}

	return nil
}

// Empty 检查是否为空（优化版本）
//
// 模板使用示例:
// {{ if empty .Items }}暂无数据{{ end }}
// {{ if empty "" }}字符串为空{{ end }}
// {{ if empty 0 }}值为零{{ end }}
func Empty(a any) bool {
	if a == nil {
		return true
	}

	// 优先处理常见类型
	switch v := a.(type) {
	case string:
		return v == ""
	case int:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return v == 0
	case bool:
		return !v
	case []string:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	case map[string]string:
		return len(v) == 0
	}

	// 回退到反射方式
	rv := reflect.ValueOf(a)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	}

	return false
}

// NotEmpty 检查是否非空
//
// 模板使用示例:
// {{ if notEmpty .Items }}有数据{{ else }}暂无数据{{ end }}
func NotEmpty(a any) bool {
	return !Empty(a)
}

// Length 返回长度（优化版本，正确处理中文）
//
// 模板使用示例:
// {{ length .Items }} <!-- 输出: 切片的长度 -->
// {{ length "Hello" }} <!-- 输出: 5 -->
// {{ length "你好" }} <!-- 输出: 2 -->
func Length(a any) int {
	// 优先处理常见类型
	switch v := a.(type) {
	case string:
		return len([]rune(v)) // 正确处理中文字符
	case []string:
		return len(v)
	case []int:
		return len(v)
	case []any:
		return len(v)
	case map[string]any:
		return len(v)
	case map[string]string:
		return len(v)
	case map[string]int:
		return len(v)
	}

	// 回退到反射方式
	rv := reflect.ValueOf(a)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len()
	case reflect.String:
		return len([]rune(rv.String())) // 正确处理中文
	}

	return 0
}

// InArray 检查元素是否在数组中（优化版本）
//
// 模板使用示例:
// {{ if inArray "admin" .Roles }}是管理员{{ end }}
// {{ if inArray 5 .AllowedIds }}ID有效{{ end }}
func InArray(needle any, haystack any) bool {
	// 优先处理常见类型
	switch arr := haystack.(type) {
	case []string:
		if needleStr, ok := needle.(string); ok {
			for _, item := range arr {
				if item == needleStr {
					return true
				}
			}
			return false
		}
	case []int:
		if needleInt, ok := needle.(int); ok {
			for _, item := range arr {
				if item == needleInt {
					return true
				}
			}
			return false
		}
	case []int64:
		if needleInt64, ok := needle.(int64); ok {
			for _, item := range arr {
				if item == needleInt64 {
					return true
				}
			}
			return false
		}
	case []any:
		for _, item := range arr {
			if reflect.DeepEqual(needle, item) {
				return true
			}
		}
		return false
	}

	// 回退到反射方式
	v := reflect.ValueOf(haystack)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(needle, v.Index(i).Interface()) {
			return true
		}
	}

	return false
}

// ========== 条件处理函数 ==========

// Default 如果值为空则返回默认值
//
// 模板使用示例:
// {{ default .Name "匿名用户" }} <!-- 如果 .Name 为空，输出 "匿名用户"，否则输出 .Name -->
func Default(value, defaultValue any) any {
	if Empty(value) {
		return defaultValue
	}
	return value
}

// Ternary 三元运算符
//
// 模板使用示例:
// {{ ternary (eq .Status 1) "启用" "禁用" }} <!-- 如果 .Status 等于 1，输出 "启用"，否则输出 "禁用" -->
func Ternary(condition bool, trueValue, falseValue any) any {
	if condition {
		return trueValue
	}
	return falseValue
}

// Eq 相等比较
//
// 模板使用示例:
// {{ if eq .Status 1 }}已启用{{ end }}
func Eq(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// Ne 不等比较
//
// 模板使用示例:
// {{ if ne .Status 0 }}非禁用状态{{ end }}
func Ne(a, b any) bool {
	return !reflect.DeepEqual(a, b)
}

// Lt 小于比较
//
// 模板使用示例:
// {{ if lt .Count 5 }}数量小于5{{ end }}
func Lt(a, b any) bool {
	return compare(a, b) < 0
}

// Lte 小于等于比较
//
// 模板使用示例:
// {{ if lte .Count 5 }}数量不超过5{{ end }}
func Lte(a, b any) bool {
	return compare(a, b) <= 0
}

// Gt 大于比较
//
// 模板使用示例:
// {{ if gt .Count 10 }}数量大于10{{ end }}
func Gt(a, b any) bool {
	return compare(a, b) > 0
}

// Gte 大于等于比较
//
// 模板使用示例:
// {{ if gte .Count 10 }}数量至少为10{{ end }}
func Gte(a, b any) bool {
	return compare(a, b) >= 0
}

// ========== 安全处理函数 ==========

// SafeHTML 安全HTML
//
// 模板使用示例:
// {{ safeHTML "<strong>加粗文本</strong>" }} <!-- 输出未转义的HTML: <strong>加粗文本</strong> -->
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// SafeJS 安全JavaScript
//
// 模板使用示例:
// {{ safeJS "alert('Hello');" }} <!-- 输出未转义的JavaScript代码 -->
func SafeJS(s string) template.JS {
	return template.JS(s)
}

// SafeCSS 安全CSS
//
// 模板使用示例:
// {{ safeCSS "body { color: red; }" }} <!-- 输出未转义的CSS代码 -->
func SafeCSS(s string) template.CSS {
	return template.CSS(s)
}

// SafeURL 安全URL
//
// 模板使用示例:
// {{ safeURL "https://example.com?param=value" }} <!-- 输出未转义的URL -->
func SafeURL(s string) template.URL {
	return template.URL(s)
}

// ========== 辅助函数 ==========

// toFloat64 将任意数值类型转换为float64
func toFloat64(v any) (float64, error) {
	if v == nil {
		return 0, fmt.Errorf("值为空")
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.String:
		var f float64
		if _, err := fmt.Sscanf(rv.String(), "%f", &f); err != nil {
			return 0, err
		}
		return f, nil
	}

	return 0, fmt.Errorf("无法转换为浮点数")
}

// compare 比较两个值
func compare(a, b any) int {
	if a == nil && b == nil {
		return 0
	}

	if a == nil {
		return -1
	}

	if b == nil {
		return 1
	}

	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	// 如果类型相同则直接比较
	if av.Kind() == bv.Kind() {
		switch av.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Int() < bv.Int() {
				return -1
			} else if av.Int() > bv.Int() {
				return 1
			}
			return 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Uint() < bv.Uint() {
				return -1
			} else if av.Uint() > bv.Uint() {
				return 1
			}
			return 0
		case reflect.Float32, reflect.Float64:
			if av.Float() < bv.Float() {
				return -1
			} else if av.Float() > bv.Float() {
				return 1
			}
			return 0
		case reflect.String:
			if av.String() < bv.String() {
				return -1
			} else if av.String() > bv.String() {
				return 1
			}
			return 0
		}
	}

	// 尝试转换为浮点数比较
	af, aErr := toFloat64(a)
	bf, bErr := toFloat64(b)

	if aErr == nil && bErr == nil {
		if af < bf {
			return -1
		} else if af > bf {
			return 1
		}
		return 0
	}

	// 无法比较，按字符串比较
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// ========== 路由URL生成函数 ==========

// Route 根据路由名称生成URL
//
// 模板使用示例:
// <a href="{{ url "user@show" }}">用户页面</a>
// <a href="{{ url "user@detail" (map "id" 123) }}">用户详情</a>
func Route(name string, params ...map[string]any) template.URL {
	url, err := router.BuildUrl(name, params...)
	if err != nil {
		panic(err) // 如果路由不存在，则触发 panic
	}
	return template.URL(url)
}

// ========== Map处理函数 ==========

// MapGet 从map中获取指定键的值
//
// 模板使用示例:
// {{ mapGet .Data "username" }} <!-- 输出: .Data 中 "username" 键对应的值 -->
func MapGet(m any, key any) any {
	v := reflect.ValueOf(m)

	if v.Kind() == reflect.Map {
		keyValue := reflect.ValueOf(key)
		if !keyValue.Type().AssignableTo(v.Type().Key()) {
			// 键类型不匹配
			return nil
		}

		value := v.MapIndex(keyValue)
		if value.IsValid() {
			return value.Interface()
		}
	}

	return nil
}

// MapHas 检查map是否包含指定的键
//
// 模板使用示例:
// {{ if mapHas .Data "error" }}存在错误信息{{ end }}
func MapHas(m any, key any) bool {
	v := reflect.ValueOf(m)

	if v.Kind() == reflect.Map {
		keyValue := reflect.ValueOf(key)
		if !keyValue.Type().AssignableTo(v.Type().Key()) {
			// 键类型不匹配
			return false
		}

		value := v.MapIndex(keyValue)
		return value.IsValid()
	}

	return false
}

// MapKeys 获取map的所有键
//
// 模板使用示例:
// <ul>
//
//	{{ range mapKeys .Data }}
//	  <li>{{ . }}: {{ mapGet $.Data . }}</li>
//	{{ end }}
//
// </ul>
func MapKeys(m any) []any {
	v := reflect.ValueOf(m)

	if v.Kind() != reflect.Map {
		return nil
	}

	keys := v.MapKeys()
	result := make([]any, len(keys))

	for i, key := range keys {
		result[i] = key.Interface()
	}

	return result
}

// MapSet 创建一个新的map并设置键值对
//
// 模板使用示例:
// {{ $newMap := mapSet nil "name" "张三" }}
// {{ $newMap = mapSet $newMap "age" 25 }}
// {{ mapGet $newMap "name" }} <!-- 输出: "张三" -->
func MapSet(m any, key any, value any) map[any]any {
	var result map[any]any

	if m == nil {
		// 创建新map
		result = make(map[any]any)
	} else {
		v := reflect.ValueOf(m)
		if v.Kind() != reflect.Map {
			// 如果不是map，创建新map
			result = make(map[any]any)
		} else {
			// 复制现有map
			result = make(map[any]any)
			iter := v.MapRange()
			for iter.Next() {
				k := iter.Key().Interface()
				v := iter.Value().Interface()
				result[k] = v
			}
		}
	}

	// 设置新的键值对
	result[key] = value
	return result
}

// NewMap 创建一个字典/映射
//
// 模板使用示例:
// {{ $data := map "name" "张三" "age" 25 "email" "zhangsan@example.com" }}
// {{ mapGet $data "name" }} <!-- 输出: "张三" -->
func NewMap(values ...any) map[string]any {
	if len(values)%2 != 0 {
		// 如果参数个数不是偶数，返回空映射
		return map[string]any{}
	}

	dict := make(map[string]any, len(values)/2)

	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			// 如果键不是字符串类型，跳过该键值对
			continue
		}
		dict[key] = values[i+1]
	}

	return dict
}

// ========== 错误处理函数 ==========

// Panic 在模板中触发panic，用于报告致命错误
//
// 模板使用示例:
// {{ if empty .User }}
//   {{ panic "用户信息不能为空" }}
// {{ end }}
//
// {{ if not (mapHas .Config "database") }}
//   {{ panic "缺少数据库配置" }}
// {{ end }}
func Panic(message string) string {
	panic(message)
}

// ========== 调试函数 ==========

// Dump 调试打印变量内容，支持数组、切片、结构体、指针等类型
//
// 模板使用示例:
// {{ dump .User }}
// {{ dump .Items }}
// {{ dump .Config }}
func Dump(v any) template.HTML {
	if v == nil {
		return template.HTML("<pre>nil</pre>")
	}

	output := dumpValue(reflect.ValueOf(v), 0)
	return template.HTML("<pre>" + template.HTMLEscapeString(output) + "</pre>")
}

// dumpValue 递归打印值的详细内容
func dumpValue(v reflect.Value, indent int) string {
	if !v.IsValid() {
		return "invalid"
	}

	// 处理指针类型
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "nil"
		}
		return "*" + dumpValue(v.Elem(), indent)
	}

	// 处理接口类型
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "nil"
		}
		return dumpValue(v.Elem(), indent)
	}

	indentStr := strings.Repeat("  ", indent)
	nextIndentStr := strings.Repeat("  ", indent+1)

	switch v.Kind() {
	case reflect.Struct:
		// 先尝试使用 JSON 序列化（更可读）
		if v.CanInterface() {
			if jsonBytes, err := json.MarshalIndent(v.Interface(), indentStr, "  "); err == nil {
				return string(jsonBytes)
			}
		}

		// 回退到字段打印
		var result strings.Builder
		result.WriteString(v.Type().String() + " {\n")

		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fieldValue := v.Field(i)

			// 跳过未导出的字段
			if !field.IsExported() {
				continue
			}

			result.WriteString(nextIndentStr)
			result.WriteString(field.Name)
			result.WriteString(": ")

			if fieldValue.CanInterface() {
				result.WriteString(dumpValue(fieldValue, indent+1))
			} else {
				result.WriteString("<unexported>")
			}

			result.WriteString("\n")
		}

		result.WriteString(indentStr + "}")
		return result.String()

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return "[]"
		}

		// 先尝试使用 JSON 序列化
		if v.CanInterface() {
			if jsonBytes, err := json.MarshalIndent(v.Interface(), indentStr, "  "); err == nil {
				return string(jsonBytes)
			}
		}

		var result strings.Builder
		result.WriteString("[\n")

		for i := 0; i < v.Len(); i++ {
			result.WriteString(nextIndentStr)
			result.WriteString(fmt.Sprintf("[%d]: ", i))
			result.WriteString(dumpValue(v.Index(i), indent+1))
			result.WriteString("\n")
		}

		result.WriteString(indentStr + "]")
		return result.String()

	case reflect.Map:
		if v.Len() == 0 {
			return "{}"
		}

		// 先尝试使用 JSON 序列化
		if v.CanInterface() {
			if jsonBytes, err := json.MarshalIndent(v.Interface(), indentStr, "  "); err == nil {
				return string(jsonBytes)
			}
		}

		var result strings.Builder
		result.WriteString("{\n")

		iter := v.MapRange()
		for iter.Next() {
			result.WriteString(nextIndentStr)
			result.WriteString(fmt.Sprintf("%v: ", iter.Key().Interface()))
			result.WriteString(dumpValue(iter.Value(), indent+1))
			result.WriteString("\n")
		}

		result.WriteString(indentStr + "}")
		return result.String()

	case reflect.String:
		return fmt.Sprintf("%q", v.String())

	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())

	default:
		if v.CanInterface() {
			return fmt.Sprintf("%v", v.Interface())
		}
		return fmt.Sprintf("<%s>", v.Kind())
	}
}
