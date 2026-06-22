package request

import (
	"mime/multipart"
	"net"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// InputType 定义 Input 函数支持的类型约束。
// 使用精确类型而非 ~ 近似约束：函数内部按精确类型分发，
// 若放开为派生类型会在运行时静默匹配失败，故在编译期就限定为这些精确类型。
// 文件上传请使用 File / Files 函数。
type InputType interface {
	string | int | int64 | float32 | float64 | bool |
		[]string | []int | []int64
}

// IsAjax 判断是否为 AJAX 请求
func IsAjax(c *gin.Context) bool {
	// 检查 X-Requested-With 头
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		return true
	}

	// 检查 Content-Type
	contentType := c.ContentType()
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// 检查 Accept 头
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
		return true
	}

	return false
}

// IsJSON 判断是否为 JSON 请求
func IsJSON(c *gin.Context) bool {
	contentType := c.ContentType()
	return strings.Contains(contentType, "application/json")
}

// IsXML 判断是否为 XML 请求
func IsXML(c *gin.Context) bool {
	contentType := c.ContentType()
	return strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml")
}

// IsFormData 判断是否为表单数据请求
func IsFormData(c *gin.Context) bool {
	contentType := c.ContentType()
	return strings.Contains(contentType, "application/x-www-form-urlencoded")
}

// IsMultipartForm 判断是否为 multipart/form-data 请求
func IsMultipartForm(c *gin.Context) bool {
	contentType := c.ContentType()
	return strings.Contains(contentType, "multipart/form-data")
}

// IsMobile 判断是否为移动设备访问
func IsMobile(c *gin.Context) bool {
	userAgent := strings.ToLower(c.GetHeader("User-Agent"))
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipad", "ipod",
		"blackberry", "windows phone", "webos",
	}

	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}

	return false
}

// GetClientIP 获取客户端真实 IP 地址。
// 委托给 gin.Context.ClientIP()，由其依据可信代理（TrustedProxies）配置安全地解析
// X-Forwarded-For/X-Real-IP；不再直接信任可被伪造的转发头。
func GetClientIP(c *gin.Context) string {
	return c.ClientIP()
}

// GetUserAgent 获取 User-Agent
func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// GetReferer 获取 Referer
func GetReferer(c *gin.Context) string {
	return c.GetHeader("Referer")
}

// IsSecure 判断是否为 HTTPS 请求
func IsSecure(c *gin.Context) bool {
	// 检查协议
	if c.Request.TLS != nil {
		return true
	}

	// 检查 X-Forwarded-Proto 头（代理场景）
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "https" {
		return true
	}

	return false
}

// GetScheme 获取请求协议（http/https）
func GetScheme(c *gin.Context) string {
	if IsSecure(c) {
		return "https"
	}
	return "http"
}

// GetHost 获取主机名
func GetHost(c *gin.Context) string {
	// 优先从 X-Forwarded-Host 获取
	host := c.GetHeader("X-Forwarded-Host")
	if host != "" {
		return host
	}

	// 使用 Request.Host
	return c.Request.Host
}

// GetFullURL 获取完整的请求 URL
func GetFullURL(c *gin.Context) string {
	scheme := GetScheme(c)
	host := GetHost(c)
	return scheme + "://" + host + c.Request.RequestURI
}

// IsLocalIP 判断是否为本地 IP
func IsLocalIP(ip string) bool {
	// 解析 IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查是否为环回地址
	if parsedIP.IsLoopback() {
		return true
	}

	// 检查是否为私有地址
	if parsedIP.IsPrivate() {
		return true
	}

	return false
}

// GetMethod 获取请求方法
func GetMethod(c *gin.Context) string {
	return c.Request.Method
}

// IsMethod 判断是否为指定的请求方法
func IsMethod(c *gin.Context, method string) bool {
	return strings.EqualFold(c.Request.Method, method)
}

// IsGET 判断是否为 GET 请求
func IsGET(c *gin.Context) bool {
	return IsMethod(c, "GET")
}

// IsPOST 判断是否为 POST 请求
func IsPOST(c *gin.Context) bool {
	return IsMethod(c, "POST")
}

// IsPUT 判断是否为 PUT 请求
func IsPUT(c *gin.Context) bool {
	return IsMethod(c, "PUT")
}

// IsDELETE 判断是否为 DELETE 请求
func IsDELETE(c *gin.Context) bool {
	return IsMethod(c, "DELETE")
}

// IsPATCH 判断是否为 PATCH 请求
func IsPATCH(c *gin.Context) bool {
	return IsMethod(c, "PATCH")
}

// GetBearerToken 从 Authorization 头获取 Bearer Token
func GetBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	// Bearer token 格式：Bearer <token>
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	return ""
}

// AcceptsJSON 判断客户端是否接受 JSON 响应
func AcceptsJSON(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "application/json") || strings.Contains(accept, "*/*")
}

// AcceptsHTML 判断客户端是否接受 HTML 响应
func AcceptsHTML(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}

// getRawValue 获取原始字符串值（内部辅助函数）
// 优先级：POST Form > Query > URL Params
func getRawValue(c *gin.Context, key string) string {
	if v := c.PostForm(key); v != "" {
		return v
	}
	if v := c.Query(key); v != "" {
		return v
	}
	if v := c.Param(key); v != "" {
		return v
	}
	return ""
}

// getArrayValues 获取数组形式的字符串值（内部辅助函数）
// 支持: ?key[]=a&key[]=b、?key=a&key=b、?key=a,b，以及两者混用（?key=a,b&key=c）
func getArrayValues(c *gin.Context, key string) []string {
	var raw []string
	switch {
	case len(c.QueryArray(key)) > 0:
		raw = c.QueryArray(key)
	case len(c.QueryArray(key+"[]")) > 0:
		raw = c.QueryArray(key + "[]")
	case len(c.PostFormArray(key)) > 0:
		raw = c.PostFormArray(key)
	case len(c.PostFormArray(key+"[]")) > 0:
		raw = c.PostFormArray(key + "[]")
	default:
		return nil
	}

	// 对取到的每个值再按逗号展开，使 ?key=a,b 与 ?key=a&key=b 两种形式都生效
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		for part := range strings.SplitSeq(item, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Input 按 key 读取请求参数并转换为目标类型 T，缺失或解析失败时返回默认值。
//
// 取值优先级：POST 表单 > Query > URL 路径参数（见 getRawValue / getArrayValues）。
// 注意：不读取 JSON 请求体；JSON 请求请使用 BindJSON 绑定到结构体。
//
// 支持的类型：
//   - 基本类型: string, int, int64, float32, float64, bool
//   - 数组类型: []string, []int, []int64
//
// 文件上传请使用 File / Files。
//
// 使用示例：
//
//	name := request.Input(c, "name", "默认名称")   // string
//	age := request.Input(c, "age", 18)            // int
//	price := request.Input(c, "price", 9.99)      // float64
//	active := request.Input(c, "active", true)    // bool
//	tags := request.Input(c, "tags", []string{})  // []string
//	ids := request.Input(c, "ids", []int{})       // []int
func Input[T InputType](c *gin.Context, key string, defaultValue ...T) T {
	var def T
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	switch any(def).(type) {
	case []string:
		if v := getArrayValues(c, key); v != nil {
			return any(v).(T)
		}

	case []int:
		if items := getArrayValues(c, key); items != nil {
			vals := make([]int, 0, len(items))
			for _, s := range items {
				if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
					vals = append(vals, n)
				}
			}
			if len(vals) > 0 {
				return any(vals).(T)
			}
		}

	case []int64:
		if items := getArrayValues(c, key); items != nil {
			vals := make([]int64, 0, len(items))
			for _, s := range items {
				if n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
					vals = append(vals, n)
				}
			}
			if len(vals) > 0 {
				return any(vals).(T)
			}
		}

	case string:
		if v := getRawValue(c, key); v != "" {
			return any(v).(T)
		}

	case int:
		if v := getRawValue(c, key); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				return any(n).(T)
			}
		}

	case int64:
		if v := getRawValue(c, key); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				return any(n).(T)
			}
		}

	case float32:
		if v := getRawValue(c, key); v != "" {
			if n, err := strconv.ParseFloat(v, 32); err == nil {
				return any(float32(n)).(T)
			}
		}

	case float64:
		if v := getRawValue(c, key); v != "" {
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				return any(n).(T)
			}
		}

	case bool:
		if v := getRawValue(c, key); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				return any(b).(T)
			}
		}
	}

	return def
}

// File 获取单个上传文件，不存在时返回 nil
func File(c *gin.Context, key string) *multipart.FileHeader {
	if file, err := c.FormFile(key); err == nil {
		return file
	}
	return nil
}

// Files 获取同名的多个上传文件，不存在时返回 nil
func Files(c *gin.Context, key string) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err != nil {
		return nil
	}
	return form.File[key]
}
