.PHONY: dev build run clean

# 默认目标
all: dev

# 开发模式（带热重载）
dev:
	./start_dev.sh

# 构建
build:
	go build -o bin/app ./cmd/main.go

# 运行（不带热重载）
run:
	go run ./cmd/main.go

# 清理临时文件
clean:
	rm -rf tmp/
	rm -f bin/app

# 帮助
help:
	@echo "Go Framework 开发工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make dev    - 以开发模式运行 (带热重载)"
	@echo "  make build  - 构建应用程序"
	@echo "  make run    - 运行应用程序 (不带热重载)"
	@echo "  make clean  - 清理临时文件" 