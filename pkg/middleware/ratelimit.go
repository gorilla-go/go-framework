package middleware

import (
	stderrors "errors" // 重命名标准库errors
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器
type RateLimiter struct {
	rate      int           // 速率（每秒请求数）
	interval  time.Duration // 间隔
	capacity  int           // 容量
	tokens    int           // 当前令牌数
	lastToken time.Time     // 上次生成令牌时间
	mu        sync.Mutex    // 互斥锁
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate int, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:      rate,
		interval:  time.Second,
		capacity:  capacity,
		tokens:    capacity,
		lastToken: time.Now(),
	}
}

// Allow 是否允许请求
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// 计算距离上次生成令牌过了多长时间
	elapsed := now.Sub(r.lastToken)

	// 计算应该增加多少令牌
	newTokens := int(float64(elapsed) / float64(r.interval) * float64(r.rate))

	// 如果有新令牌生成
	if newTokens > 0 {
		r.tokens += newTokens
		r.lastToken = now

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

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rate int, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			// 使用通用错误处理
			HandleTooManyRequests(c, "请求过于频繁，请稍后再试", stderrors.New("请求限流"))
			return
		}

		c.Next()
	}
}

// IPRateLimitMiddleware 基于IP的限流中间件
func IPRateLimitMiddleware(rate int, capacity int) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	mu := sync.Mutex{}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			limiter = NewRateLimiter(rate, capacity)
			limiters[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			// 使用通用错误处理
			HandleTooManyRequests(c, "请求过于频繁，请稍后再试", stderrors.New("IP请求限流"))
			return
		}

		c.Next()
	}
}
