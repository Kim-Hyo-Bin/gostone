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

func (h *Hub) getSAML2Metadata(c *gin.Context) {
	c.Data(http.StatusOK, "application/xml", []byte(`<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="gostone"/>`))
}

func (h *Hub) fedAuthPlaceholder(c *gin.Context) {
	httperr.NotImplemented(c, "Federation protocol auth exchange not implemented.")
}

// --- identity providers (override stub CRUD by registering real handlers on sub-router) ---
func registerV3FederationCRUD(fed *gin.RouterGroup, h *Hub) {
	idp := fed.Group("/identity_providers")
	idp.GET("", h.listIDPs)
	idp.POST("", h.createIDP)

	pr := fed.Group("/identity_providers/:idp_id/protocols")
	pr.GET("", h.listIDPProtocols)
	pr.GET("/:protocol_id", h.getIDPProtocol)

	idp.GET("/:idp_id", h.getIDP)
	idp.PATCH("/:idp_id", h.patchIDP)
	idp.PUT("/:idp_id", h.patchIDP)
	idp.DELETE("/:idp_id", h.deleteIDP)

	mp := fed.Group("/mappings")
	mp.GET("", h.listMappings)
	mp.POST("", h.createMapping)
	mp.GET("/:mapping_id", h.getMapping)
	mp.PATCH("/:mapping_id", h.patchMapping)
	mp.PUT("/:mapping_id", h.patchMapping)
	mp.DELETE("/:mapping_id", h.deleteMapping)

	sp := fed.Group("/service_providers")
	sp.GET("", h.listSPs)
	sp.POST("", h.createSP)
	sp.GET("/:service_provider_id", h.getSP)
	sp.PATCH("/:service_provider_id", h.patchSP)
	sp.PUT("/:service_provider_id", h.patchSP)
	sp.DELETE("/:service_provider_id", h.deleteSP)
}

func (h *Hub) listIDPs(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_identity_providers", nil); !ok {
		return
	}
	var list []models.IdentityProvider
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, idpRef(c, x))
	}
	c.JSON(http.StatusOK, gin.H{"identity_providers": out})
}

func (h *Hub) createIDP(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_identity_provider", nil); !ok {
		return
	}
	var body struct {
		IdentityProvider struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Enabled     *bool  `json:"enabled"`
		} `json:"identity_provider"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	id := body.IdentityProvider.ID
	if id == "" {
		id = uuid.NewString()
	}
	en := true
	if body.IdentityProvider.Enabled != nil {
		en = *body.IdentityProvider.Enabled
	}
	x := models.IdentityProvider{ID: id, Description: body.IdentityProvider.Description, Enabled: en}
	if err := h.DB.Create(&x).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"identity_provider": idpRef(c, x)})
}

func idpRef(c *gin.Context, x models.IdentityProvider) map[string]any {
	return map[string]any{
		"id":          x.ID,
		"enabled":     x.Enabled,
		"description": x.Description,
		"remote_ids":  []any{},
		"links":       gin.H{"self": selfURL(c, "/v3/OS-FEDERATION/identity_providers/"+x.ID)},
	}
}

func (h *Hub) getIDP(c *gin.Context) {
	id := c.Param("idp_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_identity_provider", nil); !ok {
		return
	}
	var x models.IdentityProvider
	if err := h.DB.Where("id = ?", id).First(&x).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"identity_provider": idpRef(c, x)})
}

func (h *Hub) patchIDP(c *gin.Context) {
	id := c.Param("idp_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_identity_provider", nil); !ok {
		return
	}
	var body struct {
		IdentityProvider struct {
			Description string `json:"description"`
			Enabled     *bool  `json:"enabled"`
		} `json:"identity_provider"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var x models.IdentityProvider
	if err := h.DB.Where("id = ?", id).First(&x).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.IdentityProvider.Description != "" {
		x.Description = body.IdentityProvider.Description
	}
	if body.IdentityProvider.Enabled != nil {
		x.Enabled = *body.IdentityProvider.Enabled
	}
	if err := h.DB.Save(&x).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"identity_provider": idpRef(c, x)})
}

func (h *Hub) deleteIDP(c *gin.Context) {
	id := c.Param("idp_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_identity_provider", nil); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.IdentityProvider{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) listIDPProtocols(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_identity_provider", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"protocols": []map[string]any{{"id": "saml2"}}})
}

func (h *Hub) getIDPProtocol(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_identity_provider", nil); !ok {
		return
	}
	pid := c.Param("protocol_id")
	c.JSON(http.StatusOK, gin.H{"protocol": gin.H{"id": pid}})
}

func (h *Hub) listMappings(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_mappings", nil); !ok {
		return
	}
	var list []models.FederationMapping
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, m := range list {
		out = append(out, map[string]any{"id": m.ID, "rules": m.Rules, "links": gin.H{"self": selfURL(c, "/v3/OS-FEDERATION/mappings/"+m.ID)}})
	}
	c.JSON(http.StatusOK, gin.H{"mappings": out})
}

func (h *Hub) createMapping(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_mapping", nil); !ok {
		return
	}
	var body struct {
		Mapping struct {
			Rules string `json:"rules"`
		} `json:"mapping"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Mapping.Rules) == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	m := models.FederationMapping{ID: uuid.NewString(), Rules: body.Mapping.Rules}
	if err := h.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"mapping": gin.H{"id": m.ID, "rules": m.Rules}})
}

func (h *Hub) getMapping(c *gin.Context) {
	id := c.Param("mapping_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_mapping", nil); !ok {
		return
	}
	var m models.FederationMapping
	if err := h.DB.Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mapping": gin.H{"id": m.ID, "rules": m.Rules}})
}

func (h *Hub) patchMapping(c *gin.Context) {
	id := c.Param("mapping_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_mapping", nil); !ok {
		return
	}
	var body struct {
		Mapping struct {
			Rules string `json:"rules"`
		} `json:"mapping"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var m models.FederationMapping
	if err := h.DB.Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Mapping.Rules != "" {
		m.Rules = body.Mapping.Rules
	}
	if err := h.DB.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mapping": gin.H{"id": m.ID, "rules": m.Rules}})
}

func (h *Hub) deleteMapping(c *gin.Context) {
	id := c.Param("mapping_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_mapping", nil); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.FederationMapping{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) listSPs(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_service_providers", nil); !ok {
		return
	}
	var list []models.ServiceProvider
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, s := range list {
		out = append(out, spRef(c, s))
	}
	c.JSON(http.StatusOK, gin.H{"service_providers": out})
}

func spRef(c *gin.Context, s models.ServiceProvider) map[string]any {
	return map[string]any{
		"id":          s.ID,
		"description": s.Description,
		"auth_url":    s.AuthURL,
		"sp_url":      s.SpURL,
		"enabled":     s.Enabled,
		"links":       gin.H{"self": selfURL(c, "/v3/OS-FEDERATION/service_providers/"+s.ID)},
	}
}

func (h *Hub) createSP(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_service_provider", nil); !ok {
		return
	}
	var body struct {
		ServiceProvider struct {
			Description string `json:"description"`
			AuthURL     string `json:"auth_url"`
			SpURL       string `json:"sp_url"`
		} `json:"service_provider"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	s := models.ServiceProvider{ID: uuid.NewString(), Description: body.ServiceProvider.Description, AuthURL: body.ServiceProvider.AuthURL, SpURL: body.ServiceProvider.SpURL, Enabled: true}
	if err := h.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"service_provider": spRef(c, s)})
}

func (h *Hub) getSP(c *gin.Context) {
	id := c.Param("service_provider_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_service_provider", nil); !ok {
		return
	}
	var s models.ServiceProvider
	if err := h.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service_provider": spRef(c, s)})
}

func (h *Hub) patchSP(c *gin.Context) {
	id := c.Param("service_provider_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_service_provider", nil); !ok {
		return
	}
	var body struct {
		ServiceProvider struct {
			Description string `json:"description"`
			Enabled     *bool  `json:"enabled"`
		} `json:"service_provider"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var s models.ServiceProvider
	if err := h.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.ServiceProvider.Description != "" {
		s.Description = body.ServiceProvider.Description
	}
	if body.ServiceProvider.Enabled != nil {
		s.Enabled = *body.ServiceProvider.Enabled
	}
	if err := h.DB.Save(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service_provider": spRef(c, s)})
}

func (h *Hub) deleteSP(c *gin.Context) {
	id := c.Param("service_provider_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_service_provider", nil); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.ServiceProvider{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}
