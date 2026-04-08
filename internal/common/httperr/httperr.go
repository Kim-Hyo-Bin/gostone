// Package httperr builds Keystone-style JSON error responses (HTTP layer).
package httperr

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/discovery"
	"github.com/gin-gonic/gin"
)

func keystoneURI(c *gin.Context) string {
	return discovery.PreferredV3URL(c.Request)
}

// Unauthorized responds like Keystone (401 + WWW-Authenticate).
func Unauthorized(c *gin.Context, message string) {
	c.Header("WWW-Authenticate", `Keystone uri="`+keystoneURI(c)+`"`)
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    http.StatusUnauthorized,
			"title":   "Unauthorized",
			"message": message,
		},
	})
}

// BadRequest is a 400 error in Keystone JSON shape.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"code":    http.StatusBadRequest,
			"title":   "Bad Request",
			"message": message,
		},
	})
}

// Forbidden is used when the token is valid but policy denies the action.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": gin.H{
			"code":    http.StatusForbidden,
			"title":   "Forbidden",
			"message": message,
		},
	})
}

// InternalServerError responds with HTTP 500 (unexpected server-side failures).
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    http.StatusInternalServerError,
			"title":   "Internal Server Error",
			"message": message,
		},
	})
}

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
