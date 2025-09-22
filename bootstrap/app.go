package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"github.com/gorilla-go/go-framework/pkg/router"
	"github.com/gorilla-go/go-framework/pkg/template"
	"go.uber.org/fx"
)

const (
	// ShutdownTimeout 服务器关闭超时时间
	ShutdownTimeout = 15 * time.Second
)

// 全局HTTP服务器实例，便于在信号处理中访问
var (
	httpServer *http.Server
)

// RegisterHooks 注册应用程序钩子
func RegisterHooks(lifecycle fx.Lifecycle, router *gin.Engine, cfg *config.Config) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			httpServer = &http.Server{
				Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
				Handler:      router,
				ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
				WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
				IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
			}

			go func() {
				logger.Infof("HTTP服务器启动在端口: %d", cfg.Server.Port)
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatalf("HTTP服务器启动失败: %v", err)
				}
			}()

			logger.Info("服务器已准备就绪")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("正在关闭HTTP服务器...")

			if httpServer == nil {
				return nil
			}

			shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
			defer cancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Errorf("服务器关闭出错: %v", err)
				return err
			}

			logger.Info("服务器已关闭")
			return nil
		},
	})
}

// NewApp 创建应用程序
func NewApp() *fx.App {
	app := fx.New(
		// 注册所有模块
		fx.Provide([]any{
			Config,
			EventBus,
			Database,
			Controllers,
			Router,
		}...),

		// 初始化
		fx.Invoke(func(cfg *config.Config) {
			// 初始化日志
			logger.InitLogger(&cfg.Log)

			// 初始化模板引擎
			template.InitTemplateManager(cfg.Template, cfg.Server.Mode == "debug")
		}),

		// 控制器初始化
		fx.Populate(router.ConvertController()...),

		// 注册钩子
		fx.Invoke(RegisterHooks),
	)

	return app
}
