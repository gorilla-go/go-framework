package controller

import (
	"go-framework/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HomeController 首页控制器
type HomeController struct{}

// NewHomeController 创建首页控制器
func NewHomeController() *HomeController {
	return &HomeController{}
}

// Index 首页
func (c *HomeController) Index(ctx *gin.Context) {
	// 准备模板数据
	data := gin.H{
		"title":   "Go Framework",
		"now":     time.Now().Format("2006-01-02 15:04:05"),
		"message": "欢迎使用Go Framework构建Web应用",
		"year":    time.Now().Year(),
		"features": []string{
			"完整的项目结构",
			"用户认证与授权",
			"错误处理",
			"日志记录",
			"数据库操作",
			"RESTful API",
			"HTML模板渲染",
			"静态资源服务",
		},
	}

	// 渲染模板
	ctx.HTML(http.StatusOK, "index.html", data)
}

// ApiHome API首页
func (c *HomeController) ApiHome(ctx *gin.Context) {
	data := gin.H{
		"name":    "Go Framework API",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	}

	response.Success(ctx, data)
}

// Health 健康检查
func (c *HomeController) Health(ctx *gin.Context) {
	data := gin.H{
		"status":  "UP",
		"time":    time.Now().Format(time.RFC3339),
		"service": "go-framework",
	}

	response.Success(ctx, data)
}
