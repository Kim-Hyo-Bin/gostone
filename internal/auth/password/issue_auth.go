package password

import (
	"errors"
	"fmt"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/catalog"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// IssueAuthToken handles POST /v3/auth/tokens for supported method sets (password, token, application_credential).
// Multiple methods (Keystone MFA) are rejected with a clear error unless every extra method is unimplemented (501 path).
func IssueAuthToken(db *gorm.DB, mgr *token.Manager, req *PasswordAuthRequest) (tokenStr string, exp time.Time, body map[string]any, err error) {
	methods, err := normalizedAuthMethods(req)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	if len(methods) > 1 {
		return issueMultiMethodAuth(db, mgr, req, methods)
	}
	switch methods[0] {
	case "password":
		return issuePasswordFlow(db, mgr, req)
	case "token":
		return issueTokenFlow(db, mgr, req)
	case "application_credential":
		return issueApplicationCredentialFlow(db, mgr, req)
	default:
		if isAuthMethodNotImplemented(methods[0]) {
			return "", time.Time{}, nil, fmt.Errorf("authentication method %q is not implemented", methods[0])
		}
		return "", time.Time{}, nil, fmt.Errorf("unsupported auth method %q", methods[0])
	}
}

func issuePasswordFlow(db *gorm.DB, mgr *token.Manager, req *PasswordAuthRequest) (tokenStr string, exp time.Time, body map[string]any, err error) {
	u := req.Auth.Identity.Password.User
	plain := u.Password
	if plain == "" {
		return "", time.Time{}, nil, errors.New("password required")
	}
	if u.Name == "" && u.ID == "" {
		return "", time.Time{}, nil, errors.New("user name or id required")
	}

	var dom models.Domain
	var user models.User

	// Keystone / Tempest: password auth by user id + password without domain (id is unique).
	if u.ID != "" && u.Domain.ID == "" && u.Domain.Name == "" {
		if err := db.Where("id = ?", u.ID).First(&user).Error; err != nil {
			return "", time.Time{}, nil, fmt.Errorf("user: %w", err)
		}
		if err := db.Where("id = ?", user.DomainID).First(&dom).Error; err != nil {
			return "", time.Time{}, nil, fmt.Errorf("domain: %w", err)
		}
	} else {
		switch {
		case u.Domain.ID != "":
			if err := db.Where("id = ?", u.Domain.ID).First(&dom).Error; err != nil {
				return "", time.Time{}, nil, fmt.Errorf("domain: %w", err)
			}
		case u.Domain.Name != "":
			if err := db.Where("name = ?", u.Domain.Name).First(&dom).Error; err != nil {
				return "", time.Time{}, nil, fmt.Errorf("domain: %w", err)
			}
		default:
			return "", time.Time{}, nil, errors.New("user domain id or name required")
		}

		q := db.Where("domain_id = ?", dom.ID)
		if u.ID != "" {
			q = q.Where("id = ?", u.ID)
		} else {
			q = q.Where("name = ?", u.Name)
		}
		if err := q.First(&user).Error; err != nil {
			return "", time.Time{}, nil, fmt.Errorf("user: %w", err)
		}
	}
	if !user.Enabled {
		return "", time.Time{}, nil, errors.New("user disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plain)); err != nil {
		return "", time.Time{}, nil, errors.New("invalid password")
	}

	rs, err := ResolveAuthScope(db, user.ID, dom.ID, req.Auth.Scope)
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
		Methods:       []string{"password"},
		JTI:           auditID,
	})
	if err != nil {
		return "", time.Time{}, nil, err
	}

	cat, err := catalog.Build(db)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	body, err = assembleTokenEnvelope(db, user, dom, rs, issued, exp, auditID, cat, []string{"password"})
	if err != nil {
		return "", time.Time{}, nil, err
	}
	return tokenStr, exp, body, nil
}

func issueTokenFlow(db *gorm.DB, mgr *token.Manager, req *PasswordAuthRequest) (tokenStr string, exp time.Time, body map[string]any, err error) {
	if req.Auth.Identity.Token == nil || req.Auth.Identity.Token.ID == "" {
		return "", time.Time{}, nil, errors.New("token id required")
	}
	subject := req.Auth.Identity.Token.ID
	claims, err := mgr.Parse(subject)
	if err != nil {
		return "", time.Time{}, nil, fmt.Errorf("invalid token: %w", err)
	}
	var user models.User
	if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return "", time.Time{}, nil, fmt.Errorf("user: %w", err)
	}
	if !user.Enabled {
		return "", time.Time{}, nil, errors.New("user disabled")
	}
	var dom models.Domain
	if err := db.Where("id = ?", user.DomainID).First(&dom).Error; err != nil {
		return "", time.Time{}, nil, fmt.Errorf("domain: %w", err)
	}

	var rs ResolvedAuthScope
	if req.Auth.Scope != nil {
		rs, err = ResolveAuthScope(db, user.ID, dom.ID, req.Auth.Scope)
		if err != nil {
			return "", time.Time{}, nil, err
		}
	} else if claims.ProjectID != "" {
		var pr []string
		pr, err = rolesForProject(db, user.ID, claims.ProjectID)
		if err != nil {
			return "", time.Time{}, nil, err
		}
		rs = ResolvedAuthScope{ProjectID: claims.ProjectID, Roles: pr}
	} else if claims.ScopeDomainID != "" {
		if claims.ScopeDomainID != user.DomainID {
			return "", time.Time{}, nil, errors.New("invalid token domain scope for user")
		}
		var sdom models.Domain
		if err := db.Where("id = ?", claims.ScopeDomainID).First(&sdom).Error; err != nil {
			return "", time.Time{}, nil, fmt.Errorf("scope domain: %w", err)
		}
		dr, err2 := rolesForDomainAssignments(db, user.ID, claims.ScopeDomainID)
		if err2 != nil {
			return "", time.Time{}, nil, err2
		}
		rs = ResolvedAuthScope{ScopeDomainID: claims.ScopeDomainID, Roles: dr, ScopedDomain: sdom}
	} else {
		rs, err = pickUnscopedOrAggregate(db, user.ID, dom.ID)
		if err != nil {
			return "", time.Time{}, nil, err
		}
	}

	auditID := uuid.NewString()
	tokenStr, issued, exp, err := mgr.IssueToken(token.TokenSubject{
		UserID:        user.ID,
		DomainID:      dom.ID,
		ProjectID:     rs.ProjectID,
		ScopeDomainID: rs.ScopeDomainID,
		Roles:         rs.Roles,
		Methods:       []string{"token"},
		JTI:           auditID,
	})
	if err != nil {
		return "", time.Time{}, nil, err
	}
	cat, err := catalog.Build(db)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	body, err = assembleTokenEnvelope(db, user, dom, rs, issued, exp, auditID, cat, []string{"token"})
	if err != nil {
		return "", time.Time{}, nil, err
	}
	return tokenStr, exp, body, nil
}
