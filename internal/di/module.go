package di

import (
	"go-framework/internal/controller"
	"go-framework/internal/router"
	"go-framework/pkg/config"
	"go-framework/pkg/database"
	"go-framework/pkg/eventbus"
	"go-framework/pkg/logger"
	"go-framework/pkg/middleware"
	"go-framework/pkg/template"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
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
func ProvideRouter(controllers []middleware.RouterAnnotation) *router.Router {
	return &router.Router{
		Controllers: controllers,
	}
}

// 提供HTTP服务器
func ProvideServer(router *router.Router) *gin.Engine {
	return router.Route()
}

// 提供控制器列表
func ProvideControllers() []middleware.RouterAnnotation {
	return controllers
}

// 提供事件注册器
func ProvideEventBus() *eventbus.EventBus {
	return eventbus.New()
}

// 转换控制器数组为任意类型数组，用于依赖注入
func ConvertControllerArrToAny(controllers []middleware.RouterAnnotation) []any {
	anyControllers := make([]any, len(controllers))
	for i, controller := range controllers {
		anyControllers[i] = any(controller)
	}
	return anyControllers
}

// 注册所有模块
var Module = fx.Options(
	fx.Provide(
		ProvideConfig,
		ProvideEventBus,
		ProvideDatabase,
		ProvideTemplateManager,
		ProvideControllers,
		ProvideRouter,
		ProvideServer,
	),
	fx.Populate(ConvertControllerArrToAny(controllers)...),
)

// 路由注册控制器
var controllers = []middleware.RouterAnnotation{
	&controller.IndexController{},
}
