package routes

import (
	"github.com/gorilla-go/go-framework/app/controller"
	"github.com/gorilla-go/go-framework/pkg/router"
)

func init() {
	router.RegisterControllers(
		&controller.IndexController{},

		// 演示控制器
		&controller.DemoAPIController{},   // GET/POST/DELETE /demo/api/users[/:id]
		&controller.DemoAuthController{},  // POST /demo/auth/login, GET /demo/auth/profile|admin-only
		&controller.DemoEventController{}, // POST/GET/DELETE /demo/events/...
	)
}
