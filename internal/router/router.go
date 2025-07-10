package router

import (
	"go-framework/pkg/config"
	"go-framework/pkg/logger"
	"go-framework/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Controllers []middleware.RouterAnnotation
}

// Route 设置路由
func (router *Router) Route() *gin.Engine {
	// 使用全局配置
	cfg := config.GetConfig()
	if cfg == nil {
		logger.Fatal("配置未初始化")
	}

	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

	// 添加全局中间件
	r.Use(middleware.RecoveryMiddleware()) // 错误恢复，应该最先加载
	r.Use(gin.Logger())                    // gin 请求日志
	r.Use(middleware.LoggerMiddleware())   // 请求日志
	r.Use(middleware.SecurityMiddleware()) // 安全相关头部
	r.Use(middleware.SessionMiddleware())  // 会话管理

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
	rb := middleware.NewRouteBuilder(r)

	// 注册控制器路由
	for _, controller := range router.Controllers {
		controller.Annotation(rb)
	}

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		middleware.HandleNotFound(c, "页面不存在", nil)
	})

	return r
}
