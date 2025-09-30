package routes

import (
	"github.com/gorilla-go/go-framework/app/controller"
	"github.com/gorilla-go/go-framework/pkg/router"
)

func init() {
	router.RegisterControllers(
		&controller.IndexController{},
		&controller.SystemController{},
	)
}
