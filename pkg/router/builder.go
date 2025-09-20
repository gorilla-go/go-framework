package router

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// RouteBuilder 路由构建器
type RouteBuilder struct {
	router *gin.Engine
	group  *gin.RouterGroup
}

type RouterAnnotation interface {
	// Annotation 注册路由
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
	if name == "" {
		name = fmt.Sprintf("%s:%s", method, path)
	}

	// HTTP方法到Gin方法的映射
	methodHandlers := map[string]func(string, ...gin.HandlerFunc) gin.IRoutes{
		"GET":     rb.router.GET,
		"POST":    rb.router.POST,
		"PUT":     rb.router.PUT,
		"DELETE":  rb.router.DELETE,
		"PATCH":   rb.router.PATCH,
		"HEAD":    rb.router.HEAD,
		"OPTIONS": rb.router.OPTIONS,
	}

	if handlerFunc, exists := methodHandlers[method]; exists {
		if rb.group != nil {
			// 使用路由组的方法映射
			groupHandlers := map[string]func(string, ...gin.HandlerFunc) gin.IRoutes{
				"GET":     rb.group.GET,
				"POST":    rb.group.POST,
				"PUT":     rb.group.PUT,
				"DELETE":  rb.group.DELETE,
				"PATCH":   rb.group.PATCH,
				"HEAD":    rb.group.HEAD,
				"OPTIONS": rb.group.OPTIONS,
			}
			groupHandlers[method](path, handler)
		} else {
			handlerFunc(path, handler)
		}
	}

	// 记录路由信息
	routesMutex.Lock()
	defer routesMutex.Unlock()

	fullPath := path
	if rb.group != nil {
		// 获取路由组的基础路径
		fullPath = rb.getGroupBasePath() + path
	}

	routes[name] = &Route{
		Name:    name,
		Path:    fullPath,
		Method:  method,
		Handler: handler,
	}
}

// 获取路由组的基础路径
func (rb *RouteBuilder) getGroupBasePath() string {
	if rb.group == nil {
		return ""
	}
	// Gin路由组的基础路径需要从路由组对象中提取
	// 这里使用一个简单的方法来获取基础路径
	// 实际可能需要更复杂的逻辑来获取完整路径
	return "/"
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

	fullPath := path
	if rb.group != nil {
		fullPath = rb.getGroupBasePath() + path
	}

	routes[name] = &Route{
		Name:    name,
		Path:    fullPath,
		Method:  "ANY",
		Handler: handler,
	}
}

// BuildUrl 根据路由名称和参数生成URL
func BuildUrl(name string, params ...map[string]any) (string, error) {
	routesMutex.RLock()
	route, exists := routes[name]
	routesMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("路由不存在: %s", name)
	}

	path := route.Path

	if len(params) > 0 {
		for key, value := range params[0] {
			paramPlaceholder := ":" + key
			strValue := fmt.Sprintf("%v", value)
			path = strings.ReplaceAll(path, paramPlaceholder, strValue)
		}
	}

	return path, nil
}
