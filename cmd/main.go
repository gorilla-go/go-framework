package main

import (
	"context"
	"go-framework/internal/app"
	"go-framework/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 创建应用
	application := app.NewApp()

	// 创建一个信号通道
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的goroutine中等待信号
	go func() {
		// 等待信号
		sig := <-sigCh
		logger.Infof("接收到信号: %s, 正在关闭应用...", sig)

		// 停止应用
		ctx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout)
		defer cancel()

		// 这将触发所有OnStop钩子
		if err := application.Stop(ctx); err != nil {
			logger.Errorf("应用停止失败: %v", err)
			os.Exit(1)
		}

		logger.Info("应用已完全关闭")
	}()

	// 在主线程中运行应用，这将阻塞直到应用停止
	logger.Info("正在启动应用...")
	application.Run()
}
