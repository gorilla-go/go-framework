package main

import (
	"fmt"
	"go-framework/internal/router"
	"go-framework/pkg/config"
	"go-framework/pkg/logger"
	"log"
	"os"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.InitLogger(&cfg.Log); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 创建路由
	r := router.SetupRouter()

	// 启动HTTP服务器
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("HTTP服务器启动在 %d 端口", cfg.Server.Port)
	log.Fatal(r.Run(port))
}
