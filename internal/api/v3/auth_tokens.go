package v3

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth/password"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Hub) postAuthTokens(c *gin.Context) {
	var req password.PasswordAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	tok, _, body, err := password.IssueAuthToken(h.DB, h.Tokens, &req)
	if err != nil {
		mapAuthError(c, err)
		return
	}
	c.Header("X-Subject-Token", tok)
	c.Header("X-Auth-Token", tok)
	c.JSON(http.StatusCreated, body)
}

func (h *Hub) getAuthTokens(c *gin.Context) {
	raw := c.GetHeader("X-Auth-Token")
	if raw == "" {
		httperr.Unauthorized(c, "Missing X-Auth-Token")
		return
	}
	claims, err := h.Tokens.Parse(raw)
	if err != nil {
		httperr.Unauthorized(c, "Invalid token.")
		return
	}
	body, err := password.BuildTokenResponse(h.DB, claims)
	if err != nil {
		httperr.Unauthorized(c, "Invalid token.")
		return
	}
	c.JSON(http.StatusOK, body)
}

func (h *Hub) headAuthTokens(c *gin.Context) {
	// Token already validated by auth middleware for this path.
	c.Status(http.StatusOK)
}

func (h *Hub) deleteAuthTokens(c *gin.Context) {
	raw := c.GetHeader("X-Auth-Token")
	if raw != "" {
		_ = h.Tokens.Revoke(raw)
	}
	c.Status(http.StatusNoContent)
}

func mapAuthError(c *gin.Context, err error) {
	msg := err.Error()
	if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(msg, "invalid password") ||
		strings.Contains(msg, "user:") || strings.Contains(msg, "domain:") {
		httperr.Unauthorized(c, "Invalid username or password")
		return
	}
	switch {
	case strings.Contains(msg, "unsupported"), strings.Contains(msg, "required"),
		strings.Contains(msg, "expected exactly one method"):
		httperr.BadRequest(c, msg)
	default:
		httperr.Unauthorized(c, "Authentication failed")
	}
}
