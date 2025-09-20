.PHONY: all dev build run stop start start-d help clean install-deps gulp-build

# 默认目标
all: dev

# 安装依赖
install-deps:
	@echo "安装 Node.js 依赖..."
	@cd static && npm install

# Gulp 构建静态资源
gulp-build:
	@echo "构建静态资源..."
	@cd static && npm run build

# 开发模式（带热重载）
dev: gulp-build 
	@AIR_PATH=~/go/bin/air; \
	if [ ! -f "$$AIR_PATH" ]; then \
		echo "Air未安装, 正在安装..."; \
		go install github.com/air-verse/air@latest; \
	fi; \
	export GIN_MODE=debug; \
	echo "启动开发环境, 监控文件变更..."; \
	(cd static && npm run watch &); \
	$$AIR_PATH

# 生产环境构建
build: install-deps gulp-build
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/app ./cmd/main.go

# 运行（不带热重载）
run:
	go run ./cmd/main.go

# 前台启动
start: build stop
	@echo "前台启动应用程序..."
	@bin/app

# 后台启动
start-d: build stop
	@echo "后台启动应用程序..."
	@mkdir -p logs
	@nohup bin/app > logs/app.out 2>&1 & echo $$! > .pid
	@echo "应用程序已在后台启动, PID: $$(cat .pid)"

# 生产环境停止
stop:
	@if [ -f ".pid" ]; then \
		PID=$$(cat .pid); \
		if ps -p $$PID > /dev/null; then \
			echo "停止进程 $$PID..."; \
			kill -TERM $$PID; \
			echo "已发送终止信号"; \
			rm -f .pid; \
		else \
			echo "PID文件存在但进程不存在, 清理PID文件"; \
			rm -f .pid; \
		fi; \
	else \
		PID=$$(pgrep -f "bin/app"); \
		if [ -n "$$PID" ]; then \
			echo "停止进程 $$PID..."; \
			kill -TERM $$PID; \
			echo "已发送终止信号"; \
		else \
			echo "未找到运行中的应用程序进程"; \
		fi; \
	fi

# 清理临时文件和日志
clean:
	@echo "清理临时文件..."
	@rm -rf tmp/*
	@rm -f .pid
	@echo "清理完成"

# 帮助
help:
	@echo "Go Framework 开发工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make dev         - 以开发模式运行 (带热重载)"
	@echo "  make build       - 构建应用程序"
	@echo "  make run         - 运行应用程序 (不带热重载)"
	@echo "  make start       - 构建并在前台启动生产服务"
	@echo "  make start-d     - 构建并在后台启动生产服务"
	@echo "  make stop        - 停止运行的应用程序"
	@echo "  make clean       - 清理临时文件和进程"
	@echo "  make install-deps - 安装前端依赖"
	@echo "  make gulp-build  - 构建静态资源" 