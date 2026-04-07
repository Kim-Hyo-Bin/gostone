// Package apierrors builds Keystone-style JSON error bodies.
package apierrors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotImplemented responds with HTTP 501 and an Identity-API-shaped error envelope.
func NotImplemented(c *gin.Context, message string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    http.StatusNotImplemented,
			"title":   "Not Implemented",
			"message": message,
		},
	})
}
