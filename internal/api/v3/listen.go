package v3

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
	"github.com/gin-gonic/gin"
)

// ListenBinding returns the HTTP bind name for this request (public, admin, internal, listen), or "" if unset.
func ListenBinding(c *gin.Context) string {
	return listenctx.Label(c.Request.Context())
}
