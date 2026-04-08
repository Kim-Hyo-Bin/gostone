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
	tokenStr, exp, err = mgr.IssueToken(token.TokenSubject{
		UserID:        user.ID,
		DomainID:      dom.ID,
		ProjectID:     rs.ProjectID,
		ScopeDomainID: rs.ScopeDomainID,
		Roles:         rs.Roles,
		Methods:       methods,
	})
	if err != nil {
		return "", time.Time{}, nil, err
	}
	cat, err := catalog.Build(db)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	var scopedPtr *models.Domain
	if rs.ScopeDomainID != "" {
		d := rs.ScopedDomain
		scopedPtr = &d
	}
	body = buildTokenEnvelope(user, dom, rs.ProjectID, scopedPtr, rs.Roles, exp, auditID, cat, methods)
	return tokenStr, exp, body, nil
}

func buildTokenEnvelope(user models.User, userDomain models.Domain, projectID string, scopedDomain *models.Domain, roles []string, exp time.Time, auditID string, catalogObjs []any, methods []string) map[string]any {
	issued := time.Now().UTC().Format(time.RFC3339Nano)
	expires := exp.UTC().Format(time.RFC3339Nano)
	tok := map[string]any{
		"methods":       methods,
		"user":          map[string]any{"id": user.ID, "name": user.Name, "domain": map[string]any{"id": userDomain.ID, "name": userDomain.Name}},
		"roles":         roles,
		"expires_at":    expires,
		"issued_at":     issued,
		"audit_ids":     []string{auditID},
		"catalog":       catalogObjs,
		"project_scope": projectID != "",
	}
	if scopedDomain != nil {
		tok["domain"] = map[string]any{"id": scopedDomain.ID, "name": scopedDomain.Name}
	}
	if projectID != "" {
		tok["project"] = map[string]any{"id": projectID}
	}
	return map[string]any{"token": tok}
}
