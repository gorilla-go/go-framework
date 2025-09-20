package router

type IController interface {
	RouterAnnotation
}

// 路由注册控制器
var Controllers = []IController{}

func RegisterControllers(controller ...IController) {
	Controllers = append(Controllers, controller...)
}

func ConvertController() []any {
	anyControllers := make([]any, len(Controllers))
	for i, controller := range Controllers {
		anyControllers[i] = any(controller)
	}
	return anyControllers
}
