<div align="center">

# Go Framework

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Gin Version](https://img.shields.io/badge/Gin-v1.10-00ADD8?style=flat)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/gorilla-go/go-framework/pulls)

**一个现代化、高性能的 Go Web 框架，基于 Gin 和 Uber FX 构建**

[特性](#-特性) • [快速开始](#-快速开始) • [文档](#-文档) • [项目结构](#-项目结构) • [贡献](#-贡献)

</div>

---

## ✨ 特性

### 🚀 核心能力

- **高性能路由** - 基于 [Gin](https://github.com/gin-gonic/gin) 框架，提供极速的 HTTP 请求处理
- **依赖注入** - 集成 [Uber FX](https://github.com/uber-go/fx)，实现自动依赖注入和生命周期管理
- **热重载开发** - 使用 [Air](https://github.com/air-verse/air) 支持代码热重载，无需手动重启
- **优雅启停** - 支持优雅关闭，确保请求正确处理完毕

### 🛠️ 开发体验

- **模块化设计** - 清晰的目录结构，易于维护和扩展
- **丰富中间件** - 内置日志、CORS、GZIP、JWT、限流、会话等常用中间件
- **事件总线** - JavaScript 风格的事件系统，支持 on/once/off/emit
- **模板引擎** - 内置 100+ 实用模板函数，支持布局系统
- **配置管理** - 基于 Viper，支持 YAML 配置和环境变量覆盖
- **结构化日志** - 集成 Zap，支持日志分级、轮转和结构化输出

### 🔧 工具链

- **资源管道** - 集成 Gulp，支持 CSS/JS 压缩和打包
- **数据库 ORM** - 内置 GORM 支持，简化数据库操作
- **会话管理** - 支持 Cookie 和 Redis 存储
- **统一响应** - 标准化的 API 响应格式和错误处理
- **安全防护** - 内置多种安全中间件（XSS、CSRF、安全头等）

---

## 📦 快速开始

### 环境要求

| 工具    | 版本要求 | 必需                    |
| ------- | -------- | ----------------------- |
| Go      | 1.24+    | ✅                      |
| Node.js | 14+      | ✅ (用于资源构建)       |
| MySQL   | 5.7+     | ⭕ (可选)               |
| Redis   | 任意版本 | ⭕ (可选，用于会话存储) |

### 安装

```bash
# 1. 克隆项目
git clone https://github.com/gorilla-go/go-framework.git
cd go-framework

# 2. 安装 Go 依赖
go mod tidy

# 3. 安装 Node.js 依赖（用于静态资源处理）
make install

# 4. 复制并配置文件（可选）
cp config/config.yaml.example config/config.yaml
# 根据需要修改配置文件
```

### 运行

```bash
# 开发模式（支持热重载）
make dev

# 生产模式
make build    # 构建二进制文件
make start    # 前台运行
make startd  # 后台运行
make stop     # 停止后台服务
```

访问 `http://localhost:8081` 查看应用。

### Hello World 示例

```go
// app/controller/hello.go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla-go/go-framework/pkg/response"
)

type HelloController struct{}

func NewHelloController() *HelloController {
    return &HelloController{}
}

func (h *HelloController) Register(r *gin.Engine) {
    r.GET("/hello", h.Hello)
}

func (h *HelloController) Hello(c *gin.Context) {
    response.Success(c, gin.H{
        "message": "Hello, Go Framework!",
    })
}
```

---

## 📚 文档

### 核心概念

#### 依赖注入

框架使用 Uber FX 实现依赖注入，所有服务都在 `bootstrap/provide.go` 中注册：

```go
// bootstrap/provide.go
func Database() *gorm.DB {
    cfg, _ := config.Fetch()
    db, _ := database.NewDatabase(&cfg.Database)
    return db
}

// 在控制器中使用
type UserController struct {
    db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
    return &UserController{db: db}
}
```

#### 事件系统

内置 JavaScript 风格的事件总线：

```go
import "github.com/gorilla-go/go-framework/pkg/eventbus"

// 注册事件
eventbus.On("user:created", func(data interface{}) {
    user := data.(User)
    fmt.Println("New user:", user.Name)
})

// 触发事件
eventbus.Emit("user:created", User{Name: "John"})

// 一次性监听
eventbus.Once("app:ready", func(data interface{}) {
    fmt.Println("App is ready!")
})

// 移除监听
eventbus.Off("user:created")
```

#### 中间件使用

```go
// routes/routes.go
func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
    // 全局中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())

    // 路由组中间件
    api := r.Group("/api")
    api.Use(middleware.RateLimit(100, 200)) // 限流: 100 req/s, burst 200
    {
        api.GET("/users", controller.GetUsers)
    }

    // 单个路由中间件
    r.POST("/admin", middleware.JWT(), controller.AdminAction)
}
```

#### 配置管理

配置文件 `config/config.yaml` 使用 Viper 加载，支持环境变量覆盖：

```yaml
server:
  port: 8081
  mode: debug # debug, release, test

database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: password
  dbname: myapp
```

环境变量覆盖示例：

```bash
export SERVER_PORT=8080
export DATABASE_HOST=192.168.1.100
```

#### 统一响应格式

```go
import "github.com/gorilla-go/go-framework/pkg/response"

// 成功响应
response.Success(c, gin.H{
    "users": users,
})

// 错误响应
response.Error(c, errors.NewNotFound("User not found"))

// 分页响应
response.Paginate(c, users, total, page, pageSize)
```

---

## 📂 项目结构

```
.
├── app/                    # 应用层
│   └── controller/         # 控制器
├── bootstrap/              # 应用启动和依赖注入
│   ├── app.go              # FX 应用配置
│   └── provide.go          # 依赖提供者
├── cmd/                    # 命令行入口
│   └── main.go             # 主函数
├── config/                 # 配置文件
│   └── config.yaml         # 应用配置
├── pkg/                    # 可重用包
│   ├── config/             # 配置加载
│   ├── database/           # 数据库连接
│   ├── eventbus/           # 事件总线
│   ├── errors/             # 错误定义
│   ├── logger/             # 日志系统
│   ├── middleware/         # 中间件
│   ├── response/           # 响应处理
│   ├── router/             # 路由构建器
│   └── template/           # 模板引擎
├── routes/                 # 路由注册
│   └── routes.go
├── static/                 # 静态资源
│   ├── dist/               # 构建产物
│   ├── src/                # 源文件
│   └── gulpfile.js         # Gulp 配置
├── templates/              # HTML 模板
│   ├── layouts/            # 布局模板
│   └── pages/              # 页面模板
├── logs/                   # 日志文件
├── tmp/                    # 临时文件（Air 使用）
├── .air.toml               # Air 热重载配置
├── Dockerfile              # Docker 镜像
├── Makefile                # 构建命令
└── go.mod                  # Go 模块定义
```

### 关键目录说明

| 目录              | 说明                                |
| ----------------- | ----------------------------------- |
| `app/controller/` | 业务控制器，实现 `IController` 接口 |
| `bootstrap/`      | 依赖注入配置和应用启动逻辑          |
| `pkg/`            | 框架核心包，可被其他项目复用        |
| `routes/`         | 路由注册和中间件配置                |
| `config/`         | YAML 配置文件                       |
| `templates/`      | HTML 模板（支持布局系统）           |
| `static/`         | 前端资源（经 Gulp 处理）            |

---

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./pkg/eventbus -v

# 运行基准测试
go test -bench=. ./pkg/template
```

---

## 🐳 部署

### Docker 部署

```bash
# 构建镜像
docker build -t go-framework:latest .

# 运行容器
docker run -d \
  -p 8081:8081 \
  -v $(pwd)/config:/app/config \
  --name go-framework \
  go-framework:latest
```

### 二进制部署

```bash
# 编译
make build

# 后台运行
make startd

# 查看状态
make status

# 停止服务
make stop
```

---

## 🔧 开发指南

### 添加新控制器

1. 在 `app/controller/` 下创建控制器文件
2. 实现 `IController` 接口：

   ```go
   type MyController struct{}

   func NewMyController() *MyController {
       return &MyController{}
   }

   func (m *MyController) Register(r *gin.Engine) {
       r.GET("/my-route", m.MyHandler)
   }
   ```

3. 在 `bootstrap/provide.go` 中注册控制器

### 创建自定义中间件

```go
// pkg/middleware/custom.go
package middleware

import "github.com/gin-gonic/gin"

func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        c.Set("custom_key", "value")

        c.Next()

        // 后置处理
    }
}
```

### Makefile 命令

| 命令              | 说明               |
| ----------------- | ------------------ |
| `make dev`        | 开发模式（热重载） |
| `make build`      | 构建生产二进制     |
| `make start`      | 前台运行           |
| `make startd`     | 后台运行           |
| `make stop`       | 停止后台服务       |
| `make install`    | 安装 Node.js 依赖  |
| `make gulp-build` | 构建静态资源       |
| `make clean`      | 清理临时文件       |

---

## 🤝 贡献

我们欢迎所有形式的贡献！

### 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 代码规范

- 遵循 Go 官方代码风格
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档
- 为新功能添加测试用例

---

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

```
MIT License

Copyright (c) 2025 Go Framework Contributors

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
```

---

## 🙏 致谢

本项目基于以下优秀的开源项目构建：

- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [Uber FX](https://github.com/uber-go/fx) - 依赖注入框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 结构化日志
- [GORM](https://gorm.io) - ORM 库
- [Air](https://github.com/air-verse/air) - 热重载工具

---

## 📮 联系方式

- 项目主页: [https://github.com/gorilla-go/go-framework](https://github.com/gorilla-go/go-framework)
- Issues: [https://github.com/gorilla-go/go-framework/issues](https://github.com/gorilla-go/go-framework/issues)
- Discussions: [https://github.com/gorilla-go/go-framework/discussions](https://github.com/gorilla-go/go-framework/discussions)

---

<div align="center">

**如果这个项目对你有帮助，请给我们一个 ⭐️ Star！**

Made with ❤️ by Go Framework Contributors

</div>
