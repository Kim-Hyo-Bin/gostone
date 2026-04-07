package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func registerV3GroupsAPI(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/groups")
	g.GET("", h.listGroups)
	g.POST("", h.createGroup)
	g.GET("/:group_id", h.getGroup)
	g.HEAD("/:group_id", h.headGroup)
	g.PATCH("/:group_id", h.patchGroup)
	g.PUT("/:group_id", h.patchGroup)
	g.DELETE("/:group_id", h.deleteGroup)

	g.GET("/:group_id/users", h.listGroupUsers)
	g.PUT("/:group_id/users/:user_id", h.addGroupUser)
	g.DELETE("/:group_id/users/:user_id", h.removeGroupUser)
}

func (h *Hub) listGroups(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_groups", nil); !ok {
		return
	}
	q := h.DB.Model(&models.Group{}).Order("name")
	if d := strings.TrimSpace(c.Query("domain_id")); d != "" {
		q = q.Where("domain_id = ?", d)
	}
	var list []models.Group
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, g := range list {
		out = append(out, groupRef(c, g))
	}
	c.JSON(http.StatusOK, gin.H{"groups": out, "links": gin.H{"self": selfURL(c, "/v3/groups")}})
}

type groupBody struct {
	Group struct {
		Name        string `json:"name"`
		DomainID    string `json:"domain_id"`
		Description string `json:"description"`
	} `json:"group"`
}

func (h *Hub) createGroup(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_group", nil); !ok {
		return
	}
	var body groupBody
	if err := c.ShouldBindJSON(&body); err != nil || body.Group.Name == "" || body.Group.DomainID == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	g := models.Group{ID: uuid.NewString(), Name: body.Group.Name, DomainID: body.Group.DomainID, Description: body.Group.Description}
	if err := h.DB.Create(&g).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"group": groupRef(c, g)})
}

func (h *Hub) getGroup(c *gin.Context) {
	id := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_group", map[string]string{"group_id": id}); !ok {
		return
	}
	var g models.Group
	if err := h.DB.Where("id = ?", id).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find group: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"group": groupRef(c, g)})
}

func (h *Hub) headGroup(c *gin.Context) {
	id := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_group", map[string]string{"group_id": id}); !ok {
		return
	}
	var g models.Group
	if err := h.DB.Where("id = ?", id).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find group: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Hub) patchGroup(c *gin.Context) {
	id := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_group", map[string]string{"group_id": id}); !ok {
		return
	}
	var body groupBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var g models.Group
	if err := h.DB.Where("id = ?", id).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find group: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Group.Name != "" {
		g.Name = body.Group.Name
	}
	if body.Group.DomainID != "" {
		g.DomainID = body.Group.DomainID
	}
	if body.Group.Description != "" {
		g.Description = body.Group.Description
	}
	if err := h.DB.Save(&g).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"group": groupRef(c, g)})
}

func (h *Hub) deleteGroup(c *gin.Context) {
	id := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_group", map[string]string{"group_id": id}); !ok {
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupProjectRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupDomainRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupSystemRole{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", id).Delete(&models.Group{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find group: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) listGroupUsers(c *gin.Context) {
	gid := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_group_users", map[string]string{"group_id": gid}); !ok {
		return
	}
	var uids []string
	if err := h.DB.Model(&models.GroupMember{}).Where("group_id = ?", gid).Pluck("user_id", &uids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	refs := make([]map[string]any, 0, len(uids))
	for _, id := range uids {
		refs = append(refs, map[string]any{"id": id})
	}
	c.JSON(http.StatusOK, gin.H{"users": refs})
}

func (h *Hub) addGroupUser(c *gin.Context) {
	gid := c.Param("group_id")
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_user_to_group", map[string]string{"group_id": gid}); !ok {
		return
	}
	var g models.Group
	if err := h.DB.Where("id = ?", gid).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find group"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	var u models.User
	if err := h.DB.Where("id = ?", uid).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find user"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	var ex models.GroupMember
	if err := h.DB.Where("group_id = ? AND user_id = ?", gid, uid).First(&ex).Error; err == nil {
		c.Status(http.StatusNoContent)
		return
	}
	if err := h.DB.Create(&models.GroupMember{GroupID: gid, UserID: uid}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) removeGroupUser(c *gin.Context) {
	gid := c.Param("group_id")
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_user_from_group", map[string]string{"group_id": gid}); !ok {
		return
	}
	res := h.DB.Where("group_id = ? AND user_id = ?", gid, uid).Delete(&models.GroupMember{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "User not in group"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func groupRef(c *gin.Context, g models.Group) map[string]any {
	m := map[string]any{
		"id":          g.ID,
		"name":        g.Name,
		"domain_id":   g.DomainID,
		"links":       gin.H{"self": selfURL(c, "/v3/groups/"+g.ID)},
		"description": nil,
	}
	if g.Description != "" {
		m["description"] = g.Description
	}
	return m
}
