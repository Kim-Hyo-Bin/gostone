package password

import (
	"errors"
	"fmt"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// ResolvedAuthScope is the effective Keystone scope after applying auth.scope JSON.
type ResolvedAuthScope struct {
	ProjectID     string
	ScopeDomainID string
	Roles         []string
	// ScopedDomain is set when ScopeDomainID is non-empty (for token JSON "domain").
	ScopedDomain models.Domain
}

// ResolveAuthScope maps auth.scope to project id, optional domain scope, and role names.
func ResolveAuthScope(db *gorm.DB, userID, userHomeDomainID string, scope *AuthScope) (ResolvedAuthScope, error) {
	if scope != nil && scope.Domain != nil && scope.Project != nil {
		return ResolvedAuthScope{}, errors.New("ambiguous scope: specify either domain or project, not both")
	}
	if scope != nil && scope.Domain != nil {
		return resolveDomainScope(db, userID, userHomeDomainID, scope)
	}
	if scope != nil && scope.Project != nil {
		projectID, roles, err := pickScopedProjectRoles(db, userID, userHomeDomainID, scope)
		if err != nil {
			return ResolvedAuthScope{}, err
		}
		return ResolvedAuthScope{ProjectID: projectID, Roles: roles}, nil
	}
	return pickUnscopedOrAggregate(db, userID, userHomeDomainID)
}

func resolveDomainScope(db *gorm.DB, userID, userHomeDomainID string, scope *AuthScope) (ResolvedAuthScope, error) {
	sd := scope.Domain
	var dom models.Domain
	switch {
	case sd.ID != "":
		if err := db.Where("id = ?", sd.ID).First(&dom).Error; err != nil {
			return ResolvedAuthScope{}, fmt.Errorf("scope domain: %w", err)
		}
	case sd.Name != "":
		if err := db.Where("name = ?", sd.Name).First(&dom).Error; err != nil {
			return ResolvedAuthScope{}, fmt.Errorf("scope domain: %w", err)
		}
	default:
		return ResolvedAuthScope{}, errors.New("domain id or name required in scope")
	}
	if dom.ID != userHomeDomainID {
		return ResolvedAuthScope{}, errors.New("domain scope must match the user's domain")
	}
	roles, err := rolesForDomainAssignments(db, userID, dom.ID)
	if err != nil {
		return ResolvedAuthScope{}, err
	}
	return ResolvedAuthScope{
		ScopeDomainID: dom.ID,
		Roles:         roles,
		ScopedDomain:  dom,
	}, nil
}

func rolesForDomainAssignments(db *gorm.DB, userID, domainID string) ([]string, error) {
	var rows []struct{ Name string }
	if err := db.Model(&models.UserDomainRole{}).
		Select("roles.name AS name").
		Joins("JOIN roles ON roles.id = user_domain_roles.role_id").
		Where("user_domain_roles.user_id = ? AND user_domain_roles.domain_id = ?", userID, domainID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Name)
	}
	return out, nil
}

// pickUnscopedOrAggregate mirrors Keystone unscoped issuance: aggregate domain + project roles,
// pick a default project id when the user has project assignments.
func pickUnscopedOrAggregate(db *gorm.DB, userID, domainID string) (ResolvedAuthScope, error) {
	type row struct {
		RoleName  string
		ProjectID string
	}
	var rows []row
	err := db.Model(&models.UserProjectRole{}).
		Select("roles.name AS role_name", "user_project_roles.project_id AS project_id").
		Joins("JOIN roles ON roles.id = user_project_roles.role_id").
		Joins("JOIN projects ON projects.id = user_project_roles.project_id").
		Where("user_project_roles.user_id = ? AND projects.domain_id = ?", userID, domainID).
		Scan(&rows).Error
	if err != nil {
		return ResolvedAuthScope{}, err
	}
	seen := map[string]struct{}{}
	var roleNames []string
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
	domRoles, err := rolesForDomainAssignments(db, userID, domainID)
	if err != nil {
		return ResolvedAuthScope{}, err
	}
	for _, name := range domRoles {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			roleNames = append(roleNames, name)
		}
	}
	return ResolvedAuthScope{ProjectID: projectID, Roles: roleNames}, nil
}

func pickScopedProjectRoles(db *gorm.DB, userID, userDomainID string, scope *AuthScope) (projectID string, roleNames []string, err error) {
	sp := scope.Project
	var proj models.Project
	switch {
	case sp.ID != "":
		if err := db.Where("id = ?", sp.ID).First(&proj).Error; err != nil {
			return "", nil, fmt.Errorf("project: %w", err)
		}
	case sp.Name != "":
		domID := userDomainID
		if sp.Domain != nil {
			if sp.Domain.ID != "" {
				domID = sp.Domain.ID
			} else if sp.Domain.Name != "" {
				var d models.Domain
				if err := db.Where("name = ?", sp.Domain.Name).First(&d).Error; err != nil {
					return "", nil, fmt.Errorf("scope domain: %w", err)
				}
				domID = d.ID
			}
		}
		if err := db.Where("domain_id = ? AND name = ?", domID, sp.Name).First(&proj).Error; err != nil {
			return "", nil, fmt.Errorf("project: %w", err)
		}
	default:
		return "", nil, errors.New("project id or name required in scope")
	}
	var u models.User
	if err := db.Where("id = ?", userID).First(&u).Error; err != nil {
		return "", nil, err
	}
	if proj.DomainID != u.DomainID {
		return "", nil, errors.New("project is not in the user domain")
	}
	type rrow struct{ Name string }
	var rrows []rrow
	if err := db.Model(&models.UserProjectRole{}).
		Select("roles.name AS name").
		Joins("JOIN roles ON roles.id = user_project_roles.role_id").
		Where("user_project_roles.user_id = ? AND user_project_roles.project_id = ?", userID, proj.ID).
		Scan(&rrows).Error; err != nil {
		return "", nil, err
	}
	if len(rrows) == 0 {
		return "", nil, errors.New("user has no access to scoped project")
	}
	for _, r := range rrows {
		roleNames = append(roleNames, r.Name)
	}
	return proj.ID, roleNames, nil
}

func rolesForProject(db *gorm.DB, userID, projectID string) ([]string, error) {
	type rrow struct{ Name string }
	var rrows []rrow
	if err := db.Model(&models.UserProjectRole{}).
		Select("roles.name AS name").
		Joins("JOIN roles ON roles.id = user_project_roles.role_id").
		Where("user_project_roles.user_id = ? AND user_project_roles.project_id = ?", userID, projectID).
		Scan(&rrows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rrows))
	for _, r := range rrows {
		out = append(out, r.Name)
	}
	return out, nil
}

// RolesForApplicationCredential returns role names bound to an application credential.
func RolesForApplicationCredential(db *gorm.DB, appCredID string) ([]string, error) {
	type rrow struct{ Name string }
	var rrows []rrow
	err := db.Model(&models.ApplicationCredentialRole{}).
		Select("roles.name AS name").
		Joins("JOIN roles ON roles.id = application_credential_roles.role_id").
		Where("application_credential_roles.app_cred_id = ?", appCredID).
		Scan(&rrows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rrows))
	for _, r := range rrows {
		out = append(out, r.Name)
	}
	return out, nil
}

func intersectScopeRoles(base ResolvedAuthScope, credRoles []string) (ResolvedAuthScope, error) {
	set := make(map[string]struct{}, len(credRoles))
	for _, r := range credRoles {
		set[r] = struct{}{}
	}
	var out []string
	for _, r := range base.Roles {
		if _, ok := set[r]; ok {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return ResolvedAuthScope{}, errors.New("no roles from application credential match this scope")
	}
	base.Roles = out
	return base, nil
}
