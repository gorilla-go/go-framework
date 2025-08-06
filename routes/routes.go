package routes

import (
	"go-framework/app/controller"
	"go-framework/pkg/router"
)

func init() {
	router.RegisterController(&controller.IndexController{})
}
