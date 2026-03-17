package controller

// DemoAuthController 演示 JWT 认证 + Group 级别中间件
//
// 覆盖的新特性：
//   - rb.Group(path, middleware...) —— 组级中间件，组内所有路由自动鉴权
//   - middleware.JWTMiddleware()    —— Bearer Token 解析，Claims 存入 Context
//   - middleware.RoleMiddleware()   —— 角色验证，可叠加在 JWT 中间件之后
//   - middleware.GenerateToken()   —— 生成 HS256 JWT Token
//   - middleware.GetClaimsFromContext() / GetUserIDFromContext() —— 从 Context 读用户信息
//
// 路由：
//   POST /demo/auth/login       登录，返回 JWT Token（不需要认证）
//   GET  /demo/auth/profile     查看当前用户信息（需要 JWT）
//   GET  /demo/auth/admin-only  管理员专属接口（需要 JWT + role=admin）

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/middleware"
	"github.com/gorilla-go/go-framework/pkg/response"
	"github.com/gorilla-go/go-framework/pkg/router"
	"go.uber.org/fx"
)

type DemoAuthController struct {
	fx.In
	Config *config.Config
}

func (d *DemoAuthController) Annotation(rb *router.RouteBuilder) {
	// 公开路由：登录接口不需要认证
	rb.POST("/demo/auth/login", response.H(d.Login), "demo@login")

	// ---- 演示 Group 级中间件 ----
	// 只需在 Group 处传入 JWTMiddleware，组内所有路由自动校验 Bearer Token
	protected := rb.Group("/demo/auth", middleware.JWTMiddleware(&d.Config.JWT))
	protected.GET("/profile", response.H(d.Profile), "demo@profile")

	// 管理员路由：叠加 RoleMiddleware，JWT 验证通过后再验证角色
	adminGroup := rb.Group("/demo/auth",
		middleware.JWTMiddleware(&d.Config.JWT),
		middleware.RoleMiddleware("admin"),
	)
	adminGroup.GET("/admin-only", response.H(d.AdminOnly), "demo@adminOnly")
}

// ---- 请求/响应结构 ----

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ---- Handlers ----

// Login POST /demo/auth/login
// 演示 JWT Token 生成。固定两个演示账号：
//
//	user / 123456  → role: user
//	admin / 123456 → role: admin
func (d *DemoAuthController) Login(c *gin.Context) error {
	var req loginRequest
	if !response.BindJSON(c, &req) {
		return nil
	}

	// 演示用账号（实际项目中应查询数据库）
	accounts := map[string]struct {
		password string
		role     string
		id       uint
	}{
		"user":  {"123456", "user", 1},
		"admin": {"123456", "admin", 2},
	}

	account, exists := accounts[req.Username]
	if !exists || account.password != req.Password {
		return errors.NewUnauthorized("用户名或密码错误", nil)
	}

	token, err := middleware.GenerateToken(account.id, req.Username, account.role, &d.Config.JWT)
	if err != nil {
		return errors.NewInternalServerError("Token 生成失败", err)
	}

	response.SuccessD(c, "登录成功", gin.H{
		"token":    token,
		"username": req.Username,
		"role":     account.role,
		"tip":      "将 Token 放入 Header: Authorization: Bearer <token>",
	})
	return nil
}

// Profile GET /demo/auth/profile
// 演示从 Context 读取 JWT 解析后的用户信息
func (d *DemoAuthController) Profile(c *gin.Context) error {
	claims, ok := middleware.GetClaimsFromContext(c)
	if !ok {
		return errors.NewUnauthorized("无法获取用户信息", nil)
	}

	response.Success(c, gin.H{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"expires":  claims.ExpiresAt,
	})
	return nil
}

// AdminOnly GET /demo/auth/admin-only
// 只有 role=admin 的 Token 能通过 RoleMiddleware 到达此处
func (d *DemoAuthController) AdminOnly(c *gin.Context) error {
	userID, _ := middleware.GetUserIDFromContext(c)

	response.Success(c, gin.H{
		"message": "欢迎，管理员！",
		"user_id": userID,
		"tip":     "只有 role=admin 的 Token 才能访问此接口",
	})
	return nil
}
