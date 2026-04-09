package password

import (
	"errors"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/catalog"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IssuePasswordToken validates password auth (single method) and issues a token.
func IssuePasswordToken(db *gorm.DB, mgr *token.Manager, req *PasswordAuthRequest) (tokenStr string, exp time.Time, body map[string]any, err error) {
	if req == nil {
		return "", time.Time{}, nil, errors.New("empty auth request")
	}
	if len(req.Auth.Identity.Methods) != 1 || req.Auth.Identity.Methods[0] != "password" {
		return "", time.Time{}, nil, errors.New("unsupported auth methods")
	}
	return issuePasswordFlow(db, mgr, req)
}

// IssueTokenForUser issues a token for an existing user id using the same role/catalog logic as password auth.
func IssueTokenForUser(db *gorm.DB, mgr *token.Manager, userID string, methods []string) (tokenStr string, exp time.Time, body map[string]any, err error) {
	if len(methods) == 0 {
		methods = []string{"token"}
	}
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return "", time.Time{}, nil, err
	}
	if !user.Enabled {
		return "", time.Time{}, nil, errors.New("user disabled")
	}
	var dom models.Domain
	if err := db.Where("id = ?", user.DomainID).First(&dom).Error; err != nil {
		return "", time.Time{}, nil, err
	}
	rs, err := pickUnscopedOrAggregate(db, user.ID, dom.ID)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	auditID := uuid.NewString()
	tokenStr, issued, exp, err := mgr.IssueToken(token.TokenSubject{
		UserID:        user.ID,
		DomainID:      dom.ID,
		ProjectID:     rs.ProjectID,
		ScopeDomainID: rs.ScopeDomainID,
		Roles:         rs.Roles,
		Methods:       methods,
		JTI:           auditID,
	})
	if err != nil {
		return "", time.Time{}, nil, err
	}
	cat, err := catalog.Build(db)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	body, err = assembleTokenEnvelope(db, user, dom, rs, issued, exp, auditID, cat, methods)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	return tokenStr, exp, body, nil
}
