package v3

import (
	"errors"
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Hub) listGroupProjectRoles(c *gin.Context) {
	pid := c.Param("project_id")
	gid := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_group_project_roles", map[string]string{"project_id": pid}); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	err := h.DB.Table("group_project_roles").
		Select("group_project_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = group_project_roles.role_id").
		Where("group_project_roles.project_id = ? AND group_project_roles.group_id = ?", pid, gid).
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"id": r.RoleID, "name": r.RoleName})
	}
	c.JSON(http.StatusOK, gin.H{"roles": out})
}

func (h *Hub) putGroupProjectRole(c *gin.Context) {
	pid := c.Param("project_id")
	gid := c.Param("group_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:assign_group_project_role", map[string]string{"project_id": pid}); !ok {
		return
	}
	if err := h.assignGroupProjectRole(pid, gid, rid); err != nil {
		if errors.Is(err, errAssignRefsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteGroupProjectRole(c *gin.Context) {
	pid := c.Param("project_id")
	gid := c.Param("group_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_group_project_role", map[string]string{"project_id": pid}); !ok {
		return
	}
	res := h.DB.Where("project_id = ? AND group_id = ? AND role_id = ?", pid, gid, rid).Delete(&models.GroupProjectRole{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) assignGroupProjectRole(projectID, groupID, roleID string) error {
	var p models.Project
	if err := h.DB.Where("id = ?", projectID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errAssignRefsNotFound
		}
		return err
	}
	var g models.Group
	if err := h.DB.Where("id = ?", groupID).First(&g).Error; err != nil {
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
	var ex models.GroupProjectRole
	if err := h.DB.Where("group_id = ? AND project_id = ? AND role_id = ?", groupID, projectID, roleID).First(&ex).Error; err == nil {
		return nil
	}
	return h.DB.Create(&models.GroupProjectRole{GroupID: groupID, ProjectID: projectID, RoleID: roleID}).Error
}
