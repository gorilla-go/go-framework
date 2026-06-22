package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Template TemplateConfig `mapstructure:"template"`
	Static   StaticConfig   `mapstructure:"static"`
	Session  SessionConfig  `mapstructure:"session"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	IdleTimeout     int    `mapstructure:"idle_timeout"`
	EnableRateLimit bool   `mapstructure:"enable_rate_limit"`
	RateLimit       int    `mapstructure:"rate_limit"` // 每秒请求数
	RateBurst       int    `mapstructure:"rate_burst"` // 突发请求数
	// 可信代理列表（IP 或 CIDR）。仅当请求的直接来源在此列表内时，
	// 才信任 X-Forwarded-For/X-Real-IP 解析真实客户端 IP，防止伪造头绕过 IP 限流。
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Format     string `mapstructure:"format"`
	Stdout     bool   `mapstructure:"stdout"` // 是否同时输出到控制台
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
	Issuer string `mapstructure:"issuer"`
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Path          string `mapstructure:"path"`
	LayoutDir     string `mapstructure:"layout_dir"`
	Extension     string `mapstructure:"extension"`
	DefaultLayout string `mapstructure:"default_layout"`
}

// StaticConfig 静态文件配置
type StaticConfig struct {
	Path string `mapstructure:"path"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	// 存储类型: cookie, redis
	Store string `mapstructure:"store"`
	// 会话名称
	Name string `mapstructure:"name"`
	// 密钥
	Secret string `mapstructure:"secret"`
	// 过期时间（分钟）
	MaxAge int `mapstructure:"max_age"`
	// 是否只在HTTPS下发送Cookie
	Secure bool `mapstructure:"secure"`
	// 是否禁止JavaScript访问Cookie
	HttpOnly bool `mapstructure:"http_only"`
	// Cookie路径
	Path string `mapstructure:"path"`
	// Cookie域
	Domain string `mapstructure:"domain"`
	// SameSite策略
	SameSite string `mapstructure:"same_site"`
}

const defaultCfg = "config/config.yaml"

var (
	globalConfig *Config
	configOnce   sync.Once
	configErr    error
)

// Fetch 加载全局配置（进程内只加载一次）
func Fetch() (*Config, error) {
	configOnce.Do(func() {
		globalConfig, configErr = load(defaultCfg)
	})
	return globalConfig, configErr
}

// load 从指定路径加载配置，应用默认值并支持环境变量覆盖。
// 优先级：环境变量 > 配置文件 > 默认值。
func load(path string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", path)
	}

	v := viper.New()

	// 1. 先注册默认值，保证配置文件缺字段时不会退化为零值
	setDefaults(v)

	// 2. 环境变量覆盖：将 key 中的 "." 映射为 "_"（如 server.port → SERVER_PORT）
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	// 对所有已知 key 显式 BindEnv，确保 Unmarshal 时嵌套 key 的环境变量覆盖可靠生效
	// （仅靠 AutomaticEnv 对嵌套结构的 Unmarshal 并不可靠）
	for _, key := range v.AllKeys() {
		_ = v.BindEnv(key)
	}

	// 3. 读取配置文件（覆盖默认值）
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 4. 解析到结构体
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config, nil
}

// setDefaults 为所有配置项注册默认值。
// 这些默认值同时承担两个作用：配置文件缺字段时的兜底，以及向 viper 注册 key 供 BindEnv 使用。
func setDefaults(v *viper.Viper) {
	// server
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", 60)
	v.SetDefault("server.write_timeout", 60)
	v.SetDefault("server.idle_timeout", 60)
	v.SetDefault("server.enable_rate_limit", false)
	v.SetDefault("server.rate_limit", 100)
	v.SetDefault("server.rate_burst", 200)
	// 默认仅信任本机回环代理（同机反向代理场景），外部直连无法伪造转发头
	v.SetDefault("server.trusted_proxies", []string{"127.0.0.1", "::1"})

	// log
	v.SetDefault("log.level", "info")
	v.SetDefault("log.filename", "logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)
	v.SetDefault("log.format", "json")
	v.SetDefault("log.stdout", false)

	// database
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.username", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", 3600)

	// redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)

	// jwt
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.expire", 24)
	v.SetDefault("jwt.issuer", "go-framework")

	// template
	v.SetDefault("template.path", "templates")
	v.SetDefault("template.layout_dir", "layouts")
	v.SetDefault("template.extension", "html")
	v.SetDefault("template.default_layout", "main")

	// static
	v.SetDefault("static.path", "./static/dist")

	// session
	v.SetDefault("session.store", "cookie")
	v.SetDefault("session.name", "go_session")
	v.SetDefault("session.secret", "")
	v.SetDefault("session.max_age", 60)
	v.SetDefault("session.secure", false)
	v.SetDefault("session.http_only", true)
	v.SetDefault("session.path", "/")
	v.SetDefault("session.domain", "")
	v.SetDefault("session.same_site", "lax")
}

func MustFetch() *Config {
	config, err := Fetch()
	if err != nil {
		panic(err)
	}
	return config
}

func (c *Config) IsDebug() bool {
	return c.Server.Mode == "debug"
}
