package bootstrap

import (
	"github.com/gin-gonic/gin"
	"go-framework/pkg/config"
	"go-framework/pkg/database"
	"go-framework/pkg/eventbus"
	"go-framework/pkg/logger"
	"go-framework/pkg/middleware"
	pkgRouter "go-framework/pkg/router"
	"go-framework/pkg/template"
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

	// 开发模式下不缓存模板且显示错误
	if cfg.Server.Mode == "debug" {
		tm.SetDevelopmentMode(true)
		tm.SetShowErrors(true)
	} else {
		tm.SetShowErrors(false)
	}

	// 确保全局实例被初始化
	template.InitGlobalTemplateManager(tm)

	return tm
}

// 提供路由器
func ProvideRouter(controllers []middleware.RouterAnnotation) *pkgRouter.Router {
	return &pkgRouter.Router{
		Controllers: controllers,
	}
}

// 提供HTTP服务器
func ProvideServer(router *pkgRouter.Router) *gin.Engine {
	return router.Route()
}

// 提供控制器列表
func ProvideControllers() []middleware.RouterAnnotation {
	return pkgRouter.GetControllers()
}

// 提供事件注册器
func ProvideEventBus() *eventbus.EventBus {
	return eventbus.New()
}
