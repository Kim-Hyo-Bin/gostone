package auth

import (
	"fmt"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BuildTokenResponse rebuilds the token JSON for GET /v3/auth/tokens from JWT claims and DB.
func BuildTokenResponse(db *gorm.DB, claims *token.Claims) (map[string]any, error) {
	var user models.User
	if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	var dom models.Domain
	if err := db.Where("id = ?", user.DomainID).First(&dom).Error; err != nil {
		return nil, fmt.Errorf("domain: %w", err)
	}
	exp := time.Now()
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	auditID := uuid.NewString()
	return buildTokenEnvelope(user, dom, claims.ProjectID, claims.Roles, exp, auditID), nil
}
