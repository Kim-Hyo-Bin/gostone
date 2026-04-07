// Package httperr builds Keystone-style JSON error responses (HTTP layer).
package httperr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func keystoneURI(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + "/v3/"
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
