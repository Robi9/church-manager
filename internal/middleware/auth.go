package middleware

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Robi9/church-manager/internal/modules/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type SessionValidator interface {
	ValidateSession(context.Context, string) (int64, error)
}

func AuthMiddleware(secret string, sessions SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || parts[0] != "Bearer" || len(parts[1]) != 43 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		digest := auth.SessionDigest(parts[1], secret)
		id, err := sessions.ValidateSession(c.Request.Context(), digest)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.AbortWithStatusJSON(401, gin.H{"error": "invalid session"})
			} else {
				c.AbortWithStatusJSON(503, gin.H{"error": "authentication temporarily unavailable"})
			}
			return
		}
		c.Set("user_id", id)
		c.Set("session_hash", digest)
		c.Next()
	}
}
