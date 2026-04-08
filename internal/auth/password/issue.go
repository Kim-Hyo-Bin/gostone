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
	type row struct {
		RoleName  string
		ProjectID string
	}
	var rows []row
	err = db.Model(&models.UserProjectRole{}).
		Select("roles.name AS role_name", "user_project_roles.project_id AS project_id").
		Joins("JOIN roles ON roles.id = user_project_roles.role_id").
		Joins("JOIN projects ON projects.id = user_project_roles.project_id").
		Where("user_project_roles.user_id = ? AND projects.domain_id = ?", user.ID, dom.ID).
		Scan(&rows).Error
	if err != nil {
		return "", time.Time{}, nil, err
	}
	roleNames := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	var projectID string
	for _, r := range rows {
		if _, ok := seen[r.RoleName]; !ok {
			seen[r.RoleName] = struct{}{}
			roleNames = append(roleNames, r.RoleName)
		}
		if projectID == "" {
			projectID = r.ProjectID
		}
	}
	var domRoleRows []struct{ Name string }
	if err := db.Model(&models.UserDomainRole{}).
		Select("roles.name AS name").
		Joins("JOIN roles ON roles.id = user_domain_roles.role_id").
		Where("user_domain_roles.user_id = ? AND user_domain_roles.domain_id = ?", user.ID, dom.ID).
		Scan(&domRoleRows).Error; err != nil {
		return "", time.Time{}, nil, err
	}
	for _, dr := range domRoleRows {
		if _, ok := seen[dr.Name]; !ok {
			seen[dr.Name] = struct{}{}
			roleNames = append(roleNames, dr.Name)
		}
	}
	auditID := uuid.NewString()
	tokenStr, exp, err = mgr.IssueWithMethods(user.ID, dom.ID, projectID, roleNames, methods)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	cat, err := catalog.Build(db)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	body = buildTokenEnvelope(user, dom, projectID, roleNames, exp, auditID, cat, methods)
	return tokenStr, exp, body, nil
}

func buildTokenEnvelope(user models.User, dom models.Domain, projectID string, roles []string, exp time.Time, auditID string, catalogObjs []any, methods []string) map[string]any {
	issued := time.Now().UTC().Format(time.RFC3339Nano)
	expires := exp.UTC().Format(time.RFC3339Nano)
	tok := map[string]any{
		"methods":       methods,
		"user":          map[string]any{"id": user.ID, "name": user.Name, "domain": map[string]any{"id": dom.ID, "name": dom.Name}},
		"roles":         roles,
		"expires_at":    expires,
		"issued_at":     issued,
		"audit_ids":     []string{auditID},
		"catalog":       catalogObjs,
		"project_scope": projectID != "",
	}
	if projectID != "" {
		tok["project"] = map[string]any{"id": projectID}
	}
	return map[string]any{"token": tok}
}
