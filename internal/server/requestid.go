package server

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// requestIDGinKey matches httperr.GinRequestIDKey for error JSON and access logs.
const requestIDGinKey = httperr.GinRequestIDKey

// requestIDMiddleware sets X-OpenStack-Request-Id (generating one if absent) for OpenStack client compatibility.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-OpenStack-Request-Id")
		if rid == "" {
			rid = c.GetHeader("X-Openstack-Request-Id")
		}
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set("X-OpenStack-Request-Id", rid)
		c.Set(requestIDGinKey, rid)
		c.Next()
	}
}
