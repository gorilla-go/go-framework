# Go Web Framework

一个基于Gin和Uber FX的高性能、模块化Go Web框架，支持依赖注入、中间件、模板渲染、会话管理和静态资源处理等功能。

## Wiki页面

框架提供了一个内置的Wiki页面，展示框架的核心功能和使用示例。启动应用后，访问首页即可查看：

```bash
make dev
# 然后在浏览器中访问 http://localhost:8081
```

Wiki页面提供了以下内容：
- 框架主要功能概览和代码示例
- 常见使用场景的详细示例代码
- 快速入门指南

## 目录

- [框架特性](#框架特性)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [配置管理](#配置管理)
- [路由系统](#路由系统)
- [控制器](#控制器)
- [中间件](#中间件)
- [模板渲染](#模板渲染)
- [会话管理](#会话管理)
- [数据库访问](#数据库访问)
- [日志系统](#日志系统)
- [错误处理](#错误处理)
- [静态资源](#静态资源)
- [依赖注入](#依赖注入)
- [开发工具](#开发工具)
- [生产部署](#生产部署)
- [常见问题](#常见问题)
- [许可证](#许可证)

## 框架特性

- **高性能**：基于[Gin](https://github.com/gin-gonic/gin)框架，提供极高的HTTP请求处理性能
- **依赖注入**：集成[Uber FX](https://github.com/uber-go/fx)实现自动依赖注入和生命周期管理
- **热重载**：开发模式支持代码热重载，提高开发效率
- **模块化设计**：清晰的目录结构和模块划分，易于维护和扩展
- **中间件系统**：提供丰富的内置中间件，包括日志、CORS、GZIP压缩、JWT、限流、会话等
- **模板引擎**：支持Go模板引擎，并提供丰富的模板函数
- **静态资源处理**：集成Gulp工作流，支持CSS、JavaScript的压缩和打包
- **会话管理**：支持Cookie和Redis两种会话存储方式
- **统一响应**：标准化的API响应格式和错误处理
- **优雅关闭**：支持应用优雅启动和关闭，确保请求正确处理
- **安全机制**：内置多种安全相关中间件和工具函数
- **路由注解**：支持声明式路由定义，简化路由管理
- **数据库集成**：内置GORM支持，简化数据库操作
- **丰富的工具函数**：提供大量实用工具函数

## 项目结构

```
.
├── bin/                # 编译后的二进制文件
├── cmd/                # 命令行入口
│   └── main.go         # 主函数
├── config/             # 配置文件
│   └── config.yaml     # 应用配置
├── internal/           # 内部代码（不对外暴露）
│   ├── app/            # 应用启动和生命周期
│   ├── controller/     # 控制器
│   ├── di/             # 依赖注入
│   └── router/         # 路由设置
├── pkg/                # 可重用包（可对外暴露）
│   ├── config/         # 配置加载
│   ├── database/       # 数据库连接
│   ├── errors/         # 错误处理
│   ├── logger/         # 日志处理
│   ├── middleware/     # 中间件
│   ├── response/       # 响应处理
│   └── template/       # 模板函数
├── static/             # 静态资源
│   ├── dist/           # 编译后的静态资源
│   └── src/            # 源静态资源
├── templates/          # HTML模板
│   └── layouts/        # 布局模板
├── logs/               # 日志文件
├── tmp/                # 临时文件
├── .air.toml           # Air热重载配置
├── .gitignore          # Git忽略文件
├── go.mod              # Go模块文件
├── go.sum              # 依赖版本锁
├── Makefile            # 构建和开发命令
└── README.md           # 项目说明
```

## 快速开始

### 环境要求

- Go 1.16+
- Node.js 14+（用于静态资源处理）
- MySQL 5.7+（可选，如果使用数据库）
- Redis（可选，如果使用Redis存储会话）

### 安装和运行

1. 克隆项目

```bash
git clone https://github.com/your-username/go-framework.git
cd go-framework
```

2. 安装依赖

```bash
# 安装Go依赖
go mod tidy

# 安装Node.js依赖（用于静态资源处理）
make install-deps
```

3. 配置应用

根据需要修改 `config/config.yaml` 配置文件

4. 开发模式运行

```bash
make dev
```

5. 生产模式运行

```bash
# 构建应用
make build

# 前台运行
make start

# 或后台运行
make start-d

# 停止后台运行的应用
make stop
```

## 配置管理

框架提供了一个完整的配置管理系统，支持YAML格式配置文件，配置结构如下：

```yaml
# 服务器配置
server:
  port: 8081
  mode: debug  # debug, release, test
  read_timeout: 60
  write_timeout: 60
  idle_timeout: 60
  enable_rate_limit: true # 是否启用全局限流
  rate_limit: 100  # 每秒最大请求数
  rate_burst: 200  # 突发情况下允许的最大请求数

# 日志配置
log:
  level: debug  # debug, info, warn, error, fatal, panic
  filename: logs/app.log
  max_size: 100  # MB
  max_backups: 10
  max_age: 30  # days
  compress: true
  format: json  # json, text

# 数据库配置
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: password
  dbname: test
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600  # seconds

# Redis配置
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

# JWT配置
jwt:
  secret: "your-secret-key"
  expire: 24  # hours
  issuer: "go-framework"

# 模板配置
template:
  path: ./templates
  layouts: ./templates/layouts
  extension: .html

# 静态文件配置
static:
  path: ./static/dist

# Gzip压缩配置
gzip:
  enabled: true
  level: 6  # 1-9，1最快，9压缩比最高

# 会话配置
session:
  store: cookie  # cookie, redis
  name: go_session
  secret: "session-secret-key"
  max_age: 60  # 分钟
  secure: false
  http_only: true
  path: /
  domain: ""
  same_site: lax  # lax, strict, none
```

框架提供了配置管理工具，可以在代码中方便地获取和使用配置项，支持自动加载和热更新配置。

## 路由系统

框架的路由系统基于Gin，并进行了扩展，提供以下功能：

- **路由注解**：支持通过控制器方法注解定义路由
- **路由分组**：支持路由分组，便于管理
- **中间件绑定**：可以为路由或路由组绑定中间件
- **命名路由**：支持命名路由，方便在模板中生成URL
- **请求参数绑定**：自动绑定请求参数到结构体
- **自定义验证**：支持请求参数的自定义验证规则
- **路由前缀**：支持设置全局路由前缀
- **路由重定向**：支持路由重定向功能
- **静态文件服务**：集成静态文件服务功能

## 控制器

框架支持结构化的控制器设计，主要特性：

- **依赖注入**：控制器可以通过依赖注入获取服务
- **路由注解**：支持在控制器方法上定义路由
- **参数绑定**：自动绑定请求参数到结构体
- **参数验证**：支持请求参数验证
- **统一响应**：标准化的API响应格式
- **错误处理**：集成的错误处理机制

## 中间件

框架提供了丰富的内置中间件，所有中间件都可配置并可按需加载：

- **Logger**：记录请求日志，包括请求时间、路径、方法、状态码、响应时间等
- **Recovery**：从panic中恢复，确保应用不会崩溃
- **CORS**：跨域资源共享，支持自定义配置
- **GZIP**：响应内容压缩，减少传输大小
- **RateLimit**：请求限流，支持全局限流和基于IP的限流
- **JWT**：JWT认证，支持自定义配置
- **Session**：会话管理，支持Cookie和Redis存储
- **Security**：安全相关头部，如XSS防护、点击劫持防护等
- **模板上下文**：为模板渲染提供全局上下文数据
- **缓存控制**：设置缓存控制头部
- **静态资源**：提供静态资源服务
- **自定义中间件**：支持自定义中间件的开发和集成

中间件可以注册为全局中间件，也可以应用于特定路由或路由组。

## 模板渲染

框架使用Go标准库的html/template包并进行了扩展，提供了强大的模板渲染功能：

- **布局系统**：支持模板布局，便于页面结构复用
- **模板缓存**：提高模板渲染性能
- **自定义函数**：提供100多个实用模板函数
- **部分模板渲染**：支持渲染部分模板
- **多模板渲染**：支持一次渲染多个模板
- **模板块**：支持在模板中定义和渲染块
- **开发模式**：开发模式下实时加载模板，无需重启服务
- **错误显示**：在开发环境下显示模板错误
- **上下文数据**：支持向模板传递上下文数据

框架内置的模板函数包括：

- **字符串处理**：trim、lower、upper、title、replace、split、join、contains等
- **数值处理**：add、subtract、multiply、divide、mod、round等
- **日期时间处理**：now、formatTime、dateFormat、humanizeTime等
- **集合处理**：first、last、empty、notEmpty、length、inArray等
- **条件处理**：default、ternary、eq、ne、lt、lte、gt、gte等
- **安全处理**：safeHTML、safeJS、safeCSS、safeURL等
- **URL处理**：url（路由生成）等
- **块处理**：render（动态渲染块）等
- **Map处理**：map、mapGet、mapHas、mapKeys、mapSet等

## 会话管理

框架提供了会话管理功能，支持多种存储后端：

- **Cookie存储**：数据存储在客户端Cookie中，适合小型应用
- **Redis存储**：数据存储在Redis中，适合大型应用和集群环境
- **会话操作**：提供SetSession、GetSession、DeleteSession、ClearSession等操作函数
- **会话配置**：支持配置会话名称、密钥、过期时间、Cookie属性等

## 数据库访问

框架集成了GORM作为ORM库，提供了以下功能：

- **数据库连接池**：支持配置连接池大小、最大空闲连接、最大连接生命周期等
- **事务支持**：支持数据库事务操作
- **模型定义**：支持定义数据库模型和关系
- **查询构建器**：提供流式API构建SQL查询
- **钩子函数**：支持模型的生命周期钩子函数
- **自动迁移**：支持自动创建/更新数据库表结构
- **多数据库支持**：支持MySQL、PostgreSQL、SQLite、SQL Server等

## 日志系统

框架提供了结构化的日志系统，主要特性：

- **日志级别**：支持Debug、Info、Warn、Error、Fatal、Panic等多个级别
- **日志格式**：支持JSON和文本两种格式
- **文件日志**：支持输出日志到文件，支持日志滚动
- **日志字段**：支持结构化日志字段，便于日志分析
- **日志钩子**：支持自定义日志钩子，如发送日志到远程服务
- **上下文集成**：支持从请求上下文中提取信息到日志
- **日志配置**：支持配置日志级别、输出位置、文件大小、备份数量等

## 错误处理

框架提供了统一的错误处理机制：

- **错误类型**：定义了AppError类型，包含错误码、错误信息、详细信息等
- **错误码**：预定义了常用的错误码，如BadRequest、Unauthorized、NotFound等
- **错误构建函数**：提供了创建各类错误的便捷函数，如NewBadRequest、NewNotFound等
- **错误中间件**：通过中间件统一处理错误，将错误转换为标准响应格式
- **错误链**：支持错误嵌套，保留原始错误信息
- **HTTP状态码映射**：自动将应用错误码映射到适当的HTTP状态码

## 静态资源

框架提供了静态资源处理功能：

- **静态文件服务**：支持提供静态文件服务，默认路径为/static
- **资源编译**：使用Gulp进行CSS、JavaScript的编译、压缩和打包
- **资源版本控制**：支持为资源添加版本号，便于缓存控制
- **开发模式**：开发模式下支持静态资源的热重载
- **GZIP压缩**：支持静态资源的GZIP压缩，减少传输大小
- **缓存控制**：支持设置静态资源的缓存控制头部

## 依赖注入

框架集成了Uber FX作为依赖注入和生命周期管理工具：

- **自动注入**：支持构造函数注入依赖
- **生命周期钩子**：支持应用启动和关闭时的生命周期钩子
- **模块化**：支持将应用分解为多个模块，每个模块可以独立提供和使用依赖
- **依赖分组**：支持将依赖分组，简化依赖管理
- **延迟初始化**：支持依赖的延迟初始化
- **依赖图**：构建清晰的依赖图，避免循环依赖
- **单例模式**：默认所有注入的依赖都是单例的

## 开发工具

框架提供了一系列开发工具，提高开发效率：

- **热重载**：使用Air实现代码热重载，无需重启服务
- **Makefile**：提供了常用开发和部署命令
- **调试模式**：支持开发环境的调试模式
- **错误页面**：在开发环境下显示友好的错误页面
- **模板重载**：开发模式下实时加载模板变更
- **静态资源监视**：开发模式下监视静态资源变更
- **性能分析**：支持生成性能分析报告
- **测试工具**：提供便捷的测试工具和函数

## 生产部署

框架支持多种部署方式：

- **二进制部署**：支持编译为单一二进制文件部署
- **Docker部署**：提供Dockerfile，支持容器化部署
- **Systemd服务**：支持作为Systemd服务运行
- **反向代理**：支持在Nginx、Caddy等反向代理后运行
- **优雅关闭**：支持应用的优雅启动和关闭
- **环境变量**：支持通过环境变量覆盖配置
- **监控集成**：支持集成Prometheus等监控工具
- **日志集成**：支持集成ELK等日志分析工具

## 常见问题

- **配置问题**：如何根据环境加载不同的配置
- **部署问题**：如何在生产环境部署应用
- **性能问题**：如何优化应用性能
- **安全问题**：如何增强应用安全性
- **扩展问题**：如何扩展框架功能

## 许可证

本项目采用 MIT 许可证，详情见 [LICENSE](./LICENSE) 文件。

```
MIT License

Copyright (c) 2023 Go Web Framework

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE. 