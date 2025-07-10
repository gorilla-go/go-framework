# 第一阶段：使用Go镜像进行构建
FROM golang:1.24.1-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache make nodejs npm

# 设置工作目录
WORKDIR /app

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 第二阶段：使用Alpine作为运行环境
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/bin/app /app/bin/

# 复制必要的目录
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static/dist /app/static/dist
COPY --from=builder /app/config /app/config

# 创建日志目录
RUN mkdir -p /app/logs

# 启动应用
CMD ["/app/bin/app"] 