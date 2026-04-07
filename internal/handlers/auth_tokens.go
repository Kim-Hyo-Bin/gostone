package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/apierrors"
	authops "github.com/Kim-Hyo-Bin/gostone/internal/ops/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Hub) postAuthTokens(c *gin.Context) {
	var req authops.PasswordAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.BadRequest(c, "Malformed request body")
		return
	}
	tok, _, body, err := authops.IssuePasswordToken(h.DB, h.Tokens, &req)
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
		apierrors.Unauthorized(c, "Missing X-Auth-Token")
		return
	}
	claims, err := h.Tokens.Parse(raw)
	if err != nil {
		apierrors.Unauthorized(c, "Invalid token.")
		return
	}
	body, err := authops.BuildTokenResponse(h.DB, claims)
	if err != nil {
		apierrors.Unauthorized(c, "Invalid token.")
		return
	}
	c.JSON(http.StatusOK, body)
}

func (h *Hub) headAuthTokens(c *gin.Context) {
	// Token already validated by auth middleware for this path.
	c.Status(http.StatusOK)
}

func (h *Hub) deleteAuthTokens(c *gin.Context) {
	// JWT has no server-side revocation yet (Fernet/UUID + revoke list would).
	c.Status(http.StatusNoContent)
}

func mapAuthError(c *gin.Context, err error) {
	msg := err.Error()
	if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(msg, "invalid password") ||
		strings.Contains(msg, "user:") || strings.Contains(msg, "domain:") {
		apierrors.Unauthorized(c, "Invalid username or password")
		return
	}
	switch {
	case strings.Contains(msg, "unsupported"), strings.Contains(msg, "required"):
		apierrors.BadRequest(c, msg)
	default:
		apierrors.Unauthorized(c, "Authentication failed")
	}
}
