package controller

// DemoAPIController 演示 REST API 相关功能
//
// 覆盖的新特性：
//   - response.H()        —— handler 直接 return error，告别 response.Fail+return 样板
//   - request.BindJSON() —— 一步完成 JSON 绑定 + 校验
//   - request.BindUri()  —— 一步完成路径参数绑定 + 校验
//   - request.BindQuery()—— 一步完成 Query 参数绑定 + 校验
//   - middleware.GetLogEntry().AddField() —— 在 handler 里追加字段到当前请求日志
//
// 路由：
//   GET    /demo/api/users       列表（支持 ?keyword= 过滤）
//   GET    /demo/api/users/:id   查询单个用户
//   POST   /demo/api/users       创建用户
//   DELETE /demo/api/users/:id   删除用户

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/middleware"
	"github.com/gorilla-go/go-framework/pkg/request"
	"github.com/gorilla-go/go-framework/pkg/response"
	"github.com/gorilla-go/go-framework/pkg/router"
	"go.uber.org/fx"
)

// ---- 演示用内存数据 ----

type demoUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

var (
	demoStore   sync.Map
	demoIDSeq   uint64 = 2
	demoStoreMu sync.Mutex
)

func init() {
	demoStore.Store(uint(1), &demoUser{ID: 1, Name: "张三", Email: "zhangsan@example.com", Role: "user"})
	demoStore.Store(uint(2), &demoUser{ID: 2, Name: "李四", Email: "lisi@example.com", Role: "admin"})
}

// ---- 控制器 ----

type DemoAPIController struct {
	fx.In
}

func (d *DemoAPIController) Annotation(rb *router.RouteBuilder) {
	api := rb.Group("/demo/api")

	api.GET("/users", d.ListUsers, "demo@listUsers")
	api.GET("/users/:id", d.GetUser, "demo@getUser")
	api.POST("/users", d.CreateUser, "demo@createUser")
	api.DELETE("/users/:id", d.DeleteUser, "demo@deleteUser")
}

// ---- ListUsers: 演示 BindQuery ----

type listUsersQuery struct {
	Keyword string `form:"keyword"` // ?keyword=张
	Role    string `form:"role"`    // ?role=admin
}

// ListUsers GET /demo/api/users
// 演示 request.BindQuery —— 绑定 Query 参数，无 keyword/role 时返回全部
func (d *DemoAPIController) ListUsers(c *gin.Context) error {
	var query listUsersQuery
	if err := request.BindQuery(c, &query); err != nil {
		return err
	}

	var result []*demoUser
	demoStore.Range(func(_, v any) bool {
		u := v.(*demoUser)
		if query.Keyword != "" && u.Name != query.Keyword {
			return true
		}
		if query.Role != "" && u.Role != query.Role {
			return true
		}
		result = append(result, u)
		return true
	})

	response.SuccessD(c, fmt.Sprintf("共 %d 条", len(result)), result)
	return nil
}

// ---- GetUser: 演示 H() + BindUri + LogEntry ----

type getUserUri struct {
	ID uint `uri:"id" binding:"required"`
}

// GetUser GET /demo/api/users/:id
// 演示 response.H() —— 方法签名改为 return error，框架自动转 HTTP 响应
// 演示 middleware.GetLogEntry().AddField() —— 追加字段到当前请求日志
func (d *DemoAPIController) GetUser(c *gin.Context) error {
	var uri getUserUri
	if err := request.BindUri(c, &uri); err != nil {
		return err
	}

	val, ok := demoStore.Load(uri.ID)
	if !ok {
		// 直接 return error，H() 会自动调用 Fail()
		return errors.NewNotFound(fmt.Sprintf("用户 %d 不存在", uri.ID), nil)
	}

	// 向当前请求日志追加业务字段，无需修改 Logger 中间件
	middleware.GetLogEntry(c).AddField("queried_user_id", uri.ID)

	response.Success(c, val)
	return nil
}

// ---- CreateUser: 演示 H() + BindJSON ----

type createUserRequest struct {
	Name  string `json:"name"  binding:"required,min=2,max=20"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"  binding:"omitempty,oneof=user admin"`
}

// CreateUser POST /demo/api/users
// 演示 request.BindJSON —— 绑定 JSON 请求体 + 校验（binding 标签）
func (d *DemoAPIController) CreateUser(c *gin.Context) error {
	var req createUserRequest
	if err := request.BindJSON(c, &req); err != nil {
		return err
	}

	if req.Role == "" {
		req.Role = "user"
	}

	newID := uint(atomic.AddUint64(&demoIDSeq, 1))
	user := &demoUser{
		ID:    newID,
		Name:  req.Name,
		Email: req.Email,
		Role:  req.Role,
	}
	demoStore.Store(newID, user)

	middleware.GetLogEntry(c).AddField("created_user_id", newID)

	response.SuccessD(c, "用户创建成功", user)
	return nil
}

// ---- DeleteUser: 演示 H() + BindUri ----

type deleteUserUri struct {
	ID uint `uri:"id" binding:"required"`
}

// DeleteUser DELETE /demo/api/users/:id
func (d *DemoAPIController) DeleteUser(c *gin.Context) error {
	var uri deleteUserUri
	if err := request.BindUri(c, &uri); err != nil {
		return err
	}

	if _, ok := demoStore.LoadAndDelete(uri.ID); !ok {
		return errors.NewNotFound(fmt.Sprintf("用户 %d 不存在", uri.ID), nil)
	}

	middleware.GetLogEntry(c).AddField("deleted_user_id", uri.ID)

	response.SuccessD(c, "删除成功", gin.H{"id": uri.ID})
	return nil
}
