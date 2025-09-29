#!/bin/bash

# 清理脚本 - 杀死孤儿进程和清理开发环境

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认端口
PORT=${1:-8081}

echo -e "${BLUE}🧹 开始清理开发环境...${NC}"

# 1. 检查并杀死占用指定端口的进程
echo -e "${YELLOW}🔍 检查端口 $PORT 使用情况...${NC}"
PIDS=$(lsof -ti :$PORT 2>/dev/null || true)

if [ -n "$PIDS" ]; then
    echo -e "${YELLOW}⚠️  发现占用端口 $PORT 的进程: $PIDS${NC}"
    for PID in $PIDS; do
        if kill -0 $PID 2>/dev/null; then
            PROCESS_NAME=$(ps -p $PID -o comm= 2>/dev/null || echo "unknown")
            echo -e "${YELLOW}🔪 杀死进程: $PID ($PROCESS_NAME)${NC}"
            kill -TERM $PID 2>/dev/null || kill -9 $PID 2>/dev/null || true
        fi
    done

    # 等待进程完全退出
    sleep 1

    # 再次检查
    REMAINING_PIDS=$(lsof -ti :$PORT 2>/dev/null || true)
    if [ -n "$REMAINING_PIDS" ]; then
        echo -e "${RED}⚠️  仍有进程占用端口，强制杀死...${NC}"
        for PID in $REMAINING_PIDS; do
            kill -9 $PID 2>/dev/null || true
        done
    fi
else
    echo -e "${GREEN}✅ 端口 $PORT 没有被占用${NC}"
fi

# 2. 杀死可能的孤儿 main 进程
echo -e "${YELLOW}🔍 检查孤儿 main 进程...${NC}"
MAIN_PIDS=$(pgrep -f "tmp/main" 2>/dev/null || true)

if [ -n "$MAIN_PIDS" ]; then
    echo -e "${YELLOW}⚠️  发现孤儿 main 进程: $MAIN_PIDS${NC}"
    for PID in $MAIN_PIDS; do
        if kill -0 $PID 2>/dev/null; then
            echo -e "${YELLOW}🔪 杀死孤儿进程: $PID${NC}"
            kill -TERM $PID 2>/dev/null || kill -9 $PID 2>/dev/null || true
        fi
    done
else
    echo -e "${GREEN}✅ 没有发现孤儿 main 进程${NC}"
fi

# 3. 清理临时文件
echo -e "${YELLOW}🗑️  清理临时文件...${NC}"
if [ -d "tmp" ]; then
    rm -rf tmp/*
    echo -e "${GREEN}✅ 清理 tmp/ 目录${NC}"
fi

# 清理日志文件
if [ -f "tmp/air.log" ]; then
    rm -f tmp/air.log
    echo -e "${GREEN}✅ 清理 air.log${NC}"
fi

# 4. 检查 Air 进程
echo -e "${YELLOW}🔍 检查 Air 进程...${NC}"
AIR_PIDS=$(pgrep -f "air" 2>/dev/null || true)

if [ -n "$AIR_PIDS" ]; then
    echo -e "${YELLOW}⚠️  发现 Air 进程: $AIR_PIDS${NC}"
    for PID in $AIR_PIDS; do
        if kill -0 $PID 2>/dev/null; then
            echo -e "${YELLOW}🔪 杀死 Air 进程: $PID${NC}"
            kill -TERM $PID 2>/dev/null || true
        fi
    done
else
    echo -e "${GREEN}✅ 没有发现残留的 Air 进程${NC}"
fi

# 5. 最终检查
sleep 1
echo -e "${BLUE}🔍 最终检查...${NC}"
FINAL_CHECK=$(lsof -ti :$PORT 2>/dev/null || true)

if [ -n "$FINAL_CHECK" ]; then
    echo -e "${RED}❌ 端口 $PORT 仍被占用: $FINAL_CHECK${NC}"
    exit 1
else
    echo -e "${GREEN}✅ 端口 $PORT 已释放${NC}"
fi

echo -e "${GREEN}🎉 清理完成！现在可以安全启动开发服务器${NC}"

# 6. 可选：显示端口状态
if command -v netstat >/dev/null 2>&1; then
    echo -e "${BLUE}📊 当前端口使用情况:${NC}"
    netstat -an | grep ":$PORT " || echo "端口 $PORT 未被使用"
fi