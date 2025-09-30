#!/bin/bash

# 从 config.yaml 读取端口号
# 用法: ./scripts/get-port.sh

CONFIG_FILE="config/config.yaml"
DEFAULT_PORT="8080"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "$DEFAULT_PORT"
    exit 0
fi

# 优先使用 yq (YAML 专用解析器)
if command -v yq >/dev/null 2>&1; then
    PORT=$(yq eval '.server.port' "$CONFIG_FILE" 2>/dev/null)
    if [ "$PORT" != "null" ] && [ -n "$PORT" ]; then
        echo "$PORT"
        exit 0
    fi
fi

# 备用方案1: 使用 Python (大多数系统都有)
if command -v python3 >/dev/null 2>&1; then
    PORT=$(python3 -c "
import yaml
try:
    with open('$CONFIG_FILE', 'r', encoding='utf-8') as f:
        config = yaml.safe_load(f)
        print(config.get('server', {}).get('port', '$DEFAULT_PORT'))
except:
    print('$DEFAULT_PORT')
" 2>/dev/null)
    if [ -n "$PORT" ]; then
        echo "$PORT"
        exit 0
    fi
fi

# 备用方案2: 使用 Ruby (macOS 自带)
if command -v ruby >/dev/null 2>&1; then
    PORT=$(ruby -ryaml -e "
begin
    config = YAML.load_file('$CONFIG_FILE')
    puts config['server']['port'] || '$DEFAULT_PORT'
rescue
    puts '$DEFAULT_PORT'
end
" 2>/dev/null)
    if [ -n "$PORT" ]; then
        echo "$PORT"
        exit 0
    fi
fi

# 备用方案3: 简单的 grep/awk (最后兜底)
PORT=$(grep -A 10 "^server:" "$CONFIG_FILE" | grep "^\s*port:" | awk '{print $2}' | head -1)

if [ -z "$PORT" ] || [ "$PORT" = "null" ]; then
    echo "$DEFAULT_PORT"
else
    echo "$PORT"
fi