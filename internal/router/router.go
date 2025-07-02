package router

import (
	"go-framework/internal/controller"
	"go-framework/pkg/config"
	"go-framework/pkg/middleware"
	"go-framework/pkg/response"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 设置运行模式
	gin.SetMode(config.GetConfig().Server.Mode)

	// 创建路由
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.RateLimitMiddleware(100, 100)) // 每秒100个请求，容量100

	// 静态文件
	r.Static("/static", config.GetConfig().Static.Path)

	// 模板
	templatePath := filepath.Join(config.GetConfig().Template.Path, "*"+config.GetConfig().Template.Extension)
	r.LoadHTMLGlob(templatePath)

	// 初始化控制器
	homeController := controller.NewHomeController()
	demoController := controller.NewDemoController()

	// 创建路由构建器 - 支持Flask风格的路由参数
	rb := middleware.NewRouteBuilder(r)

	// Web路由 - 集中定义
	rb.GET("/", homeController.Index)
	rb.GET("/api", homeController.ApiHome)
	rb.GET("/health", homeController.Health)

	// API路由组 - 集中定义
	apiGroup := rb.Group("/api/v1")

	// 演示API
	apiGroup.GET("/demo", demoController.Demo)

	// 带正则验证的用户路由
	apiGroup.GET("/user/<id:\\d+>", func(c *gin.Context) {
		id := c.Param("id")
		response.Success(c, gin.H{
			"id":      id,
			"name":    "用户: " + id,
			"message": "获取用户详情成功",
		})
	})

	// 带多参数验证的产品路由
	apiGroup.GET("/product/<category:[a-zA-Z0-9-]+>/<id:\\d+>", func(c *gin.Context) {
		category := c.Param("category")
		id := middleware.ParseParam(c, "id", 0).(int)
		response.Success(c, gin.H{
			"category": category,
			"id":       id,
			"name":     category + "产品" + strconv.Itoa(id),
			"message":  "获取产品详情成功",
		})
	})

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path[:4] == "/api" {
			// API 404
			c.JSON(404, gin.H{
				"code":    404,
				"message": "API路径不存在",
				"data":    nil,
			})
		} else {
			// Web 404
			c.HTML(404, "index.html", gin.H{
				"title":   "页面不存在 - Gin模板项目",
				"message": "页面不存在",
			})
		}
	})

	return r
}
