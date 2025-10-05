package session

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	gormsession "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/database"
)

// Start 启动会话中间件
func Start(sessionConfig *config.SessionConfig, redisConfig *config.RedisConfig, dbConfig *config.DatabaseConfig) gin.HandlerFunc {
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
			panic(fmt.Sprintf("Redis 会话存储初始化失败: %v", err))
		}

	case "gorm":
		// 使用GORM数据库存储
		if dbConfig == nil {
			panic("GORM 会话存储初始化失败: 数据库配置为空")
		}

		// 初始化数据库连接
		gormDB, err := database.Init(dbConfig)
		if err != nil {
			panic(fmt.Sprintf("GORM 会话存储初始化失败: %v", err))
		}

		// NewStore 参数: db, expiredSessionCleanup, keyPairs
		// expiredSessionCleanup: 是否启用过期会话自动清理
		store = gormsession.NewStore(gormDB, true, []byte(sessionConfig.Secret))

	case "memory":
		// 使用内存存储
		store = memstore.NewStore([]byte(sessionConfig.Secret))

	default:
		// 默认使用Cookie存储
		store = cookie.NewStore([]byte(sessionConfig.Secret))
	}

	// 解析 SameSite
	sameSite := parseSameSite(sessionConfig.SameSite)
	secure := sessionConfig.Secure

	// 安全性检查：SameSite=None 必须配合 Secure=true
	if sameSite == http.SameSiteNoneMode && !secure {
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

// Get 获取会话
func Get(c *gin.Context) sessions.Session {
	return sessions.Default(c)
}

// Set 设置会话值
func Set(c *gin.Context, key string, value interface{}) error {
	session := Get(c)
	session.Set(key, value)
	if err := session.Save(); err != nil {
		return fmt.Errorf("保存会话失败: %w", err)
	}
	return nil
}

// GetValue 获取会话值
func GetValue(c *gin.Context, key string) interface{} {
	session := Get(c)
	return session.Get(key)
}

// Delete 删除会话值
func Delete(c *gin.Context, key string) error {
	session := Get(c)
	session.Delete(key)
	if err := session.Save(); err != nil {
		return fmt.Errorf("删除会话值后保存失败: %w", err)
	}
	return nil
}

// Clear 清除会话
func Clear(c *gin.Context) error {
	session := Get(c)
	session.Clear()
	if err := session.Save(); err != nil {
		return fmt.Errorf("清除会话失败: %w", err)
	}
	return nil
}

// SetFlash 设置一次性消息
func SetFlash(c *gin.Context, key string, value interface{}) error {
	session := Get(c)
	session.AddFlash(value, key)
	if err := session.Save(); err != nil {
		return fmt.Errorf("保存闪存消息失败: %w", err)
	}
	return nil
}

// GetFlash 获取一次性消息
func GetFlash(c *gin.Context, key string) (interface{}, error) {
	session := Get(c)
	flashes := session.Flashes(key)
	if err := session.Save(); err != nil {
		return nil, fmt.Errorf("读取闪存消息后保存会话失败: %w", err)
	}
	if len(flashes) > 0 {
		return flashes[0], nil
	}
	return nil, nil
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
