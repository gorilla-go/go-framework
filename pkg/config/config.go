package config

import (
	"fmt"
	"os"

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
	Gzip     GzipConfig     `mapstructure:"gzip"`
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
	Path      string `mapstructure:"path"`
	Layouts   string `mapstructure:"layouts"`
	Extension string `mapstructure:"extension"`
}

// StaticConfig 静态文件配置
type StaticConfig struct {
	Path    string `mapstructure:"path"`
	SrcPath string `mapstructure:"src_path"`
}

// GzipConfig Gzip压缩配置
type GzipConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Level   int  `mapstructure:"level"`
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

var (
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
)

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认配置文件路径
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// 设置配置文件路径
	v.SetConfigFile(configPath)

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置到结构体
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}
