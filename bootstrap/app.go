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

// weakSecrets 已知的占位/弱密钥集合
var weakSecrets = map[string]bool{
	"":                   true,
	"your-secret-key":    true,
	"secret-key":         true,
	"session-secret-key": true,
	"change-me":          true,
}

// warnInsecureConfig 在生产（release）模式下检测 JWT/Session 密钥是否为空或默认占位值，
// 若是则发出安全告警，提示通过配置或环境变量设置强随机密钥。
func warnInsecureConfig(cfg *config.Config) {
	if cfg.IsDebug() {
		return
	}
	if weakSecrets[cfg.JWT.Secret] {
		logger.Warn("⚠️  安全告警: jwt.secret 为空或使用默认占位密钥，生产环境请通过环境变量 JWT_SECRET 设置强随机值")
	}
	if weakSecrets[cfg.Session.Secret] {
		logger.Warn("⚠️  安全告警: session.secret 为空或使用默认占位密钥，生产环境请通过环境变量 SESSION_SECRET 设置强随机值")
	}
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

			// 安全检查：生产模式下使用默认/空密钥时发出告警
			warnInsecureConfig(cfg)

			// 初始化模板引擎
			template.InitTemplateManager(cfg.Template, Config().IsDebug())
		}),

		// 控制器初始化（FX 注入控制器依赖）
		fx.Populate(func() []any {
			deps := make([]any, len(router.Controllers))
			for i, c := range router.Controllers {
				deps[i] = c
			}
			return deps
		}()...),

		// 注册钩子
		fx.Invoke(RegisterHooks),
	}

	// 根据运行模式设置日志级别
	if !Config().IsDebug() {
		fxOptions = append(fxOptions, fx.NopLogger)
	}

	return fx.New(fxOptions...)
}
