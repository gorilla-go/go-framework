package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// RouteBuilder 路由构建器
type RouteBuilder struct {
	router *gin.Engine
	group  *gin.RouterGroup
}

// RouterAnnotation 路由注册接口
// 所有需要注册路由的控制器都应实现此接口
type RouterAnnotation interface {
	// Annotation 注册路由
	// rb 是路由构建器
	// 每个控制器负责注册自己的路由
	Annotation(rb *RouteBuilder)
}

// Route 路由信息
type Route struct {
	Name    string
	Path    string
	Method  string
	Handler gin.HandlerFunc
}

// 全局路由注册表
var (
	routes      = make(map[string]*Route)
	routesMutex sync.RWMutex
)

// NewRouteBuilder 创建路由构建器
func NewRouteBuilder(router *gin.Engine) *RouteBuilder {
	return &RouteBuilder{
		router: router,
	}
}

// Group 创建路由组
func (rb *RouteBuilder) Group(path string) *RouteBuilder {
	return &RouteBuilder{
		router: rb.router,
		group:  rb.router.Group(path),
	}
}

// GET 注册GET请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) GET(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("GET", path, name, handler)
}

// POST 注册POST请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) POST(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("POST", path, name, handler)
}

// PUT 注册PUT请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) PUT(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("PUT", path, name, handler)
}

// DELETE 注册DELETE请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) DELETE(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("DELETE", path, name, handler)
}

// 注册路由，内部函数
func (rb *RouteBuilder) registerRoute(method, path, name string, handler gin.HandlerFunc) {
	// 如果没有提供名称，使用默认命名规则
	if name == "" {
		name = fmt.Sprintf("%s:%s", method, path)
	}

	// 注册到gin
	switch method {
	case "GET":
		if rb.group != nil {
			rb.group.GET(path, handler)
		} else {
			rb.router.GET(path, handler)
		}
	case "POST":
		if rb.group != nil {
			rb.group.POST(path, handler)
		} else {
			rb.router.POST(path, handler)
		}
	case "PUT":
		if rb.group != nil {
			rb.group.PUT(path, handler)
		} else {
			rb.router.PUT(path, handler)
		}
	case "DELETE":
		if rb.group != nil {
			rb.group.DELETE(path, handler)
		} else {
			rb.router.DELETE(path, handler)
		}
	}

	// 记录路由信息
	routesMutex.Lock()
	defer routesMutex.Unlock()

	var fullPath string
	if rb.group != nil {
		// 获取路由组的路径前缀
		groupPrefix := ""
		if rb.group != nil && len(rb.group.Handlers) > 0 {
			// Gin的路由组没有直接暴露前缀，这里需要一个变通的方法
			// 实际项目中，你可能需要单独记录路由组前缀
			groupPrefix = "/"
		}
		fullPath = groupPrefix + path
	} else {
		fullPath = path
	}

	routes[name] = &Route{
		Name:    name,
		Path:    fullPath,
		Method:  method,
		Handler: handler,
	}
}

// PATCH 注册PATCH请求路由
func (rb *RouteBuilder) PATCH(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("PATCH", path, name, handler)
}

// HEAD 注册HEAD请求路由
func (rb *RouteBuilder) HEAD(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("HEAD", path, name, handler)
}

// OPTIONS 注册OPTIONS请求路由
func (rb *RouteBuilder) OPTIONS(path string, handler gin.HandlerFunc, name string) {
	rb.registerRoute("OPTIONS", path, name, handler)
}

// ANY 注册所有HTTP方法路由
func (rb *RouteBuilder) ANY(path string, handler gin.HandlerFunc, name string) {
	// 注册到gin
	if rb.group != nil {
		rb.group.Any(path, handler)
	} else {
		rb.router.Any(path, handler)
	}

	// 记录路由信息
	routesMutex.Lock()
	defer routesMutex.Unlock()

	var fullPath string
	if rb.group != nil {
		// 获取路由组的路径前缀
		groupPrefix := ""
		if rb.group != nil && len(rb.group.Handlers) > 0 {
			groupPrefix = "/"
		}
		fullPath = groupPrefix + path
	} else {
		fullPath = path
	}

	routes[name] = &Route{
		Name:    name,
		Path:    fullPath,
		Method:  "ANY",
		Handler: handler,
	}
}

// ParseParam 将参数解析为指定类型
func ParseParam(c *gin.Context, name string, defaultVal interface{}) interface{} {
	val := c.Param(name)
	if val == "" {
		return defaultVal
	}

	switch defaultVal.(type) {
	case int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return defaultVal
		}
		return intVal
	case int64:
		int64Val, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return defaultVal
		}
		return int64Val
	case float64:
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return defaultVal
		}
		return floatVal
	case bool:
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			return defaultVal
		}
		return boolVal
	default:
		return val
	}
}

// BuildUrl 根据路由名称和参数生成URL
func BuildUrl(name string, params ...map[string]interface{}) (string, error) {
	routesMutex.RLock()
	route, exists := routes[name]
	routesMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("路由不存在: %s", name) // 路由不存在时返回错误
	}

	path := route.Path

	// 替换参数
	if len(params) > 0 {
		for key, value := range params[0] {
			paramPlaceholder := ":" + key
			// 将参数值转换为字符串
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				strValue = fmt.Sprintf("%d", v)
			case float32, float64:
				strValue = fmt.Sprintf("%g", v)
			case bool:
				strValue = fmt.Sprintf("%t", v)
			default:
				strValue = fmt.Sprintf("%v", v)
			}
			path = strings.Replace(path, paramPlaceholder, strValue, -1)
		}
	}

	return path, nil
}
