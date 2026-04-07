package v3

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/gin-gonic/gin"
)

func registerV3RoleAssignments(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/role_assignments")
	g.GET("", h.listRoleAssignments)
	g.POST("", h.stubRoute("POST /v3/role_assignments"))
}

func (h *Hub) listRoleAssignments(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:list_role_assignments", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
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

	assignments := make([]map[string]any, 0, len(rows))
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
	c.JSON(http.StatusOK, gin.H{
		"role_assignments": assignments,
		"links":            gin.H{"self": selfURL(c, "/v3/role_assignments")},
	})
}
