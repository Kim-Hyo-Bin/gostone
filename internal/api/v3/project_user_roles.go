package v3

import (
	"errors"
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errAssignRefsNotFound = errors.New("project user or role not found")

func (h *Hub) listProjectUserRoles(c *gin.Context) {
	projectID := c.Param("project_id")
	userID := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_project_user_roles", map[string]string{"project_id": projectID}); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	err := h.DB.Table("user_project_roles").
		Select("user_project_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = user_project_roles.role_id").
		Where("user_project_roles.project_id = ? AND user_project_roles.user_id = ?", projectID, userID).
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id":   r.RoleID,
			"name": r.RoleName,
		})
	}
	c.JSON(http.StatusOK, gin.H{"roles": out})
}

func (h *Hub) putProjectUserRole(c *gin.Context) {
	projectID := c.Param("project_id")
	userID := c.Param("user_id")
	roleID := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:assign_project_user_role", map[string]string{"project_id": projectID}); !ok {
		return
	}
	if err := h.assignProjectUserRole(projectID, userID, roleID); err != nil {
		if errors.Is(err, errAssignRefsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteProjectUserRole(c *gin.Context) {
	projectID := c.Param("project_id")
	userID := c.Param("user_id")
	roleID := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_project_user_role", map[string]string{"project_id": projectID}); !ok {
		return
	}
	res := h.DB.Where("project_id = ? AND user_id = ? AND role_id = ?", projectID, userID, roleID).Delete(&models.UserProjectRole{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Role assignment not found."}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) assignProjectUserRole(projectID, userID, roleID string) error {
	var p models.Project
	if err := h.DB.Where("id = ?", projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errAssignRefsNotFound
		}
		return err
	}
	var u models.User
	if err := h.DB.Where("id = ?", userID).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errAssignRefsNotFound
		}
		return err
	}
	var r models.Role
	if err := h.DB.Where("id = ?", roleID).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errAssignRefsNotFound
		}
		return err
	}
	var existing models.UserProjectRole
	err := h.DB.Where("user_id = ? AND project_id = ? AND role_id = ?", userID, projectID, roleID).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return h.DB.Create(&models.UserProjectRole{UserID: userID, ProjectID: projectID, RoleID: roleID}).Error
}

func roleAssignmentRef(c *gin.Context, projectID, userID, roleID, roleName string) map[string]any {
	return map[string]any{
		"scope": map[string]any{
			"project": map[string]any{"id": projectID},
		},
		"role": map[string]any{
			"id":   roleID,
			"name": roleName,
		},
		"user": map[string]any{"id": userID},
	}
}

func roleAssignmentRefDomain(c *gin.Context, domainID, userID, roleID, roleName string) map[string]any {
	return map[string]any{
		"scope": map[string]any{
			"domain": map[string]any{"id": domainID},
		},
		"role": map[string]any{
			"id":   roleID,
			"name": roleName,
		},
		"user": map[string]any{"id": userID},
	}
}

func (h *Hub) createRoleAssignment(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_role_assignment", nil); !ok {
		return
	}
	var body struct {
		RoleAssignment struct {
			Scope struct {
				Project struct {
					ID string `json:"id"`
				} `json:"project"`
				Domain struct {
					ID string `json:"id"`
				} `json:"domain"`
			} `json:"scope"`
			User struct {
				ID string `json:"id"`
			} `json:"user"`
			Role struct {
				ID string `json:"id"`
			} `json:"role"`
		} `json:"role_assignment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	did := body.RoleAssignment.Scope.Domain.ID
	pid := body.RoleAssignment.Scope.Project.ID
	uid := body.RoleAssignment.User.ID
	rid := body.RoleAssignment.Role.ID
	if uid == "" || rid == "" || (did == "" && pid == "") {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	if did != "" && pid != "" {
		httperr.BadRequest(c, "Specify either domain or project scope, not both.")
		return
	}
	var r models.Role
	if did != "" {
		if err := h.assignDomainUserRole(did, uid, rid); err != nil {
			if errors.Is(err, errAssignRefsNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": err.Error()}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		if err := h.DB.Where("id = ?", rid).First(&r).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"role_assignment": roleAssignmentRefDomain(c, did, uid, rid, r.Name)})
		return
	}
	if err := h.assignProjectUserRole(pid, uid, rid); err != nil {
		if errors.Is(err, errAssignRefsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if err := h.DB.Where("id = ?", rid).First(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"role_assignment": roleAssignmentRef(c, pid, uid, rid, r.Name)})
}
