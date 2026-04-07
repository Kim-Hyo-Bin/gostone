package v3

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func registerV3RolesAPI(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/roles")
	g.GET("", h.listRoles)
	g.POST("", h.createRole)
	g.GET("/:role_id", h.getRole)
	g.HEAD("/:role_id", h.stubRoute("HEAD /v3/roles/:role_id"))
	g.PATCH("/:role_id", h.patchRole)
	g.PUT("/:role_id", h.stubRoute("PUT /v3/roles/:role_id"))
	g.DELETE("/:role_id", h.deleteRole)
}

func (h *Hub) listRoles(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:list_roles", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var list []models.Role
	if err := h.DB.Order("name").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, r := range list {
		out = append(out, roleRef(c, r))
	}
	c.JSON(http.StatusOK, gin.H{"roles": out, "links": gin.H{"self": selfURL(c, "/v3/roles")}})
}

type roleJSON struct {
	Role struct {
		Name     string `json:"name"`
		DomainID string `json:"domain_id"`
	} `json:"role"`
}

func (h *Hub) createRole(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:create_role", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var body roleJSON
	if err := c.ShouldBindJSON(&body); err != nil || body.Role.Name == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	r := models.Role{ID: uuid.NewString(), Name: body.Role.Name, DomainID: body.Role.DomainID}
	if err := h.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"role": roleRef(c, r)})
}

func (h *Hub) getRole(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("role_id")
	if !h.Policy.Allow("identity:get_role", actx, map[string]string{"role_id": id}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var r models.Role
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find role: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": roleRef(c, r)})
}

func (h *Hub) patchRole(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("role_id")
	if !h.Policy.Allow("identity:update_role", actx, map[string]string{"role_id": id}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var body roleJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var r models.Role
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find role: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Role.Name != "" {
		r.Name = body.Role.Name
	}
	if body.Role.DomainID != "" {
		r.DomainID = body.Role.DomainID
	}
	if err := h.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": roleRef(c, r)})
}

func (h *Hub) deleteRole(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("role_id")
	if !h.Policy.Allow("identity:delete_role", actx, map[string]string{"role_id": id}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	res := h.DB.Delete(&models.Role{}, "id = ?", id)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find role: " + id}})
		return
	}
	c.Status(http.StatusNoContent)
}

func roleRef(c *gin.Context, r models.Role) map[string]any {
	m := map[string]any{
		"id":    r.ID,
		"name":  r.Name,
		"links": gin.H{"self": selfURL(c, "/v3/roles/"+r.ID)},
	}
	if r.DomainID != "" {
		m["domain_id"] = r.DomainID
	}
	return m
}
