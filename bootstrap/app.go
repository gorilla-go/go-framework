package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"github.com/gorilla-go/go-framework/pkg/router"
	"go.uber.org/fx"
)

const (
	// ShutdownTimeout 服务器关闭超时时间
	ShutdownTimeout = 15 * time.Second
)

// 全局HTTP服务器实例，便于在信号处理中访问
var (
	httpServer     *http.Server
	httpServerLock sync.Mutex
)

// RegisterHooks 注册应用程序钩子
func RegisterHooks(lifecycle fx.Lifecycle, router *gin.Engine, cfg *config.Config) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 创建HTTP服务器
			srv := &http.Server{
				Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
				Handler: router,
				// 添加读写超时设置
				ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
				WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
				IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
			}

			// 保存到全局变量
			httpServerLock.Lock()
			httpServer = srv
			httpServerLock.Unlock()

			// 在单独的goroutine中启动服务器
			go func() {
				logger.Infof("HTTP服务器启动在 %d 端口", cfg.Server.Port)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatalf("HTTP服务器启动失败: %v", err)
				}
			}()

			logger.Info("服务器已准备就绪")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("正在关闭HTTP服务器...")

			// 获取服务器实例
			httpServerLock.Lock()
			srv := httpServer
			httpServerLock.Unlock()

			if srv == nil {
				logger.Warn("HTTP服务器实例为空，无法关闭")
				return nil
			}

			// 给服务器关闭的超时时间
			shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
			defer cancel()

			// 优雅关闭服务器
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Errorf("服务器关闭出错: %v", err)
				return err
			}

			logger.Info("服务器已优雅关闭")
			return nil
		},
	})
}

// NewApp 创建应用程序
func NewApp() *fx.App {
	app := fx.New(
		// 注册所有模块
		fx.Provide(provides...),
		fx.Populate(router.ConvertController()...),

		// 初始化日志
		fx.Invoke(func(cfg *config.Config) {
			logger.InitLogger(&cfg.Log)
		}),

		// 注册钩子
		fx.Invoke(RegisterHooks),
	)

	return app
}
