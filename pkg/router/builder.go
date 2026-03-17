package router

import (
	stderrors "errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
)

// HandlerFunc 支持直接返回 error 的 handler 类型
type HandlerFunc func(*gin.Context) error

// wrapH 将 HandlerFunc 包装为标准 gin.HandlerFunc，统一在 router 层处理错误
func wrapH(f HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := f(c); err != nil {
			var appErr *errors.AppError
			if stderrors.As(err, &appErr) {
				response.Fail(c, appErr)
			} else {
				response.Fail(c, errors.NewInternalServerError(err.Error(), err))
			}
		}
	}
}

// RouteBuilder 路由构建器
type RouteBuilder struct {
	router   *gin.Engine
	group    *gin.RouterGroup
	basePath string
}

// Route 路由信息
type Route struct {
	Name   string
	Path   string
	Method string
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

// Group 创建路由组，可选传入组级中间件（参考 Echo group middleware 设计）
// 组级中间件只作用于该组内的路由，例如：rb.Group("/admin", middleware.JWT())
func (rb *RouteBuilder) Group(path string, middleware ...gin.HandlerFunc) *RouteBuilder {
	var group *gin.RouterGroup
	newBasePath := rb.basePath + path

	if rb.group != nil {
		group = rb.group.Group(path, middleware...)
	} else {
		group = rb.router.Group(path, middleware...)
	}

	return &RouteBuilder{
		router:   rb.router,
		group:    group,
		basePath: newBasePath,
	}
}

// GET 注册GET请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) GET(path string, handler HandlerFunc, name string) {
	rb.registerRoute("GET", path, name, handler)
}

// POST 注册POST请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) POST(path string, handler HandlerFunc, name string) {
	rb.registerRoute("POST", path, name, handler)
}

// PUT 注册PUT请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) PUT(path string, handler HandlerFunc, name string) {
	rb.registerRoute("PUT", path, name, handler)
}

// DELETE 注册DELETE请求路由，name参数用于在模板中使用route函数生成URL
func (rb *RouteBuilder) DELETE(path string, handler HandlerFunc, name string) {
	rb.registerRoute("DELETE", path, name, handler)
}

// PATCH 注册PATCH请求路由
func (rb *RouteBuilder) PATCH(path string, handler HandlerFunc, name string) {
	rb.registerRoute("PATCH", path, name, handler)
}

// HEAD 注册HEAD请求路由
func (rb *RouteBuilder) HEAD(path string, handler HandlerFunc, name string) {
	rb.registerRoute("HEAD", path, name, handler)
}

// OPTIONS 注册OPTIONS请求路由
func (rb *RouteBuilder) OPTIONS(path string, handler HandlerFunc, name string) {
	rb.registerRoute("OPTIONS", path, name, handler)
}

// ANY 注册所有HTTP方法路由
func (rb *RouteBuilder) ANY(path string, handler HandlerFunc, name string) {
	rb.registerRoute("ANY", path, name, handler)
}

// 注册路由，内部函数
func (rb *RouteBuilder) registerRoute(method, path, name string, handler HandlerFunc) {
	if name == "" {
		name = fmt.Sprintf("%s:%s", method, path)
	}

	wrapped := wrapH(handler)

	// 注册到Gin
	target := rb.getRouteTarget()
	switch method {
	case "GET":
		target.GET(path, wrapped)
	case "POST":
		target.POST(path, wrapped)
	case "PUT":
		target.PUT(path, wrapped)
	case "DELETE":
		target.DELETE(path, wrapped)
	case "PATCH":
		target.PATCH(path, wrapped)
	case "HEAD":
		target.HEAD(path, wrapped)
	case "OPTIONS":
		target.OPTIONS(path, wrapped)
	case "ANY":
		target.Any(path, wrapped)
	}

	// 记录路由信息
	fullPath := rb.basePath + path

	routesMutex.Lock()
	routes[name] = &Route{
		Name:   name,
		Path:   fullPath,
		Method: method,
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
