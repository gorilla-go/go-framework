package service

import (
	"errors"
	"go-framework/internal/model"
	"go-framework/internal/repository"
	appErrors "go-framework/pkg/errors"
	"go-framework/pkg/middleware"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Register 注册用户
func (s *UserService) Register(req model.RegisterRequest) (*model.User, error) {
	// 检查用户名是否存在
	exists, err := s.repo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, appErrors.NewDatabaseError("检查用户名失败", err)
	}
	if exists {
		return nil, appErrors.NewBadRequest("用户名已存在", errors.New("用户名已被注册"))
	}

	// 检查邮箱是否存在
	exists, err = s.repo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, appErrors.NewDatabaseError("检查邮箱失败", err)
	}
	if exists {
		return nil, appErrors.NewBadRequest("邮箱已存在", errors.New("邮箱已被注册"))
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErrors.NewInternalServerError("密码加密失败", err)
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Nickname: req.Nickname,
		Role:     "user", // 默认角色
		Status:   1,      // 默认启用
	}

	// 保存用户
	if err := s.repo.Create(user); err != nil {
		return nil, appErrors.NewDatabaseError("创建用户失败", err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req model.LoginRequest) (*model.User, string, error) {
	// 根据用户名获取用户
	user, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		return nil, "", appErrors.NewNotFound("用户不存在", err)
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, "", appErrors.NewForbidden("账户已被禁用", errors.New("账户状态异常"))
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", appErrors.NewUnauthorized("密码错误", err)
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", appErrors.NewInternalServerError("生成令牌失败", err)
	}

	return user, token, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, appErrors.NewNotFound("用户不存在", err)
	}
	return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, req model.UserUpdateRequest) (*model.User, error) {
	// 获取用户
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, appErrors.NewNotFound("用户不存在", err)
	}

	// 更新用户信息
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" && user.Email != req.Email {
		// 检查邮箱是否被其他用户使用
		exists, err := s.repo.ExistsByEmailExceptUser(req.Email, id)
		if err != nil {
			return nil, appErrors.NewDatabaseError("检查邮箱失败", err)
		}
		if exists {
			return nil, appErrors.NewBadRequest("邮箱已存在", errors.New("邮箱已被其他用户注册"))
		}
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// 保存用户
	if err := s.repo.Update(user); err != nil {
		return nil, appErrors.NewDatabaseError("更新用户失败", err)
	}

	return user, nil
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(page, size int) ([]*model.User, int64, error) {
	users, total, err := s.repo.List(page, size)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError("获取用户列表失败", err)
	}
	return users, total, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return appErrors.NewDatabaseError("删除用户失败", err)
	}
	return nil
}
