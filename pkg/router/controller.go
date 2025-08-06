package router

// 路由注册控制器
var Controllers = []RouterAnnotation{}

func RegisterControllers(controller ...RouterAnnotation) {
	Controllers = append(Controllers, controller...)
}

func ConvertController() []any {
	anyControllers := make([]any, len(Controllers))
	for i, controller := range Controllers {
		anyControllers[i] = any(controller)
	}
	return anyControllers
}
