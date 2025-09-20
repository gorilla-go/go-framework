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

	_ "github.com/gorilla-go/go-framework/routes"
)

// 全局配置
func Config() *config.Config {
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Fatalf("加载配置失败: %v", err)
	}
	return cfg
}

// 提供数据库连接
func Database(cfg *config.Config) *gorm.DB {
	db, err := database.InitDB(&cfg.Database)
	if err != nil {
		logger.Fatalf("初始化数据库失败: %v", err)
	}
	return db
}

// 提供模板管理器
func TemplateManager(cfg *config.Config) *template.TemplateManager {
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
func Router(controllers []router.IController, cfg *config.Config) *gin.Engine {
	router := &router.Router{
		Controllers: controllers,
		Cfg:         cfg,
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
