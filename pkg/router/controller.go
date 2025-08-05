package router

import "go-framework/pkg/middleware"

// 路由注册控制器
var controllers = []middleware.RouterAnnotation{}

func GetControllers() []middleware.RouterAnnotation {
	return controllers
}

func RegisterController(controller middleware.RouterAnnotation) {
	controllers = append(controllers, controller)
}

func ConvertController() []any {
	cs := GetControllers()
	anyControllers := make([]any, len(cs))
	for i, controller := range cs {
		anyControllers[i] = any(controller)
	}
	return anyControllers
}
