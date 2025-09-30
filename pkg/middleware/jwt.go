package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla-go/go-framework/pkg/config"
	pkgErrors "github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
)

// Context keys for storing user information
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRole     = "role"
	ContextKeyClaims   = "claims"
)

// JWT相关错误
var (
	ErrConfigNotLoaded    = errors.New("JWT配置未加载")
	ErrInvalidToken       = errors.New("无效的令牌")
	ErrInvalidSignMethod  = errors.New("无效的签名算法")
	ErrMissingAuth        = errors.New("缺少Authorization头")
	ErrInvalidAuthFormat  = errors.New("无效的Authorization格式")
	ErrUserNotAuth        = errors.New("用户未认证")
	ErrInsufficientPerms  = errors.New("权限不足")
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username string, role string, cfg *config.JWTConfig) (string, error) {
	if cfg == nil {
		return "", ErrConfigNotLoaded
	}

	now := time.Now()
	expireTime := now.Add(time.Duration(cfg.Expire) * time.Hour)

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("令牌签名失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, cfg *config.JWTConfig) (*JWTClaims, error) {
	if cfg == nil {
		return nil, ErrConfigNotLoaded
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSignMethod, token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// JWTMiddleware JWT认证中间件
func JWTMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, pkgErrors.NewUnauthorized("未提供认证信息", ErrMissingAuth))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Fail(c, pkgErrors.NewUnauthorized("认证格式错误", ErrInvalidAuthFormat))
			return
		}

		claims, err := ParseToken(parts[1], cfg)
		if err != nil {
			response.Fail(c, pkgErrors.NewUnauthorized("无效的认证信息", err))
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// RoleMiddleware 角色验证中间件
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			response.Fail(c, pkgErrors.NewUnauthorized("未认证", ErrUserNotAuth))
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Fail(c, pkgErrors.NewUnauthorized("未认证", ErrUserNotAuth))
			return
		}

		hasRole := false
		for _, r := range roles {
			if r == userRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Fail(c, pkgErrors.NewForbidden("权限不足", ErrInsufficientPerms))
			return
		}

		c.Next()
	}
}

// GetClaimsFromContext 从 Gin 上下文中获取 JWT Claims
func GetClaimsFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*JWTClaims)
	return jwtClaims, ok
}

// GetUserIDFromContext 从 Gin 上下文中获取用户 ID
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}
