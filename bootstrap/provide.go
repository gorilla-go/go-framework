package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/database"
	"github.com/gorilla-go/go-framework/pkg/eventbus"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"github.com/gorilla-go/go-framework/pkg/router"
	"github.com/gorilla-go/go-framework/pkg/template"
	"gorm.io/gorm"
)

// 提供配置
func ProvideConfig() *config.Config {
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Fatalf("加载配置失败: %v", err)
	}
	return cfg
}

// 提供数据库连接
func ProvideDatabase(cfg *config.Config) (*gorm.DB, error) {
	return database.InitDB(&cfg.Database)
}

// 提供模板管理器
func ProvideTemplateManager(cfg *config.Config) *template.TemplateManager {
	tm := template.NewTemplateManager(
		cfg.Template.Path,
		cfg.Template.Layouts,
		cfg.Template.Extension,
		cfg.Server.Mode == "debug",
	)

	// 确保全局实例被初始化
	template.InitGlobalTemplateManager(tm)

	return tm
}

// 提供路由器
func ProvideRouter(controllers []router.RouterAnnotation, cfg *config.Config) *router.Router {
	return &router.Router{
		Controllers: controllers,
		Cfg:         cfg,
	}
}

// 提供HTTP服务器
func ProvideServer(router *router.Router) *gin.Engine {
	return router.Route()
}

// 提供控制器列表
func ProvideControllers() []router.RouterAnnotation {
	return router.Controllers
}

// 提供事件注册器
func ProvideEventBus() *eventbus.EventBus {
	return eventbus.New()
}
