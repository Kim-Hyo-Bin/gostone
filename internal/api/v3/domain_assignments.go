package v3

import (
	"errors"
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Hub) listDomainUserRoles(c *gin.Context) {
	did := c.Param("domain_id")
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_domain_user_roles", map[string]string{"domain_id": did}); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	err := h.DB.Table("user_domain_roles").
		Select("user_domain_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = user_domain_roles.role_id").
		Where("user_domain_roles.domain_id = ? AND user_domain_roles.user_id = ?", did, uid).
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

func (h *Hub) putDomainUserRole(c *gin.Context) {
	did := c.Param("domain_id")
	uid := c.Param("user_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:assign_domain_user_role", map[string]string{"domain_id": did}); !ok {
		return
	}
	if err := h.assignDomainUserRole(did, uid, rid); err != nil {
		if errors.Is(err, errAssignRefsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteDomainUserRole(c *gin.Context) {
	did := c.Param("domain_id")
	uid := c.Param("user_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_domain_user_role", map[string]string{"domain_id": did}); !ok {
		return
	}
	res := h.DB.Where("domain_id = ? AND user_id = ? AND role_id = ?", did, uid, rid).Delete(&models.UserDomainRole{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) assignDomainUserRole(domainID, userID, roleID string) error {
	var d models.Domain
	if err := h.DB.Where("id = ?", domainID).First(&d).Error; err != nil {
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
	var ex models.UserDomainRole
	if err := h.DB.Where("user_id = ? AND domain_id = ? AND role_id = ?", userID, domainID, roleID).First(&ex).Error; err == nil {
		return nil
	}
	return h.DB.Create(&models.UserDomainRole{UserID: userID, DomainID: domainID, RoleID: roleID}).Error
}

func (h *Hub) listDomainGroupRoles(c *gin.Context) {
	did := c.Param("domain_id")
	gid := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_domain_group_roles", map[string]string{"domain_id": did}); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	err := h.DB.Table("group_domain_roles").
		Select("group_domain_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = group_domain_roles.role_id").
		Where("group_domain_roles.domain_id = ? AND group_domain_roles.group_id = ?", did, gid).
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

func (h *Hub) putDomainGroupRole(c *gin.Context) {
	did := c.Param("domain_id")
	gid := c.Param("group_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:assign_domain_group_role", map[string]string{"domain_id": did}); !ok {
		return
	}
	if err := h.assignDomainGroupRole(did, gid, rid); err != nil {
		if errors.Is(err, errAssignRefsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteDomainGroupRole(c *gin.Context) {
	did := c.Param("domain_id")
	gid := c.Param("group_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_domain_group_role", map[string]string{"domain_id": did}); !ok {
		return
	}
	res := h.DB.Where("domain_id = ? AND group_id = ? AND role_id = ?", did, gid, rid).Delete(&models.GroupDomainRole{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) assignDomainGroupRole(domainID, groupID, roleID string) error {
	var d models.Domain
	if err := h.DB.Where("id = ?", domainID).First(&d).Error; err != nil {
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
	var ex models.GroupDomainRole
	if err := h.DB.Where("group_id = ? AND domain_id = ? AND role_id = ?", groupID, domainID, roleID).First(&ex).Error; err == nil {
		return nil
	}
	return h.DB.Create(&models.GroupDomainRole{GroupID: groupID, DomainID: domainID, RoleID: roleID}).Error
}
