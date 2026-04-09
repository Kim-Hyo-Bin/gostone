package password

import (
	"fmt"
	"strings"
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
	exp := time.Now().UTC()
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	issuedAt := time.Now().UTC()
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	auditID := strings.TrimSpace(claims.ID)
	if auditID == "" {
		auditID = uuid.NewString()
	}
	cat, err := catalog.Build(db)
	if err != nil {
		return nil, err
	}
	methods := claims.Methods
	if len(methods) == 0 {
		methods = []string{"password"}
	}
	rs := ResolvedAuthScope{
		ProjectID:     claims.ProjectID,
		ScopeDomainID: claims.ScopeDomainID,
		Roles:         claims.Roles,
	}
	if claims.ScopeDomainID != "" {
		var d models.Domain
		if err := db.Where("id = ?", claims.ScopeDomainID).First(&d).Error; err != nil {
			return nil, fmt.Errorf("scope domain: %w", err)
		}
		rs.ScopedDomain = d
	}
	return assembleTokenEnvelope(db, user, dom, rs, issuedAt, exp, auditID, cat, methods)
}
