package router

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// RouteBuilder 路由构建器
type RouteBuilder struct {
	router   *gin.Engine
	group    *gin.RouterGroup
	basePath string
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
	var group *gin.RouterGroup
	newBasePath := rb.basePath + path

	if rb.group != nil {
		group = rb.group.Group(path)
	} else {
		group = rb.router.Group(path)
	}

	return &RouteBuilder{
		router:   rb.router,
		group:    group,
		basePath: newBasePath,
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

	// 注册到Gin
	target := rb.getRouteTarget()
	switch method {
	case "GET":
		target.GET(path, handler)
	case "POST":
		target.POST(path, handler)
	case "PUT":
		target.PUT(path, handler)
	case "DELETE":
		target.DELETE(path, handler)
	case "PATCH":
		target.PATCH(path, handler)
	case "HEAD":
		target.HEAD(path, handler)
	case "OPTIONS":
		target.OPTIONS(path, handler)
	case "ANY":
		target.Any(path, handler)
	}

	// 记录路由信息
	fullPath := rb.basePath + path

	routesMutex.Lock()
	routes[name] = &Route{
		Name:    name,
		Path:    fullPath,
		Method:  method,
		Handler: handler,
	}
	routesMutex.Unlock()
}

// getRouteTarget 获取路由注册目标（路由组或根路由）
func (rb *RouteBuilder) getRouteTarget() gin.IRoutes {
	if rb.group != nil {
		return rb.group
	}
	return rb.router
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
	rb.registerRoute("ANY", path, name, handler)
}

// BuildUrl 根据路由名称和参数生成URL，路由不存在或缺少参数时返回错误
func BuildUrl(name string, params ...map[string]any) (string, error) {
	routesMutex.RLock()
	route, exists := routes[name]
	routesMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("路由不存在: %s", name)
	}

	path := route.Path

	// 替换路径参数
	if len(params) > 0 {
		for key, value := range params[0] {
			path = strings.ReplaceAll(path, ":"+key, fmt.Sprintf("%v", value))
		}
	}

	// 检查是否还有未替换的参数
	if strings.Contains(path, ":") {
		var missing []string
		parts := strings.SplitSeq(path, "/")
		for part := range parts {
			if after, ok := strings.CutPrefix(part, ":"); ok {
				missing = append(missing, after)
			}
		}
		return "", fmt.Errorf("缺少路径参数: %s", strings.Join(missing, ", "))
	}

	return path, nil
}
