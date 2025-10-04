.PHONY: all dev build run stop start startd help clean install gulp-build

# 默认目标
all: devs

# 安装依赖
install:
	@echo "安装 Node.js 依赖..."
	@cd static && npm install

# Gulp 构建静态资源
gulp-build:
	@echo "构建静态资源..."
	@cd static && npm run build


# 清理并启动开发环境
devs: clean dev

# 开发模式（带热重载）
dev: gulp-build
	@PORT=$$(./scripts/get-port.sh); \
	echo "🔍 检查端口 $$PORT 占用情况..."; \
	if lsof -ti :$$PORT >/dev/null 2>&1; then \
		echo "⚠️  端口 $$PORT 被占用，请运行 'make clean' 清理后重试"; \
		echo "💡 或者直接运行 'make devs' 自动清理并启动"; \
		exit 1; \
	fi; \
	AIR_PATH=~/go/bin/air; \
	if [ ! -f "$$AIR_PATH" ]; then \
		echo "Air未安装, 正在安装..."; \
		go install github.com/air-verse/air@latest; \
	fi; \
	export SERVER_MODE=debug; \
	echo "🚀 启动开发环境 (端口: $$PORT), 监控文件变更..."; \
	(cd static && npm run watch &); \
	$$AIR_PATH

# 生产环境构建
build: install gulp-build
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/app ./cmd/main.go

# 运行（不带热重载）
run:
	go run ./cmd/main.go

# 前台启动
start: build
	@echo "前台启动应用程序..."
	@export SERVER_MODE=release; bin/app

# 后台启动
startd: build
	@echo "后台启动应用程序..."
	@mkdir -p logs
	@export SERVER_MODE=release; nohup bin/app > logs/app.out 2>&1 & echo $$! > .pid
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

# 清理临时文件和日志（不清理进程）
clean:
	@echo "🧹 清理开发环境..."
	@./scripts/cleanup.sh
	@echo "清理临时文件..."
	@rm -rf tmp/*
	@rm -f .pid
	@echo "清理完成"

# 帮助
help:
	@echo "Go Framework 开发工具"
	@echo ""
	@echo "可用命令:"
	@echo "  🚀 开发相关:"
	@echo "    make dev         - 以开发模式运行 (带热重载)"
	@echo "    make devs        - 清理并启动开发环境 (推荐)"
	@echo ""
	@echo "  🏗️  构建相关:"
	@echo "    make build       - 构建应用程序"
	@echo "    make run         - 运行应用程序 (不带热重载)"
	@echo "    make install     - 安装前端依赖"
	@echo "    make gulp-build  - 构建静态资源"
	@echo ""
	@echo "  🔧 生产环境:"
	@echo "    make start       - 构建并在前台启动生产服务"
	@echo "    make startd     - 构建并在后台启动生产服务"
	@echo "    make stop        - 停止运行的应用程序"
	@echo ""
	@echo "  🧹 清理相关:"
	@echo "    make clean       - 清理临时文件"
	@echo ""
	@echo "  💡 推荐流程:"
	@echo "    1. 开发时: make devs"
	@echo "    2. 遇到端口冲突: make clean"
	@echo "    3. Ctrl+C 退出后: make clean (清理孤儿进程)"