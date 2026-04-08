package password

import (
	"errors"
	"fmt"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// pickProjectAndRoles selects project id and role names for a user in a domain.
// With no project scope, aggregates all project roles in the domain (first project id wins for token scope).
// With project scope, narrows to that project and verifies the user has an assignment.
func pickProjectAndRoles(db *gorm.DB, userID, domainID string, scope *AuthScope) (projectID string, roleNames []string, err error) {
	if scope != nil && scope.Project != nil {
		return pickScopedProjectRoles(db, userID, domainID, scope)
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
		Where("user_project_roles.user_id = ? AND projects.domain_id = ?", userID, domainID).
		Scan(&rows).Error
	if err != nil {
		return "", nil, err
	}
	seen := map[string]struct{}{}
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
		Where("user_domain_roles.user_id = ? AND user_domain_roles.domain_id = ?", userID, domainID).
		Scan(&domRoleRows).Error; err != nil {
		return "", nil, err
	}
	for _, dr := range domRoleRows {
		if _, ok := seen[dr.Name]; !ok {
			seen[dr.Name] = struct{}{}
			roleNames = append(roleNames, dr.Name)
		}
	}
	return projectID, roleNames, nil
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
