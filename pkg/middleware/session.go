package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla-go/go-framework/pkg/config"
	"github.com/gorilla-go/go-framework/pkg/session"
)

// SessionStart 会话中间件
func SessionStart(sessionConfig *config.SessionConfig, redisConfig *config.RedisConfig, dbConfig *config.DatabaseConfig) gin.HandlerFunc {
	return session.Start(sessionConfig, redisConfig, dbConfig)
}
