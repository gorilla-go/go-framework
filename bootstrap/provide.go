package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/database"
	"github.com/gorilla-go/go-framework/pkg/eventbus"
	"github.com/gorilla-go/go-framework/pkg/router"
	"gorm.io/gorm"

	_ "github.com/gorilla-go/go-framework/routes"
)

// 全局注册器
var Providers = []any{
	Config,
	EventBus,
	Database,
	Controllers,
	Middlewares,
	Router,
}

// 全局配置
func Config() *config.Config {
	cfg, err := config.Fetch()
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	return cfg
}

// 提供数据库连接
func Database(cfg *config.Config) *gorm.DB {
	db, err := database.Init(&cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("初始化数据库失败: %v", err))
	}
	return db
}

func Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// 提供路由器
func Router(controllers []router.IController, cfg *config.Config, middlewares []gin.HandlerFunc) *gin.Engine {
	router := &router.Router{
		Controllers: controllers,
		Cfg:         cfg,
		Middlewares: middlewares,
	}
	return router.Route()
}

// 提供控制器列表
func Controllers() []router.IController {
	return router.Controllers
}

// 提供事件注册器
func EventBus() *eventbus.EventBus {
	return eventbus.New()
}
