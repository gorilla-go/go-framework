package request

import (
	"fmt"
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
// 优先级：POST > GET > URL Params > JSON Body
func getRawValue(c *gin.Context, key string) string {
	var value string

	// 1. 尝试从 POST Form 获取
	if value = c.PostForm(key); value != "" {
		return value
	}

	// 2. 尝试从 Query 获取
	if value = c.Query(key); value != "" {
		return value
	}

	// 3. 尝试从 URL Params 获取
	if value = c.Param(key); value != "" {
		return value
	}

	// 4. 尝试从 JSON Body 获取（如果是 JSON 请求）
	if IsJSON(c) {
		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err == nil {
			if val, ok := jsonData[key]; ok {
				// 将任意类型转换为字符串
				switch v := val.(type) {
				case string:
					return v
				case float64:
					return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", v), "0"), ".")
				case bool:
					return fmt.Sprintf("%t", v)
				default:
					return fmt.Sprintf("%v", v)
				}
			}
		}
	}

	return ""
}

// getArrayValues 获取数组形式的字符串值（内部辅助函数）
// 支持: ?key[]=a&key[]=b, ?key=a&key=b, ?key=a,b 等格式
func getArrayValues(c *gin.Context, key string) []string {
	// 1. 尝试从 Query 获取数组
	if values := c.QueryArray(key); len(values) > 0 {
		return values
	}
	if values := c.QueryArray(key + "[]"); len(values) > 0 {
		return values
	}

	// 2. 尝试从 POST Form 获取数组
	if values := c.PostFormArray(key); len(values) > 0 {
		return values
	}
	if values := c.PostFormArray(key + "[]"); len(values) > 0 {
		return values
	}

	// 3. 尝试获取单个值，按逗号分隔
	if value := getRawValue(c, key); value != "" {
		return strings.Split(value, ",")
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
	// 获取默认值（用于类型推断和错误时返回）
	var def T
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	// 处理文件类型
	switch any(def).(type) {
	case *multipart.FileHeader:
		file, err := c.FormFile(key)
		if err == nil {
			return any(file).(T)
		}
		return def

	case []*multipart.FileHeader:
		form, err := c.MultipartForm()
		if err == nil {
			if files, ok := form.File[key]; ok && len(files) > 0 {
				return any(files).(T)
			}
		}
		return def
	}

	// 处理数组类型
	switch any(def).(type) {
	case []string:
		if values := getArrayValues(c, key); values != nil {
			return any(values).(T)
		}
		return def

	case []int:
		if strValues := getArrayValues(c, key); strValues != nil {
			intValues := make([]int, 0, len(strValues))
			for _, v := range strValues {
				if intVal, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
					intValues = append(intValues, intVal)
				}
			}
			if len(intValues) > 0 {
				return any(intValues).(T)
			}
		}
		return def

	case []int64:
		if strValues := getArrayValues(c, key); strValues != nil {
			int64Values := make([]int64, 0, len(strValues))
			for _, v := range strValues {
				if int64Val, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
					int64Values = append(int64Values, int64Val)
				}
			}
			if len(int64Values) > 0 {
				return any(int64Values).(T)
			}
		}
		return def
	}

	// 处理单值类型
	value := getRawValue(c, key)
	if value == "" {
		return def
	}

	// 根据类型进行转换
	switch any(def).(type) {
	case string:
		return any(value).(T)
	case int:
		if v, err := strconv.Atoi(value); err == nil {
			return any(v).(T)
		}
	case int64:
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return any(v).(T)
		}
	case float32:
		if v, err := strconv.ParseFloat(value, 32); err == nil {
			return any(float32(v)).(T)
		}
	case float64:
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return any(v).(T)
		}
	case bool:
		if v, err := strconv.ParseBool(value); err == nil {
			return any(v).(T)
		}
	}

	return def
}
