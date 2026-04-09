package auth

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
)

// Middleware validates Authorization: Bearer or X-Auth-Token for protected Identity routes (Keystone-style).
func Middleware(mgr *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Vary", "X-Auth-Token, Authorization")

		if SkipAuth(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		raw := BearerOrXAuthToken(c)
		if raw == "" {
			httperr.Unauthorized(c, "The request you have made requires authentication.")
			c.Abort()
			return
		}

		claims, err := mgr.Parse(raw)
		if err != nil {
			httperr.Unauthorized(c, "Invalid token.")
			c.Abort()
			return
		}

		c.Set(GinKey, Context{
			UserID:        claims.UserID,
			DomainID:      claims.DomainID,
			ProjectID:     claims.ProjectID,
			ScopeDomainID: claims.ScopeDomainID,
			Roles:         claims.Roles,
		})
		c.Next()
	}
}
