package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const bearerPrefix = "Bearer "

// BearerOrXAuthToken returns the token from Authorization: Bearer <token> if present,
// otherwise the X-Auth-Token header (Keystone-compatible).
func BearerOrXAuthToken(c *gin.Context) string {
	h := strings.TrimSpace(c.GetHeader("Authorization"))
	if len(h) > len(bearerPrefix) && strings.EqualFold(h[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(h[len(bearerPrefix):])
	}
	return strings.TrimSpace(c.GetHeader("X-Auth-Token"))
}

// SubjectOrBearerToken returns X-Subject-Token when set (Keystone token to validate or show),
// otherwise the same token source as BearerOrXAuthToken (caller credential).
func SubjectOrBearerToken(c *gin.Context) string {
	if s := strings.TrimSpace(c.GetHeader("X-Subject-Token")); s != "" {
		return s
	}
	return BearerOrXAuthToken(c)
}
