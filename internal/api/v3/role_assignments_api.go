package v3

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerV3RoleAssignments(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/role_assignments")
	g.GET("", h.listRoleAssignments)
	g.POST("", h.createRoleAssignment)
}

func (h *Hub) listRoleAssignments(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_role_assignments", nil); !ok {
		return
	}

	type row struct {
		UserID    string
		ProjectID string
		RoleID    string
		RoleName  string
	}
	q := h.DB.Table("user_project_roles").
		Select("user_project_roles.user_id, user_project_roles.project_id, user_project_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = user_project_roles.role_id")

	if v := c.Query("user.id"); v != "" {
		q = q.Where("user_project_roles.user_id = ?", v)
	}
	if v := c.Query("scope.project.id"); v != "" {
		q = q.Where("user_project_roles.project_id = ?", v)
	}
	if v := c.Query("role.id"); v != "" {
		q = q.Where("user_project_roles.role_id = ?", v)
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	assignments := make([]map[string]any, 0, len(rows)+32)
	for _, r := range rows {
		assignments = append(assignments, map[string]any{
			"scope": map[string]any{
				"project": map[string]any{"id": r.ProjectID},
			},
			"role": map[string]any{
				"id":   r.RoleID,
				"name": r.RoleName,
			},
			"user": map[string]any{"id": r.UserID},
		})
	}

	type drow struct {
		UserID   string
		DomainID string
		RoleID   string
		RoleName string
	}
	dq := h.DB.Table("user_domain_roles").
		Select("user_domain_roles.user_id, user_domain_roles.domain_id, user_domain_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = user_domain_roles.role_id")
	if v := c.Query("user.id"); v != "" {
		dq = dq.Where("user_domain_roles.user_id = ?", v)
	}
	if v := c.Query("scope.domain.id"); v != "" {
		dq = dq.Where("user_domain_roles.domain_id = ?", v)
	}
	if v := c.Query("role.id"); v != "" {
		dq = dq.Where("user_domain_roles.role_id = ?", v)
	}
	var drows []drow
	if err := dq.Scan(&drows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	for _, r := range drows {
		assignments = append(assignments, map[string]any{
			"scope": map[string]any{
				"domain": map[string]any{"id": r.DomainID},
			},
			"role": map[string]any{
				"id":   r.RoleID,
				"name": r.RoleName,
			},
			"user": map[string]any{"id": r.UserID},
		})
	}

	type grow struct {
		GroupID   string
		ProjectID string
		RoleID    string
		RoleName  string
	}
	gq := h.DB.Table("group_project_roles").
		Select("group_project_roles.group_id, group_project_roles.project_id, group_project_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = group_project_roles.role_id")
	if v := c.Query("group.id"); v != "" {
		gq = gq.Where("group_project_roles.group_id = ?", v)
	}
	if v := c.Query("scope.project.id"); v != "" {
		gq = gq.Where("group_project_roles.project_id = ?", v)
	}
	if v := c.Query("role.id"); v != "" {
		gq = gq.Where("group_project_roles.role_id = ?", v)
	}
	var grows []grow
	if err := gq.Scan(&grows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	for _, r := range grows {
		assignments = append(assignments, map[string]any{
			"scope": map[string]any{
				"project": map[string]any{"id": r.ProjectID},
			},
			"role": map[string]any{
				"id":   r.RoleID,
				"name": r.RoleName,
			},
			"group": map[string]any{"id": r.GroupID},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"role_assignments": assignments,
		"links":            gin.H{"self": selfURL(c, "/v3/role_assignments")},
	})
}
