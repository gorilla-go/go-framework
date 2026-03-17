package router

// IController 控制器接口，所有控制器必须实现该接口
type IController interface {
	Annotation(rb *RouteBuilder)
}

// Controllers 已注册的控制器列表
var Controllers = []IController{}

// RegisterControllers 注册控制器
func RegisterControllers(controller ...IController) {
	Controllers = append(Controllers, controller...)
}
