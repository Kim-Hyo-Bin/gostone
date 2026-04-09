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
	c.JSON(http.StatusUnauthorized, errorEnvelope(c, http.StatusUnauthorized, "Unauthorized", message))
}

// BadRequest is a 400 error in Keystone JSON shape.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, errorEnvelope(c, http.StatusBadRequest, "Bad Request", message))
}

// Forbidden is used when the token is valid but policy denies the action.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, errorEnvelope(c, http.StatusForbidden, "Forbidden", message))
}

// NotFound is HTTP 404 with Keystone-style JSON (e.g. revoked or unknown token).
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, errorEnvelope(c, http.StatusNotFound, "Not Found", message))
}

// InternalServerError responds with HTTP 500 (unexpected server-side failures).
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, errorEnvelope(c, http.StatusInternalServerError, "Internal Server Error", message))
}

// NotImplemented responds with HTTP 501 and an Identity-API-shaped error envelope.
func NotImplemented(c *gin.Context, message string) {
	c.JSON(http.StatusNotImplemented, errorEnvelope(c, http.StatusNotImplemented, "Not Implemented", message))
}
