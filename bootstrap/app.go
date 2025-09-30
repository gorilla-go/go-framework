package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
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

// 全局注册器
var Providers = []any{
	Config,
	EventBus,
	Database,
	Controllers,
	Router,
}

// 全局HTTP服务器实例，便于在信号处理中访问
var (
	httpServer *http.Server
)

// printStartupBanner 打印启动 Logo 和服务信息
func printStartupBanner(cfg *config.Config) {
	banner := `
   ____           _____                                        __
  / ___| ___     |  ___| __ __ _ _ __ ___   _____      _____ _ __ | | __
 | |  _ / _ \    | |_ | '__/ _' | '_ ' _ \ / _ \ \ /\ / / _ \ '__|| |/ /
 | |_| | (_) |   |  _|| | | (_| | | | | | |  __/\ V  V / (_) | |   |   <
  \____|\___/    |_|  |_|  \__,_|_| |_| |_|\___| \_/\_/ \___/|_|   |_|\_\
`
	// ANSI 颜色代码
	const (
		colorReset  = "\033[0m"
		colorCyan   = "\033[36m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorBold   = "\033[1m"
	)

	fmt.Println(colorCyan + banner + colorReset)
	fmt.Printf("%s%s🚀 Server is running!%s\n\n", colorBold, colorGreen, colorReset)
	fmt.Printf("  %s➜%s Local:    %shttp://0.0.0.0:%d%s\n", colorGreen, colorReset, colorCyan, cfg.Server.Port, colorReset)
	fmt.Printf("  %s➜%s Mode:     %s%s%s\n", colorGreen, colorReset, colorYellow, cfg.Server.Mode, colorReset)
	fmt.Printf("  %s➜%s PID:      %s%d%s\n\n", colorGreen, colorReset, colorBlue, os.Getpid(), colorReset)

	if cfg.Server.EnableRateLimit {
		fmt.Printf("  %s⚡ Rate Limit:%s %d req/s (burst: %d)\n", colorPurple, colorReset, cfg.Server.RateLimit, cfg.Server.RateBurst)
	}

	fmt.Printf("\n  %sPress Ctrl+C to stop%s\n\n", colorYellow, colorReset)
}

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
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatalf("HTTP服务器启动失败: %v", err)
				}
			}()

			// 打印启动 Logo
			printStartupBanner(cfg)
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

	// 根据运行模式设置 FX 选项
	fxOptions := []fx.Option{
		// 注册所有模块
		fx.Provide(Providers...),

		// 初始化
		fx.Invoke(func(cfg *config.Config) {
			// 初始化日志
			logger.InitLogger(&cfg.Log)

			// 初始化模板引擎
			template.InitTemplateManager(cfg.Template, Config().IsDebug())
		}),

		// 控制器初始化
		fx.Populate(router.ConvertController()...),

		// 注册钩子
		fx.Invoke(RegisterHooks),
	}

	// 根据运行模式设置日志级别
	if !Config().IsDebug() {
		fxOptions = append(fxOptions, fx.NopLogger)
	}

	return fx.New(fxOptions...)
}
