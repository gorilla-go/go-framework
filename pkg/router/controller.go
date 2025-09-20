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
	controllers := make([]any, len(Controllers))
	for i, v := range Controllers {
		controllers[i] = v
	}
	return controllers
}
