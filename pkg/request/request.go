package request

import (
	"mime/multipart"
	"net"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// InputType 定义 Input 函数支持的类型约束
type InputType interface {
	~string | ~int | ~int64 | ~float32 | ~float64 | ~bool |
		~[]string | ~[]int | ~[]int64 |
		*multipart.FileHeader | []*multipart.FileHeader
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

// GetClientIP 获取客户端真实 IP 地址
func GetClientIP(c *gin.Context) string {
	// 优先从 X-Real-IP 获取
	clientIP := c.GetHeader("X-Real-IP")
	if clientIP != "" {
		return clientIP
	}

	// 从 X-Forwarded-For 获取（取第一个）
	clientIP = c.GetHeader("X-Forwarded-For")
	if clientIP != "" {
		ips := strings.Split(clientIP, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 使用 Gin 的 ClientIP 方法
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
// 支持: ?key[]=a&key[]=b, ?key=a&key=b, ?key=a,b 等格式
func getArrayValues(c *gin.Context, key string) []string {
	if v := c.QueryArray(key); len(v) > 0 {
		return v
	}
	if v := c.QueryArray(key + "[]"); len(v) > 0 {
		return v
	}
	if v := c.PostFormArray(key); len(v) > 0 {
		return v
	}
	if v := c.PostFormArray(key + "[]"); len(v) > 0 {
		return v
	}
	// 兜底：单值按逗号分割（支持 ?key=a,b,c 或 form key=a,b,c）
	if v := c.Query(key); v != "" {
		return strings.Split(v, ",")
	}
	if v := c.PostForm(key); v != "" {
		return strings.Split(v, ",")
	}
	return nil
}

// 优先级：POST > GET > URL Params > JSON Body
//
// 支持的类型：
//   - 基本类型: string, int, int64, float32, float64, bool
//   - 数组类型: []string, []int, []int64
//   - 文件类型: *multipart.FileHeader, []*multipart.FileHeader
//
// 使用示例：
//
//	name := request.Input(c, "name", "默认名称")                              // string
//	age := request.Input(c, "age", 18)                                      // int
//	price := request.Input(c, "price", 9.99)                                // float64
//	active := request.Input(c, "active", true)                              // bool
//	tags := request.Input(c, "tags", []string{})                            // []string
//	ids := request.Input(c, "ids", []int{})                                 // []int
//	file := request.Input[*multipart.FileHeader](c, "avatar")               // 单个文件
//	files := request.Input[*multipart.FileHeader](c, "images")              // 多个文件
func Input[T InputType](c *gin.Context, key string, defaultValue ...T) T {
	var def T
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	switch any(def).(type) {
	case *multipart.FileHeader:
		if file, err := c.FormFile(key); err == nil {
			return any(file).(T)
		}

	case []*multipart.FileHeader:
		if form, err := c.MultipartForm(); err == nil {
			if files, ok := form.File[key]; ok && len(files) > 0 {
				return any(files).(T)
			}
		}

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
