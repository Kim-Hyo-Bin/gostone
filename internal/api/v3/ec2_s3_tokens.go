package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth/password"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Hub) postEc2Tokens(c *gin.Context) {
	var req struct {
		Auth *struct {
			Identity *struct {
				Methods []string `json:"methods"`
				EC2     *struct {
					Token   string            `json:"token"`
					Creds   string            `json:"credential"`
					Access  string            `json:"access"`
					Secret  string            `json:"secret"`
					Params  map[string]string `json:"params"`
					Headers map[string]string `json:"headers"`
				} `json:"ec2"`
			} `json:"identity"`
		} `json:"auth"`
		// Legacy-style body
		EC2Creds *struct {
			Access string `json:"access"`
			Secret string `json:"secret"`
		} `json:"ec2Credentials"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var access, secret string
	if req.EC2Creds != nil {
		access, secret = req.EC2Creds.Access, req.EC2Creds.Secret
	}
	if req.Auth != nil && req.Auth.Identity != nil && req.Auth.Identity.EC2 != nil {
		e := req.Auth.Identity.EC2
		if e.Access != "" {
			access = e.Access
		}
		if e.Secret != "" {
			secret = e.Secret
		}
	}
	if access == "" || secret == "" {
		httperr.Unauthorized(c, "EC2 access and secret required")
		return
	}
	var cred models.EC2Credential
	if err := h.DB.Where("access_key = ?", access).First(&cred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httperr.Unauthorized(c, "Invalid EC2 credentials")
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if cred.SecretKey != secret {
		httperr.Unauthorized(c, "Invalid EC2 credentials")
		return
	}
	tok, _, body, err := password.IssueTokenForUser(h.DB, h.Tokens, cred.UserID, []string{"ec2"})
	if err != nil {
		mapAuthError(c, err)
		return
	}
	c.Header("X-Subject-Token", tok)
	c.Header("X-Auth-Token", tok)
	c.JSON(http.StatusCreated, body)
}

func (h *Hub) postS3Tokens(c *gin.Context) {
	var req password.PasswordAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	if len(req.Auth.Identity.Methods) == 0 {
		req.Auth.Identity.Methods = []string{"password"}
	}
	for i, m := range req.Auth.Identity.Methods {
		req.Auth.Identity.Methods[i] = strings.ToLower(strings.TrimSpace(m))
		if req.Auth.Identity.Methods[i] == "s3" {
			req.Auth.Identity.Methods[i] = "password"
		}
	}
	tok, _, body, err := password.IssuePasswordToken(h.DB, h.Tokens, &req)
	if err != nil {
		mapAuthError(c, err)
		return
	}
	if tokm, ok := body["token"].(map[string]any); ok {
		tokm["methods"] = []string{"s3"}
	}
	c.Header("X-Subject-Token", tok)
	c.Header("X-Auth-Token", tok)
	c.JSON(http.StatusCreated, body)
}
