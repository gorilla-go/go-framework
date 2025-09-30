package middleware

import (
	stderrors "errors" // 重命名标准库errors
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/errors"
	"github.com/gorilla-go/go-framework/pkg/response"
)

// RateLimiter 限流器
type RateLimiter struct {
	rate       int           // 速率（每秒请求数）
	interval   time.Duration // 间隔
	capacity   int           // 容量
	tokens     int           // 当前令牌数
	lastToken  time.Time     // 上次生成令牌时间
	lastAccess time.Time     // 上次访问时间（用于清理）
	mu         sync.Mutex    // 互斥锁
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

	// 计算距离上次生成令牌过了多长时间
	elapsed := now.Sub(r.lastToken)

	// 计算应该增加多少令牌（避免浮点数运算）
	newTokens := int(elapsed.Nanoseconds() * int64(r.rate) / r.interval.Nanoseconds())

	// 如果有新令牌生成
	if newTokens > 0 {
		r.tokens += newTokens
		// 使用更精确的时间计算
		r.lastToken = r.lastToken.Add(time.Duration(newTokens) * r.interval / time.Duration(r.rate))

		// 确保令牌数不超过容量
		if r.tokens > r.capacity {
			r.tokens = r.capacity
		}
	}

	// 如果有令牌，则消耗一个令牌
	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// IsExpired 检查限流器是否过期（超过指定时间未使用）
func (r *RateLimiter) IsExpired(ttl time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return time.Since(r.lastAccess) > ttl
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rate int, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Fail(c, errors.New(errors.TooManyRequests, "请求过于频繁，请稍后再试", stderrors.New("请求限流")))
			return
		}

		c.Next()
	}
}

// IPRateLimitMiddleware 基于IP的限流中间件
func IPRateLimitMiddleware(rate int, capacity int) gin.HandlerFunc {
	limiters := &sync.Map{} // 使用 sync.Map 减少锁竞争
	cleanupInterval := 10 * time.Minute
	ttl := 1 * time.Hour

	// 启动后台清理协程，定期清理过期的限流器
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			limiters.Range(func(key, value interface{}) bool {
				limiter := value.(*RateLimiter)
				if limiter.IsExpired(ttl) {
					limiters.Delete(key)
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 获取或创建限流器
		value, _ := limiters.LoadOrStore(ip, NewRateLimiter(rate, capacity))
		limiter := value.(*RateLimiter)

		if !limiter.Allow() {
			response.Fail(c, errors.New(errors.TooManyRequests, "请求过于频繁，请稍后再试", stderrors.New("IP请求限流")))
			return
		}

		c.Next()
	}
}
