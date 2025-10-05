package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/session"
	"go.uber.org/zap"
)

// SessionMiddleware 会话中间件
func SessionMiddleware(sessionConfig *config.SessionConfig, redisConfig *config.RedisConfig, logger *zap.Logger) gin.HandlerFunc {
	return session.Start(sessionConfig, redisConfig, logger)
}
