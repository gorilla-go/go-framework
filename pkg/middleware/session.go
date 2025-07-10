package middleware

import (
	"go-framework/pkg/config"
	"net/http"

	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// SessionMiddleware 会话中间件
func SessionMiddleware() gin.HandlerFunc {
	// 使用全局配置，避免重复加载
	cfg := config.GetConfig()
	if cfg == nil {
		panic("配置未初始化")
	}

	// 会话配置
	sessionConfig := cfg.Session

	// 创建存储
	var store sessions.Store
	var err error

	// 根据配置选择存储类型
	switch sessionConfig.Store {
	case "redis":
		// 使用全局Redis配置
		redisAddr := cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port)

		// redis.NewStore 参数: size, network, address, username, password, keyPairs
		store, err = redis.NewStore(10, "tcp", redisAddr, "", cfg.Redis.Password, []byte(sessionConfig.Secret))
		if err != nil {
			panic(err)
		}
	default:
		// 默认使用Cookie存储
		store = cookie.NewStore([]byte(sessionConfig.Secret))
	}

	// 设置Cookie选项
	store.Options(sessions.Options{
		Path:     sessionConfig.Path,
		Domain:   sessionConfig.Domain,
		MaxAge:   sessionConfig.MaxAge * 60, // 转换为秒
		Secure:   sessionConfig.Secure,
		HttpOnly: sessionConfig.HttpOnly,
		SameSite: parseSameSite(sessionConfig.SameSite),
	})

	return sessions.Sessions(sessionConfig.Name, store)
}

// GetSession 获取会话
func GetSession(c *gin.Context) sessions.Session {
	return sessions.Default(c)
}

// SetSession 设置会话值
func SetSession(c *gin.Context, key string, value interface{}) {
	session := GetSession(c)
	session.Set(key, value)
	session.Save()
}

// GetSessionValue 获取会话值
func GetSessionValue(c *gin.Context, key string) interface{} {
	session := GetSession(c)
	return session.Get(key)
}

// DeleteSession 删除会话值
func DeleteSession(c *gin.Context, key string) {
	session := GetSession(c)
	session.Delete(key)
	session.Save()
}

// ClearSession 清除会话
func ClearSession(c *gin.Context) {
	session := GetSession(c)
	session.Clear()
	session.Save()
}

// SetFlash 设置一次性消息
func SetFlash(c *gin.Context, key string, value interface{}) {
	session := GetSession(c)
	session.AddFlash(value, key)
	session.Save()
}

// GetFlash 获取一次性消息
func GetFlash(c *gin.Context, key string) interface{} {
	session := GetSession(c)
	flashes := session.Flashes(key)
	session.Save()
	if len(flashes) > 0 {
		return flashes[0]
	}
	return nil
}

// SessionAuthMiddleware 会话认证中间件
func SessionAuthMiddleware(redirectPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		user := session.Get("user")
		if user == nil {
			c.Redirect(302, redirectPath)
			c.Abort()
			return
		}
		c.Next()
	}
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
