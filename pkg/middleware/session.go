package middleware

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
	"go.uber.org/zap"
)

// SessionMiddleware 会话中间件
func SessionMiddleware(sessionConfig *config.SessionConfig, redisConfig *config.RedisConfig, logger *zap.Logger) gin.HandlerFunc {
	// 创建存储
	var store sessions.Store
	var err error

	// 根据配置选择存储类型
	switch sessionConfig.Store {
	case "redis":
		// 使用全局Redis配置
		redisAddr := redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port)

		// 动态设置连接池大小（默认 10，最小 5，最大 100）
		poolSize := 10
		if redisConfig.PoolSize > 0 {
			poolSize = redisConfig.PoolSize
			if poolSize < 5 {
				poolSize = 5
			} else if poolSize > 100 {
				poolSize = 100
			}
		}

		// redis.NewStore 参数: size, network, address, username, password, keyPairs
		store, err = redis.NewStore(poolSize, "tcp", redisAddr, "", redisConfig.Password, []byte(sessionConfig.Secret))
		if err != nil {
			logger.Error("创建 Redis 会话存储失败", zap.Error(err), zap.String("addr", redisAddr))
			panic(fmt.Sprintf("Redis 会话存储初始化失败: %v", err))
		}
		logger.Info("Redis 会话存储已初始化", zap.String("addr", redisAddr), zap.Int("poolSize", poolSize))
	default:
		// 默认使用Cookie存储
		store = cookie.NewStore([]byte(sessionConfig.Secret))
		logger.Info("Cookie 会话存储已初始化")
	}

	// 解析 SameSite
	sameSite := parseSameSite(sessionConfig.SameSite)
	secure := sessionConfig.Secure

	// 安全性检查：SameSite=None 必须配合 Secure=true
	if sameSite == http.SameSiteNoneMode && !secure {
		logger.Warn("SameSite=None 需要配合 Secure=true，已自动修正")
		secure = true
	}

	// 设置Cookie选项
	store.Options(sessions.Options{
		Path:     sessionConfig.Path,
		Domain:   sessionConfig.Domain,
		MaxAge:   sessionConfig.MaxAge * 60, // 转换为秒
		Secure:   secure,
		HttpOnly: sessionConfig.HttpOnly,
		SameSite: sameSite,
	})

	return sessions.Sessions(sessionConfig.Name, store)
}

// GetSession 获取会话
func GetSession(c *gin.Context) sessions.Session {
	return sessions.Default(c)
}

// SetSession 设置会话值
func SetSession(c *gin.Context, key string, value interface{}) error {
	session := GetSession(c)
	session.Set(key, value)
	if err := session.Save(); err != nil {
		return fmt.Errorf("保存会话失败: %w", err)
	}
	return nil
}

// GetSessionValue 获取会话值
func GetSessionValue(c *gin.Context, key string) interface{} {
	session := GetSession(c)
	return session.Get(key)
}

// DeleteSession 删除会话值
func DeleteSession(c *gin.Context, key string) error {
	session := GetSession(c)
	session.Delete(key)
	if err := session.Save(); err != nil {
		return fmt.Errorf("删除会话值后保存失败: %w", err)
	}
	return nil
}

// ClearSession 清除会话
func ClearSession(c *gin.Context) error {
	session := GetSession(c)
	session.Clear()
	if err := session.Save(); err != nil {
		return fmt.Errorf("清除会话失败: %w", err)
	}
	return nil
}

// SetFlash 设置一次性消息
func SetFlash(c *gin.Context, key string, value interface{}) error {
	session := GetSession(c)
	session.AddFlash(value, key)
	if err := session.Save(); err != nil {
		return fmt.Errorf("保存闪存消息失败: %w", err)
	}
	return nil
}

// GetFlash 获取一次性消息
func GetFlash(c *gin.Context, key string) (interface{}, error) {
	session := GetSession(c)
	flashes := session.Flashes(key)
	if err := session.Save(); err != nil {
		return nil, fmt.Errorf("读取闪存消息后保存会话失败: %w", err)
	}
	if len(flashes) > 0 {
		return flashes[0], nil
	}
	return nil, nil
}

// SessionAuthMiddleware 会话认证中间件
// userKey: 存储在 session 中的用户键名，默认为 "user"
// redirectPath: 未认证时的重定向路径（仅用于网页请求）
func SessionAuthMiddleware(userKey, redirectPath string) gin.HandlerFunc {
	if userKey == "" {
		userKey = "user"
	}

	return func(c *gin.Context) {
		session := GetSession(c)
		user := session.Get(userKey)

		if user == nil {
			// 判断是 API 请求还是网页请求
			if isAPIRequest(c) {
				// API 请求返回 JSON
				response.Fail(c, errors.New(errors.Unauthorized, "未登录或登录已过期", stderrors.New("会话认证失败")))
			} else {
				// 网页请求重定向到登录页
				c.Redirect(http.StatusFound, redirectPath)
			}
			c.Abort()
			return
		}

		// 将用户信息存入 context，方便后续使用
		c.Set(userKey, user)
		c.Next()
	}
}

// isAPIRequest 判断是否为 API 请求
func isAPIRequest(c *gin.Context) bool {
	// 检查 Accept 头
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	// 检查 Content-Type 头
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// 检查 X-Requested-With 头（AJAX 请求）
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		return true
	}

	// 检查路径前缀
	path := c.Request.URL.Path
	return strings.HasPrefix(path, "/api/")
}

// parseSameSite 解析SameSite策略
func parseSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
