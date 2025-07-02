package middleware

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RouteBuilder 路由构建器
type RouteBuilder struct {
	router *gin.Engine
	group  *gin.RouterGroup
}

// 路由参数定义
type routeParam struct {
	Name    string
	Pattern string
}

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

// 解析路由路径，提取参数和模式
// 将类似 "/user/<id:\\d+>" 的路由转换为 "/user/:id" 并返回参数验证器
func (rb *RouteBuilder) parsePath(path string) (string, []routeParam) {
	paramPattern := regexp.MustCompile(`<([^>:]+)(?::([^>]+))?>`)
	matches := paramPattern.FindAllStringSubmatch(path, -1)

	params := []routeParam{}
	result := path

	for _, match := range matches {
		// 完整匹配的文本
		fullMatch := match[0]
		// 参数名
		paramName := match[1]
		// 正则模式（如果有）
		paramPattern := ".*"
		if len(match) > 2 && match[2] != "" {
			paramPattern = match[2]
		}

		// 将 <param:pattern> 替换为 :param
		result = strings.Replace(result, fullMatch, ":"+paramName, 1)

		// 添加到参数列表
		params = append(params, routeParam{
			Name:    paramName,
			Pattern: paramPattern,
		})
	}

	return result, params
}

// 创建参数验证中间件
func createParamValidators(params []routeParam) []gin.HandlerFunc {
	validators := []gin.HandlerFunc{}

	for _, param := range params {
		pattern := regexp.MustCompile(param.Pattern)

		validators = append(validators, func(c *gin.Context) {
			value := c.Param(param.Name)
			if !pattern.MatchString(value) {
				c.AbortWithStatusJSON(400, gin.H{
					"code":    400,
					"message": "参数 " + param.Name + " 格式不正确",
					"data":    nil,
				})
				return
			}
			c.Next()
		})
	}

	return validators
}

// GET 注册GET请求路由，支持Flask风格路径参数
func (rb *RouteBuilder) GET(path string, handlers ...gin.HandlerFunc) {
	ginPath, params := rb.parsePath(path)
	validators := createParamValidators(params)

	if rb.group != nil {
		rb.group.GET(ginPath, append(validators, handlers...)...)
	} else {
		rb.router.GET(ginPath, append(validators, handlers...)...)
	}
}

// POST 注册POST请求路由，支持Flask风格路径参数
func (rb *RouteBuilder) POST(path string, handlers ...gin.HandlerFunc) {
	ginPath, params := rb.parsePath(path)
	validators := createParamValidators(params)

	if rb.group != nil {
		rb.group.POST(ginPath, append(validators, handlers...)...)
	} else {
		rb.router.POST(ginPath, append(validators, handlers...)...)
	}
}

// PUT 注册PUT请求路由，支持Flask风格路径参数
func (rb *RouteBuilder) PUT(path string, handlers ...gin.HandlerFunc) {
	ginPath, params := rb.parsePath(path)
	validators := createParamValidators(params)

	if rb.group != nil {
		rb.group.PUT(ginPath, append(validators, handlers...)...)
	} else {
		rb.router.PUT(ginPath, append(validators, handlers...)...)
	}
}

// DELETE 注册DELETE请求路由，支持Flask风格路径参数
func (rb *RouteBuilder) DELETE(path string, handlers ...gin.HandlerFunc) {
	ginPath, params := rb.parsePath(path)
	validators := createParamValidators(params)

	if rb.group != nil {
		rb.group.DELETE(ginPath, append(validators, handlers...)...)
	} else {
		rb.router.DELETE(ginPath, append(validators, handlers...)...)
	}
}

// 其他HTTP方法可以按需添加...

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
