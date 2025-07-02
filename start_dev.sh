#!/bin/bash

# 确保air命令存在
AIR_PATH=~/go/bin/air
if [ ! -f "$AIR_PATH" ]; then
    echo "Air未安装，正在安装..."
    go install github.com/air-verse/air@latest
fi

# 设置开发环境变量
export GIN_MODE=debug

# 使用air运行程序
echo "启动开发环境，监控文件变更..."
$AIR_PATH 