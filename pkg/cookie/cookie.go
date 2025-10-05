package cookie

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Options Cookie 配置选项
type Options struct {
	MaxAge   int           // 过期时间（秒），0表示会话cookie，-1表示删除
	Path     string        // Cookie 路径，默认为 "/"
	Domain   string        // Cookie 域名
	Secure   bool          // 是否仅HTTPS传输
	HttpOnly bool          // 是否禁止JavaScript访问
	SameSite http.SameSite // SameSite 策略
}

// DefaultOptions 默认配置
var DefaultOptions = Options{
	MaxAge:   0,
	Path:     "/",
	Domain:   "",
	Secure:   false,
	HttpOnly: true,
	SameSite: http.SameSiteLaxMode,
}

// Set 设置Cookie（使用默认配置）
func Set(c *gin.Context, name, value string) {
	SetWithOptions(c, name, value, DefaultOptions)
}

// SetWithExpire 设置带过期时间的Cookie（秒）
func SetWithExpire(c *gin.Context, name, value string, maxAge int) {
	opts := DefaultOptions
	opts.MaxAge = maxAge
	SetWithOptions(c, name, value, opts)
}

// SetWithDuration 设置带过期时间的Cookie（使用 time.Duration）
func SetWithDuration(c *gin.Context, name, value string, duration time.Duration) {
	opts := DefaultOptions
	opts.MaxAge = int(duration.Seconds())
	SetWithOptions(c, name, value, opts)
}

// SetWithOptions 设置Cookie（自定义配置）
func SetWithOptions(c *gin.Context, name, value string, opts Options) {
	c.SetSameSite(opts.SameSite)
	c.SetCookie(
		name,
		value,
		opts.MaxAge,
		opts.Path,
		opts.Domain,
		opts.Secure,
		opts.HttpOnly,
	)
}

// Get 获取Cookie值
func Get(c *gin.Context, name string) (string, error) {
	return c.Cookie(name)
}

// GetWithDefault 获取Cookie值，如果不存在则返回默认值
func GetWithDefault(c *gin.Context, name, defaultValue string) string {
	value, err := c.Cookie(name)
	if err != nil {
		return defaultValue
	}
	return value
}

// Has 检查Cookie是否存在
func Has(c *gin.Context, name string) bool {
	_, err := c.Cookie(name)
	return err == nil
}

// Delete 删除Cookie
func Delete(c *gin.Context, name string) {
	opts := DefaultOptions
	opts.MaxAge = -1
	SetWithOptions(c, name, "", opts)
}
