package server

import (
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
	"github.com/gin-gonic/gin"
)

// adminOnlyListenerMiddleware rejects requests on non-admin listeners when the path matches a prefix.
// Labels "admin" and "listen" (single-interface mode) always pass.
func adminOnlyListenerMiddleware(prefixes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		label := listenctx.Label(c.Request.Context())
		if label == "admin" || label == "listen" || label == "" {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		for _, p := range prefixes {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if strings.HasPrefix(path, p) {
				httperr.Forbidden(c, "This resource is available only on the admin API endpoint.")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
