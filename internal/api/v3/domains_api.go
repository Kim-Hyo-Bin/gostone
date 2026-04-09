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

func registerV3DomainsAPI(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/domains")
	g.GET("/:domain_id/users/:user_id/roles", h.listDomainUserRoles)
	g.PUT("/:domain_id/users/:user_id/roles/:role_id", h.putDomainUserRole)
	g.DELETE("/:domain_id/users/:user_id/roles/:role_id", h.deleteDomainUserRole)
	g.GET("/:domain_id/groups/:group_id/roles", h.listDomainGroupRoles)
	g.PUT("/:domain_id/groups/:group_id/roles/:role_id", h.putDomainGroupRole)
	g.DELETE("/:domain_id/groups/:group_id/roles/:role_id", h.deleteDomainGroupRole)

	registerV3DomainConfigRoutes(g, h)

	g.GET("", h.listDomains)
	g.POST("", h.createDomain)
	g.GET("/:domain_id", h.getDomain)
	g.HEAD("/:domain_id", h.headDomain)
	g.PATCH("/:domain_id", h.patchDomain)
	g.DELETE("/:domain_id", h.deleteDomain)
}

func (h *Hub) listDomains(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_domains", nil); !ok {
		return
	}
	q := h.DB.Model(&models.Domain{}).Order("name")
	if n := strings.TrimSpace(c.Query("name")); n != "" {
		q = q.Where("name = ?", n)
	}
	var list []models.Domain
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, d := range list {
		out = append(out, domainRef(c, d))
	}
	c.JSON(http.StatusOK, gin.H{"domains": out, "links": gin.H{"self": selfURL(c, "/v3/domains")}})
}

type domainJSON struct {
	Domain struct {
		Name    string `json:"name"`
		Enabled *bool  `json:"enabled"`
	} `json:"domain"`
}

func (h *Hub) createDomain(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_domain", nil); !ok {
		return
	}
	var body domainJSON
	if err := c.ShouldBindJSON(&body); err != nil || body.Domain.Name == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	en := true
	if body.Domain.Enabled != nil {
		en = *body.Domain.Enabled
	}
	d := models.Domain{ID: uuid.NewString(), Name: body.Domain.Name, Enabled: en}
	if err := h.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"domain": domainRef(c, d)})
}

func (h *Hub) getDomain(c *gin.Context) {
	id := c.Param("domain_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain", map[string]string{"domain_id": id}); !ok {
		return
	}
	var d models.Domain
	if err := h.DB.Where("id = ?", id).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find domain: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"domain": domainRef(c, d)})
}

func (h *Hub) headDomain(c *gin.Context) {
	id := c.Param("domain_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain", map[string]string{"domain_id": id}); !ok {
		return
	}
	var d models.Domain
	if err := h.DB.Where("id = ?", id).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find domain: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Hub) patchDomain(c *gin.Context) {
	id := c.Param("domain_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_domain", map[string]string{"domain_id": id}); !ok {
		return
	}
	var body domainJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var d models.Domain
	if err := h.DB.Where("id = ?", id).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find domain: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Domain.Name != "" {
		d.Name = body.Domain.Name
	}
	if body.Domain.Enabled != nil {
		d.Enabled = *body.Domain.Enabled
	}
	if err := h.DB.Save(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"domain": domainRef(c, d)})
}

func (h *Hub) deleteDomain(c *gin.Context) {
	id := c.Param("domain_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_domain", map[string]string{"domain_id": id}); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.Domain{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find domain: " + id}})
		return
	}
	c.Status(http.StatusNoContent)
}

func domainRef(c *gin.Context, d models.Domain) map[string]any {
	return map[string]any{
		"id":          d.ID,
		"name":        d.Name,
		"enabled":     d.Enabled,
		"links":       gin.H{"self": selfURL(c, "/v3/domains/"+d.ID)},
		"description": nil,
	}
}
