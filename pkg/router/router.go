package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/logger"
	"github.com/gorilla-go/go-framework/pkg/middleware"
)

type Router struct {
	Controllers []IController
	Cfg         *config.Config
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
		middleware.Recovery(),
		middleware.Logger(cfg.IsDebug()),
		middleware.SessionStart(
			&router.Cfg.Session,
			&router.Cfg.Redis,
			&router.Cfg.Database,
		),
	)

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

	// 404处理：根据 Accept 头返回 JSON 或纯文本
	r.NoRoute(func(c *gin.Context) {
		if c.NegotiateFormat(gin.MIMEJSON, gin.MIMEHTML) == gin.MIMEJSON {
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "Not Found"})
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	return r
}
