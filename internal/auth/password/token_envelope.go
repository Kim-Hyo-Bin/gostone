package password

import (
	"fmt"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/timefmt"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// assembleTokenEnvelope builds a Keystone-shaped token JSON (roles as objects, full project ref).
func assembleTokenEnvelope(db *gorm.DB, user models.User, userDomain models.Domain, rs ResolvedAuthScope, issuedAt, exp time.Time, auditID string, catalogObjs []any, methods []string) (map[string]any, error) {
	roleObjs, err := resolveTokenRoles(db, user.DomainID, rs.Roles)
	if err != nil {
		return nil, err
	}
	var projRef map[string]any
	if rs.ProjectID != "" {
		projRef, err = projectTokenRef(db, rs.ProjectID)
		if err != nil {
			return nil, err
		}
	}
	var scopedPtr *models.Domain
	if rs.ScopeDomainID != "" {
		d := rs.ScopedDomain
		scopedPtr = &d
	}
	return buildTokenEnvelope(user, userDomain, projRef, scopedPtr, roleObjs, issuedAt, exp, auditID, catalogObjs, methods), nil
}

func resolveTokenRoles(db *gorm.DB, userDomainID string, names []string) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(names))
	for _, n := range names {
		r, err := lookupRoleByName(db, userDomainID, n)
		if err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": r.ID, "name": r.Name})
	}
	return out, nil
}

func lookupRoleByName(db *gorm.DB, userDomainID, name string) (models.Role, error) {
	var r models.Role
	// Domain-scoped role first, then global (empty domain_id) like typical Keystone.
	if err := db.Where("name = ? AND domain_id = ?", name, userDomainID).First(&r).Error; err == nil {
		return r, nil
	}
	if err := db.Where("name = ? AND domain_id = ?", name, "").First(&r).Error; err == nil {
		return r, nil
	}
	if err := db.Where("name = ?", name).First(&r).Error; err != nil {
		return r, fmt.Errorf("role %q: %w", name, err)
	}
	return r, nil
}

func projectTokenRef(db *gorm.DB, projectID string) (map[string]any, error) {
	var p models.Project
	if err := db.Where("id = ?", projectID).First(&p).Error; err != nil {
		return nil, err
	}
	var d models.Domain
	if err := db.Where("id = ?", p.DomainID).First(&d).Error; err != nil {
		return nil, err
	}
	return map[string]any{
		"id":     p.ID,
		"name":   p.Name,
		"domain": map[string]any{"id": d.ID, "name": d.Name},
	}, nil
}

func buildTokenEnvelope(user models.User, userDomain models.Domain, projectRef map[string]any, scopedDomain *models.Domain, roles []map[string]any, issuedAt, exp time.Time, auditID string, catalogObjs []any, methods []string) map[string]any {
	if issuedAt.IsZero() {
		issuedAt = time.Now().UTC()
	}
	issued := timefmt.KeystoneUTC(issuedAt)
	expires := timefmt.KeystoneUTC(exp)
	tok := map[string]any{
		"methods":       methods,
		"user":          map[string]any{"id": user.ID, "name": user.Name, "domain": map[string]any{"id": userDomain.ID, "name": userDomain.Name}},
		"roles":         roles,
		"expires_at":    expires,
		"issued_at":     issued,
		"audit_ids":     []string{auditID},
		"catalog":       catalogObjs,
		"project_scope": projectRef != nil,
	}
	if scopedDomain != nil {
		tok["domain"] = map[string]any{"id": scopedDomain.ID, "name": scopedDomain.Name}
	}
	if projectRef != nil {
		tok["project"] = projectRef
	}
	return map[string]any{"token": tok}
}
