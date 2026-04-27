package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "authorization header is required", nil)
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.UnauthorizedResponse(c, "invalid authorization header", nil)
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenParts[1], s.config.JWT.Secret)
		if err != nil {
			utils.UnauthorizedResponse(c, "invalid token", err)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func (s *Server) roleMiddleware(allowedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != allowedRole {
			utils.ForbiddenResponse(c, "Forbidden", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) rateLimitMiddleware(req int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("rate_limit:%s:%s", c.ClientIP(), c.FullPath())

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		pipe := s.redis.RDB().Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		count := incr.Val()
		if count > int64(req) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
