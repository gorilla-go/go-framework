package router

import "go-framework/pkg/middleware"

// 路由注册控制器
var Controllers = []middleware.RouterAnnotation{}

func RegisterControllers(controller ...middleware.RouterAnnotation) {
	Controllers = append(Controllers, controller...)
}

func ConvertController() []any {
	anyControllers := make([]any, len(Controllers))
	for i, controller := range Controllers {
		anyControllers[i] = any(controller)
	}
	return anyControllers
}
