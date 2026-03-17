package middleware

import (
	stderrors "errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
)

// RateLimiter 令牌桶限流器
type RateLimiter struct {
	rate       int           // 速率（每秒请求数）
	interval   time.Duration // 间隔
	capacity   int           // 容量
	tokens     int           // 当前令牌数
	lastToken  time.Time     // 上次生成令牌时间
	lastAccess time.Time     // 上次访问时间（用于清理）
	mu         sync.Mutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate int, capacity int) *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		rate:       rate,
		interval:   time.Second,
		capacity:   capacity,
		tokens:     capacity,
		lastToken:  now,
		lastAccess: now,
	}
}

// Allow 是否允许请求
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.lastAccess = now

	elapsed := now.Sub(r.lastToken)
	newTokens := int(elapsed.Nanoseconds() * int64(r.rate) / r.interval.Nanoseconds())
	if newTokens > 0 {
		r.tokens += newTokens
		r.lastToken = r.lastToken.Add(time.Duration(newTokens) * r.interval / time.Duration(r.rate))
		if r.tokens > r.capacity {
			r.tokens = r.capacity
		}
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// IsExpired 检查限流器是否超过 ttl 未使用
func (r *RateLimiter) IsExpired(ttl time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return time.Since(r.lastAccess) > ttl
}

// ---- Functional Options（参考 Hertz 设计）----

// rateLimitConfig 限流中间件配置
type rateLimitConfig struct {
	rate    int
	burst   int
	skipper func(*gin.Context) bool // 返回 true 时跳过限流
}

// RateLimitOption 限流配置选项
type RateLimitOption func(*rateLimitConfig)

// WithRate 设置每秒允许的请求数（默认 100）
func WithRate(rate int) RateLimitOption {
	return func(c *rateLimitConfig) { c.rate = rate }
}

// WithBurst 设置突发容量（默认等于 rate）
func WithBurst(burst int) RateLimitOption {
	return func(c *rateLimitConfig) { c.burst = burst }
}

// WithSkipper 设置跳过函数，返回 true 时该请求不受限流约束
// 常用于跳过健康检查、内部路由等
func WithSkipper(fn func(*gin.Context) bool) RateLimitOption {
	return func(c *rateLimitConfig) { c.skipper = fn }
}

func newRateLimitConfig(opts []RateLimitOption) *rateLimitConfig {
	cfg := &rateLimitConfig{rate: 100}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.burst <= 0 {
		cfg.burst = cfg.rate
	}
	return cfg
}

// RateLimitMiddleware 全局限流中间件
//
// 用法（向后兼容旧签名，仍可通过 options 扩展）：
//
//	middleware.RateLimitMiddleware(middleware.WithRate(100), middleware.WithBurst(200))
//	middleware.RateLimitMiddleware(middleware.WithRate(50), middleware.WithSkipper(func(c *gin.Context) bool {
//	    return c.Request.URL.Path == "/health"
//	}))
func RateLimitMiddleware(opts ...RateLimitOption) gin.HandlerFunc {
	cfg := newRateLimitConfig(opts)
	limiter := NewRateLimiter(cfg.rate, cfg.burst)

	return func(c *gin.Context) {
		if cfg.skipper != nil && cfg.skipper(c) {
			c.Next()
			return
		}
		if !limiter.Allow() {
			response.Fail(c, errors.New(errors.TooManyRequests, "请求过于频繁，请稍后再试", stderrors.New("请求限流")))
			return
		}
		c.Next()
	}
}

// IPRateLimitMiddleware 基于客户端 IP 的限流中间件
func IPRateLimitMiddleware(opts ...RateLimitOption) gin.HandlerFunc {
	cfg := newRateLimitConfig(opts)
	limiters := &sync.Map{}
	cleanupInterval := 10 * time.Minute
	ttl := 1 * time.Hour

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			limiters.Range(func(key, value any) bool {
				if value.(*RateLimiter).IsExpired(ttl) {
					limiters.Delete(key)
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		if cfg.skipper != nil && cfg.skipper(c) {
			c.Next()
			return
		}
		ip := c.ClientIP()
		value, _ := limiters.LoadOrStore(ip, NewRateLimiter(cfg.rate, cfg.burst))
		if !value.(*RateLimiter).Allow() {
			response.Fail(c, errors.New(errors.TooManyRequests, "请求过于频繁，请稍后再试", stderrors.New("IP请求限流")))
			return
		}
		c.Next()
	}
}
