package controller

import (
	"go-framework/internal/model"
	"go-framework/internal/service"
	"go-framework/pkg/errors"
	"go-framework/pkg/logger"
	"go-framework/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	userService *service.UserService
}

// NewUserController 创建用户控制器
func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// Register 用户注册
func (c *UserController) Register(ctx *gin.Context) {
	var req model.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Errorf("注册参数绑定失败: %v", err)
		response.ValidationError(ctx, "无效的请求参数", err)
		return
	}

	user, err := c.userService.Register(req)
	if err != nil {
		logger.Errorf("注册失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, user)
}

// Login 用户登录
func (c *UserController) Login(ctx *gin.Context) {
	var req model.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Errorf("登录参数绑定失败: %v", err)
		response.ValidationError(ctx, "无效的请求参数", err)
		return
	}

	user, token, err := c.userService.Login(req)
	if err != nil {
		logger.Errorf("登录失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetProfile 获取用户资料
func (c *UserController) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		errObj := errors.NewUnauthorized("未认证", nil)
		response.Fail(ctx, errObj)
		return
	}

	user, err := c.userService.GetUserByID(userID.(uint))
	if err != nil {
		logger.Errorf("获取用户信息失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, user)
}

// UpdateProfile 更新用户资料
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		errObj := errors.NewUnauthorized("未认证", nil)
		response.Fail(ctx, errObj)
		return
	}

	var req model.UserUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Errorf("更新参数绑定失败: %v", err)
		response.ValidationError(ctx, "无效的请求参数", err)
		return
	}

	user, err := c.userService.UpdateUser(userID.(uint), req)
	if err != nil {
		logger.Errorf("更新用户信息失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, user)
}

// GetUser 根据ID获取用户
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Errorf("用户ID解析失败: %v", err)
		response.BadRequest(ctx, "无效的用户ID", err)
		return
	}

	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		logger.Errorf("获取用户信息失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, user)
}

// GetUsers 获取用户列表
func (c *UserController) GetUsers(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "10")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	users, total, err := c.userService.GetUsers(page, size)
	if err != nil {
		logger.Errorf("获取用户列表失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Errorf("用户ID解析失败: %v", err)
		response.BadRequest(ctx, "无效的用户ID", err)
		return
	}

	if err := c.userService.DeleteUser(uint(id)); err != nil {
		logger.Errorf("删除用户失败: %v", err)
		response.Fail(ctx, err)
		return
	}

	response.SuccessWithDetail(ctx, "用户已删除", nil)
}
