# Go Framework

这是一个基于 Go 语言和 Gin 框架的 Web 应用框架，提供完整的项目结构和常用功能。

## 功能特性

- 完整的项目结构和分层架构
- 统一的错误处理和响应格式
- 请求日志记录
- 基于JWT的用户认证
- 基于角色的访问控制
- 参数验证
- HTML模板渲染
- 静态资源服务
- RESTful API
- 安全中间件
- 限流控制
- 支持优雅关闭
- 开发模式下代码热重载

## 项目结构

```
├── bin                 # 可执行文件
│   └── main.go         # 主函数
├── config              # 配置文件
│   └── config.yaml     # 应用配置
├── internal            # 内部包
│   ├── controller      # 控制器层
│   ├── model           # 数据模型层
│   ├── repository      # 数据访问层
│   ├── router          # 路由配置
│   └── service         # 业务逻辑层
├── pkg                 # 公共包
│   ├── config          # 配置加载
│   ├── database        # 数据库连接
│   ├── errors          # 错误处理
│   ├── logger          # 日志处理
│   ├── middleware      # 中间件
│   └── response        # 响应处理
├── static              # 静态资源
│   ├── css             # 样式表
│   └── js              # JavaScript脚本
├── templates           # HTML模板
├── .air.toml           # Air热重载配置
├── Makefile            # 构建和开发命令
├── dev.sh              # 开发模式启动脚本
├── go.mod              # Go模块文件
├── go.sum              # 依赖校验和
└── README.md           # 项目说明
```

## 快速开始

### 环境要求

- Go 1.16+
- MySQL 5.7+

### 安装和运行

1. 克隆项目

```bash
git clone https://github.com/your-username/go-framework.git
cd go-framework
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置数据库

修改 `config/config.yaml` 中的数据库配置信息

4. 运行应用

```bash
# 生产模式运行
go run cmd/main.go

# 开发模式运行（支持热重载）
make dev
# 或
./dev.sh
```

### 热重载开发

项目使用 [Air](https://github.com/air-verse/air) 工具提供开发时的热重载功能，在修改代码后自动重新编译和运行应用。

支持监控的文件类型：
- Go源代码文件（.go）
- HTML模板文件（.html, .tpl, .tmpl）
- 配置文件（.yaml, .yml, .json, .toml）

使用方式：

```bash
# 使用make命令
make dev

# 或直接运行脚本
./dev.sh
```

> 提示：首次运行dev.sh时，如果系统中未安装Air，脚本会自动安装。

## API文档

### 用户相关接口

#### 注册用户

```
POST /api/v1/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "123456",
  "email": "test@example.com",
  "nickname": "测试用户"
}
```

#### 用户登录

```
POST /api/v1/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "123456"
}
```

#### 获取用户资料

```
GET /api/v1/profile
Authorization: Bearer {token}
```

#### 更新用户资料

```
PUT /api/v1/profile
Authorization: Bearer {token}
Content-Type: application/json

{
  "nickname": "新昵称",
  "email": "newemail@example.com"
}
```

## 开发指南

### 添加新的控制器

1. 在 `internal/controller` 目录下创建新的控制器文件
2. 实现控制器方法
3. 在 `internal/router/router.go` 中注册新的路由，指向控制器方法

### 添加新的模型

1. 在 `internal/model` 目录下创建新的模型文件
2. 定义模型结构体和相关方法
3. 在 `cmd/main.go` 中添加模型的自动迁移

### 使用路由参数

项目支持类似Flask风格的路由参数定义，例如：

```go
// 使用正则表达式验证的路由参数
apiGroup.GET("/user/<id:\\d+>", userController.GetUser)

// 多个路由参数
apiGroup.GET("/product/<category:[a-zA-Z0-9-]+>/<id:\\d+>", productController.GetProduct)
```

## 可用的Makefile命令

```bash
make dev    # 以开发模式运行（带热重载）
make build  # 构建应用程序
make run    # 运行应用程序（不带热重载）
make clean  # 清理临时文件
make help   # 显示帮助信息
```

## 许可证

MIT License 