package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla-go/go-framework/bootstrap"
	"github.com/gorilla-go/go-framework/pkg/logger"
)

func main() {
	app := bootstrap.NewApp()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Infof("接收到信号: %s, 正在关闭应用...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), bootstrap.ShutdownTimeout)
		defer cancel()

		if err := app.Stop(ctx); err != nil {
			logger.Errorf("应用停止失败: %v", err)
			os.Exit(1)
		}
	}()

	app.Run()
}
