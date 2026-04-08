package password

import (
	"fmt"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/catalog"
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
	cat, err := catalog.Build(db)
	if err != nil {
		return nil, err
	}
	methods := claims.Methods
	if len(methods) == 0 {
		methods = []string{"password"}
	}
	var scopedDom *models.Domain
	if claims.ScopeDomainID != "" {
		var d models.Domain
		if err := db.Where("id = ?", claims.ScopeDomainID).First(&d).Error; err != nil {
			return nil, fmt.Errorf("scope domain: %w", err)
		}
		scopedDom = &d
	}
	return buildTokenEnvelope(user, dom, claims.ProjectID, scopedDom, claims.Roles, exp, auditID, cat, methods), nil
}
