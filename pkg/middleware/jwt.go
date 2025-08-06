package middleware

import (
	"errors"
	"fmt"
	"go-framework/pkg/config"
	appErrors "go-framework/pkg/errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	// 获取配置
	if cfg == nil {
		return "", errors.New("配置未加载")
	}

	// 设置过期时间
	expireTime := time.Now().Add(time.Duration(cfg.Expire) * time.Hour)

	// 创建声明
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, cfg *config.JWTConfig) (*JWTClaims, error) {
	if cfg == nil {
		return nil, errors.New("配置未加载")
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名算法: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// JWTMiddleware JWT认证中间件
func JWTMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			HandleUnauthorized(c, "未提供认证信息", errors.New("缺少Authorization头"))
			return
		}

		// 检查格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			HandleUnauthorized(c, "认证格式错误", errors.New("无效的Authorization格式"))
			return
		}

		// 解析令牌
		claims, err := ParseToken(parts[1], cfg)
		if err != nil {
			HandleUnauthorized(c, "无效的认证信息", err)
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RoleMiddleware 角色验证中间件
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		role, exists := c.Get("role")
		if !exists {
			HandleUnauthorized(c, "未认证", errors.New("用户未认证"))
			return
		}

		// 检查角色
		userRole := role.(string)
		hasRole := false
		for _, r := range roles {
			if r == userRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			// 使用通用错误处理
			appErr := appErrors.NewForbidden("权限不足", errors.New("用户没有所需角色"))
			HandleAppError(c, appErr)
			return
		}

		c.Next()
	}
}
