package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/router"
	"github.com/gorilla-go/go-framework/pkg/template"
	"go.uber.org/fx"
)

type IndexController struct {
	fx.In
}

func (i *IndexController) Annotation(rb *router.RouteBuilder) {
	// 使用带名称的GET路由
	rb.GET("/", i.Index, "index@index")
}

func (i *IndexController) Index(ctx *gin.Context) {
	data := gin.H{
		"Title": "Go Framework Wiki",
		"Features": []map[string]any{
			{
				"Name":        "路由系统",
				"Description": "基于Gin的高性能路由系统，支持路由组、命名路由和路由参数",
				"Example":     `rb.GET("/users/:id", controller.Get, "user@get")`,
			},
			{
				"Name":        "依赖注入",
				"Description": "基于Uber FX的依赖注入系统，自动管理组件生命周期",
				"Example": `type Controller struct {
  fx.In
  Config *config.Config
  Logger *logger.Logger
}`,
			},
			{
				"Name":        "中间件系统",
				"Description": "提供丰富的内置中间件，包括日志、CORS、限流、会话等",
				"Example": `router.Use(middleware.Logger())
router.Use(middleware.Recovery())`,
			},
			{
				"Name":        "模板引擎",
				"Description": "强大的模板系统，支持布局、片段、自定义函数等",
				"Example": `{{ define "content" }}
  <h1>{{ .Title }}</h1>
{{ end }}`,
			},
			{
				"Name":        "配置管理",
				"Description": "基于YAML的配置系统，支持环境变量覆盖",
				"Example": `config := config.Fetch()
port := config.Server.Port`,
			},
			{
				"Name":        "日志系统",
				"Description": "结构化日志系统，支持多种日志级别和输出格式",
				"Example":     `logger.Info("服务启动成功", logger.Field("port", 8080))`,
			},
			{
				"Name":        "数据库访问",
				"Description": "集成GORM，支持多种数据库和连接池",
				"Example": `db := database.GetDB()
db.Where("id = ?", id).First(&user)`,
			},
			{
				"Name":        "事件总线",
				"Description": "JavaScript风格的事件管理器，支持线程安全的事件注册、触发和移除，提供松耦合的事件驱动编程",
				"Example": `// 注册事件监听器
eventbus.On("user.login", func(args ...interface{}) {
    username := args[0].(string)
    fmt.Printf("用户登录: %s\n", username)
})

// 注册一次性监听器
eventbus.Once("app.start", func(args ...interface{}) {
    fmt.Println("应用启动完成")
})

// 触发事件
eventbus.Emit("user.login", "张三")

// 移除监听器
eventbus.Off("user.login")`,
			},
			{
				"Name":        "会话管理",
				"Description": "支持Cookie、Redis、GORM、Memory四种存储方式，提供完整的会话操作和Flash消息功能",
				"Example": `// 设置会话
session.Set(c, "user_id", 123)

// 获取会话
userID := session.GetValue(c, "user_id")

// Flash消息（一次性）
session.SetFlash(c, "success", "操作成功")
msg, _ := session.GetFlash(c, "success")`,
			},
			{
				"Name":        "Cookie操作",
				"Description": "简洁的Cookie读写工具，支持设置过期时间、路径、域名等选项",
				"Example": `// 设置Cookie（3600秒后过期）
cookie.Set(c, "token", "abc123", 3600)

// 获取Cookie
token := cookie.Get(c, "token")

// 删除Cookie
cookie.Delete(c, "token")`,
			},
			{
				"Name":        "统一响应",
				"Description": "标准化的API响应格式，统一错误处理和状态码映射",
				"Example": `// 成功响应
response.Success(c, data)

// 带消息的成功响应
response.SuccessD(c, "操作成功", data)

// 错误响应
response.Fail(c, errors.NewNotFound("未找到", nil))

// 重定向
response.Redirect(c, "/login")`,
			},
			{
				"Name":        "请求处理",
				"Description": "便捷的请求参数绑定和验证工具",
				"Example": `// 绑定JSON请求
var req CreateUserRequest
if err := request.BindJSON(c, &req); err != nil {
    return
}

// 自动验证和错误处理
// 支持各种请求格式`,
			},
		},
		"Examples": []gin.H{
			{
				"Title":       "创建一个REST API",
				"Description": "展示如何创建一个简单的REST API",
				"Code": `// 控制器定义
type UserController struct {
  fx.In
}

// 注册路由
func (u *UserController) Annotation(rb *router.RouteBuilder) {
  rb.GET("/api/users", u.List, "user@list")
  rb.GET("/api/users/:id", u.Get, "user@get")
  rb.POST("/api/users", u.Create, "user@create")
}

// 获取用户列表
func (u *UserController) List(ctx *gin.Context) {
  users := []map[string]interface{}{
    {"id": 1, "name": "张三"},
    {"id": 2, "name": "李四"},
  }
  response.Success(ctx, users)
}`,
			},
			{
				"Title":       "渲染HTML模板",
				"Description": "展示如何渲染HTML模板",
				"Code": `// 控制器方法
func (c *Controller) Show(ctx *gin.Context) {
  data := gin.H{
    "Title": "用户详情",
    "User": gin.H{
      "ID": 1,
      "Name": "张三",
      "Email": "zhangsan@example.com",
    },
  }
  template.RenderL(ctx.Writer, "user/show", data)
}

// 模板文件 (templates/user/show.html)
{{ define "content" }}
<div class="user-profile">
  <h1>{{ .Title }}</h1>
  <div class="user-info">
    <p>ID: {{ .User.ID }}</p>
    <p>姓名: {{ .User.Name }}</p>
    <p>邮箱: {{ .User.Email }}</p>
  </div>
</div>
{{ end }}`,
			},
			{
				"Title":       "事件驱动编程",
				"Description": "使用事件总线实现松耦合的模块间通信，支持异步事件处理",
				"Code": `// 在控制器中触发事件
func (c *UserController) Login(ctx *gin.Context) {
    // 登录逻辑...
    eventbus.Emit("user.login", user.ID, user.Username, ctx.ClientIP())
    response.Success(ctx, user)
}

// 在服务中监听事件
func init() {
    // 记录登录日志
    eventbus.On("user.login", func(args ...interface{}) {
        userID := args[0]
        username := args[1]
        ip := args[2]
        logger.Info("用户登录", "user_id", userID, "username", username, "ip", ip)
    })

    // 更新最后登录时间
    eventbus.On("user.login", func(args ...interface{}) {
        userID := args[0]
        userService.UpdateLastLoginTime(userID)
    })
}`,
			},
			{
				"Title":       "会话和Cookie使用",
				"Description": "完整的会话管理和Cookie操作示例",
				"Code": `// 登录时设置会话和Cookie
func (c *AuthController) Login(ctx *gin.Context) {
    // 验证用户...

    // 设置会话
    session.Set(ctx, "user_id", user.ID)
    session.Set(ctx, "username", user.Name)

    // 设置Cookie（7天有效期）
    cookie.Set(ctx, "remember_token", token, 7*24*3600)

    // 设置Flash消息
    session.SetFlash(ctx, "success", "登录成功")

    response.Redirect(ctx, "/dashboard")
}

// 登出时清除会话
func (c *AuthController) Logout(ctx *gin.Context) {
    session.Clear(ctx)
    cookie.Delete(ctx, "remember_token")
    response.Redirect(ctx, "/login")
}`,
			},
		},
		"TechStack": []map[string]string{
			{"Name": "Gin", "Description": "高性能HTTP Web框架"},
			{"Name": "Uber FX", "Description": "依赖注入和生命周期管理"},
			{"Name": "GORM", "Description": "强大的ORM库"},
			{"Name": "Viper", "Description": "灵活的配置管理"},
			{"Name": "Zap", "Description": "高性能结构化日志"},
			{"Name": "Air", "Description": "开发热重载工具"},
			{"Name": "Gin Sessions", "Description": "会话管理"},
			{"Name": "Gulp", "Description": "前端资源构建管道"},
		},
	}

	template.RenderL(ctx.Writer, "index", data)
}
