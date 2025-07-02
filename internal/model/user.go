package model

import "time"

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string    `json:"-" gorm:"size:100;not null"` // 密码不输出到JSON
	Email     string    `json:"email" gorm:"uniqueIndex;size:100"`
	Nickname  string    `json:"nickname" gorm:"size:50"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	Role      string    `json:"role" gorm:"size:20;default:'user'"` // 角色: admin, user
	Status    int       `json:"status" gorm:"default:1"`            // 状态: 0-禁用, 1-启用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"omitempty"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Nickname string `json:"nickname" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
	Avatar   string `json:"avatar" binding:"omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
