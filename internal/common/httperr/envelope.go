package httperr

import "github.com/gin-gonic/gin"

// GinRequestIDKey matches the Gin context key set by server.requestIDMiddleware.
const GinRequestIDKey = "request_id"

func errorEnvelope(c *gin.Context, code int, title, message string) gin.H {
	obj := gin.H{
		"code":    code,
		"title":   title,
		"message": message,
	}
	if rid, ok := c.Get(GinRequestIDKey); ok {
		if s, ok := rid.(string); ok && s != "" {
			obj["request_id"] = s
		}
	}
	return gin.H{"error": obj}
}
