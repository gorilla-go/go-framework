package routes

import (
	"go-framework/app/controller"
	"go-framework/pkg/router"
)

func init() {
	router.RegisterControllers(
		&controller.IndexController{},
	)
}
