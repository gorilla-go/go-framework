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
	// 准备框架功能数据

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
				"Example": `config := config.GetConfig()
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
				"Name":        "统一响应",
				"Description": "标准化API响应格式，统一错误处理",
				"Example": `response.Success(ctx, data)
response.Error(ctx, "错误信息", 400)`,
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
func (u *UserController) Annotation(rb *middleware.RouteBuilder) {
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
  data := map[string]interface{}{
    "Title": "用户详情",
    "User": map[string]interface{}{
      "ID": 1,
      "Name": "张三",
      "Email": "zhangsan@example.com",
    },
  }
  c.TemplateManager.RenderWithDefaultLayout(ctx.Writer, "user/show", data)
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
		},
	}

	// 使用模板引擎渲染模板
	template.RenderWithDefaultLayout(ctx.Writer, "index", data)
}
