package controller

import (
	"go-framework/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// DemoController 演示控制器
type DemoController struct{}

// NewDemoController 创建演示控制器
func NewDemoController() *DemoController {
	return &DemoController{}
}

// Demo 演示API
func (c *DemoController) Demo(ctx *gin.Context) {
	data := gin.H{
		"name":    "演示API",
		"version": "1.0.0",
		"time": gin.H{
			"now": time.Now().Format(time.RFC3339),
		},
		"features": []string{
			"这是一个演示API",
			"在开发模式下不需要数据库",
			"可以用于测试前端",
		},
	}

	response.Success(ctx, data)
}
