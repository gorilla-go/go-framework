package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"github.com/gorilla-go/go-framework/pkg/middleware"
	"github.com/gorilla-go/go-framework/pkg/response"
)

type Router struct {
	Controllers []IController
	Cfg         *config.Config
	Middlewares []gin.HandlerFunc
}

// Route 设置路由
func (router *Router) Route() *gin.Engine {
	// 使用全局配置
	cfg := router.Cfg
	if cfg == nil {
		logger.Fatal("配置未初始化")
	}

	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

	// 添加全局中间件
	r.Use(
		middleware.RecoveryMiddleware(),
		gin.Logger(),
		middleware.LoggerMiddleware(),
		middleware.SecurityMiddleware(),
		middleware.SessionMiddleware(
			&router.Cfg.Session,
			&router.Cfg.Redis,
		),
	)

	// 添加自定义中间件
	if len(router.Middlewares) > 0 {
		r.Use(router.Middlewares...)
	}

	// 根据配置启用 gzip 压缩
	if cfg.Gzip.Enabled {
		r.Use(middleware.GzipWithLevelMiddleware(cfg.Gzip.Level))
	}

	// 根据配置启用全局限流
	if cfg.Server.EnableRateLimit {
		r.Use(middleware.RateLimitMiddleware(cfg.Server.RateLimit, cfg.Server.RateBurst))
	}

	// 静态文件
	r.Static("/static", cfg.Static.Path)

	// 创建路由构建器
	rb := NewRouteBuilder(r)

	// 注册控制器路由
	for _, controller := range router.Controllers {
		controller.Annotation(rb)
	}

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, errors.NewNotFound("", nil))
	})

	return r
}
