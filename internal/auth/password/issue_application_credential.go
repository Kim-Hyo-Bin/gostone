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

func issueApplicationCredentialFlow(db *gorm.DB, mgr *token.Manager, req *PasswordAuthRequest) (tokenStr string, exp time.Time, body map[string]any, err error) {
	ac := req.Auth.Identity.ApplicationCredential
	if ac == nil || ac.Secret == "" {
		return "", time.Time{}, nil, errors.New("application_credential id and secret required")
	}
	if ac.ID == "" {
		return "", time.Time{}, nil, errors.New("application_credential id and secret required")
	}

	var row models.ApplicationCredential
	if err := db.Where("id = ?", ac.ID).First(&row).Error; err != nil {
		return "", time.Time{}, nil, fmt.Errorf("application credential: invalid id or secret")
	}
	if row.ExpiresAt != nil && time.Now().After(*row.ExpiresAt) {
		return "", time.Time{}, nil, errors.New("application credential expired")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.SecretHash), []byte(ac.Secret)); err != nil {
		return "", time.Time{}, nil, fmt.Errorf("application credential: invalid id or secret")
	}

	var user models.User
	if err := db.Where("id = ?", row.UserID).First(&user).Error; err != nil {
		return "", time.Time{}, nil, err
	}
	if !user.Enabled {
		return "", time.Time{}, nil, errors.New("user disabled")
	}
	var dom models.Domain
	if err := db.Where("id = ?", user.DomainID).First(&dom).Error; err != nil {
		return "", time.Time{}, nil, err
	}

	var rs ResolvedAuthScope
	if row.Unrestricted {
		rs, err = ResolveAuthScope(db, user.ID, dom.ID, req.Auth.Scope)
		if err != nil {
			return "", time.Time{}, nil, err
		}
	} else {
		credRoles, err2 := RolesForApplicationCredential(db, row.ID)
		if err2 != nil {
			return "", time.Time{}, nil, err2
		}
		if len(credRoles) == 0 {
			return "", time.Time{}, nil, errors.New("application credential has no roles")
		}
		if req.Auth.Scope != nil {
			base, err3 := ResolveAuthScope(db, user.ID, dom.ID, req.Auth.Scope)
			if err3 != nil {
				return "", time.Time{}, nil, err3
			}
			rs, err = intersectScopeRoles(base, credRoles)
			if err != nil {
				return "", time.Time{}, nil, err
			}
		} else {
			rs = ResolvedAuthScope{Roles: credRoles}
		}
	}

	methods := []string{"application_credential"}
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
