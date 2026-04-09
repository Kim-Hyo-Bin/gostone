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

func registerV3CatalogAPI(v3 *gin.RouterGroup, h *Hub) {
	reg := v3.Group("/regions")
	reg.GET("", h.listRegions)
	reg.POST("", h.createRegion)
	reg.GET("/:region_id", h.getRegion)
	reg.HEAD("/:region_id", h.headRegion)
	reg.PATCH("/:region_id", h.patchRegion)
	reg.PUT("/:region_id", h.patchRegion)
	reg.DELETE("/:region_id", h.deleteRegion)

	svc := v3.Group("/services")
	svc.GET("", h.listServices)
	svc.POST("", h.createService)
	svc.GET("/:service_id", h.getService)
	svc.HEAD("/:service_id", h.headService)
	svc.PATCH("/:service_id", h.patchService)
	svc.PUT("/:service_id", h.patchService)
	svc.DELETE("/:service_id", h.deleteService)

	ep := v3.Group("/endpoints")
	ep.GET("", h.listEndpoints)
	ep.POST("", h.createEndpoint)
	ep.GET("/:endpoint_id", h.getEndpoint)
	ep.HEAD("/:endpoint_id", h.headEndpoint)
	ep.PATCH("/:endpoint_id", h.patchEndpoint)
	ep.PUT("/:endpoint_id", h.patchEndpoint)
	ep.DELETE("/:endpoint_id", h.deleteEndpoint)
}

// --- regions ---

func (h *Hub) listRegions(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_regions", nil); !ok {
		return
	}
	var list []models.Region
	if err := h.DB.Order("id").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, r := range list {
		out = append(out, regionRef(c, r))
	}
	c.JSON(http.StatusOK, gin.H{"regions": out, "links": gin.H{"self": selfURL(c, "/v3/regions")}})
}

type regionJSON struct {
	Region struct {
		ID             string `json:"id"`
		Description    string `json:"description"`
		ParentRegionID string `json:"parent_region_id"`
	} `json:"region"`
}

func (h *Hub) createRegion(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_region", nil); !ok {
		return
	}
	var body regionJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	id := strings.TrimSpace(body.Region.ID)
	if id == "" {
		id = uuid.NewString()
	}
	r := models.Region{
		ID:          id,
		Description: body.Region.Description,
		ParentID:    body.Region.ParentRegionID,
	}
	if err := h.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"region": regionRef(c, r)})
}

func (h *Hub) getRegion(c *gin.Context) {
	id := c.Param("region_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_region", map[string]string{"region_id": id}); !ok {
		return
	}
	var r models.Region
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find region: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"region": regionRef(c, r)})
}

func (h *Hub) headRegion(c *gin.Context) {
	id := c.Param("region_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_region", map[string]string{"region_id": id}); !ok {
		return
	}
	var r models.Region
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find region: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Hub) patchRegion(c *gin.Context) {
	id := c.Param("region_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_region", map[string]string{"region_id": id}); !ok {
		return
	}
	var body regionJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var r models.Region
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find region: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Region.Description != "" {
		r.Description = body.Region.Description
	}
	if body.Region.ParentRegionID != "" {
		r.ParentID = body.Region.ParentRegionID
	}
	if err := h.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"region": regionRef(c, r)})
}

func (h *Hub) deleteRegion(c *gin.Context) {
	id := c.Param("region_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_region", map[string]string{"region_id": id}); !ok {
		return
	}
	var n int64
	if err := h.DB.Model(&models.Endpoint{}).Where("region_id = ?", id).Count(&n).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if n > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": "Region still referenced by endpoints."}})
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.Region{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find region: " + id}})
		return
	}
	c.Status(http.StatusNoContent)
}

func regionRef(c *gin.Context, r models.Region) map[string]any {
	m := map[string]any{
		"id":          r.ID,
		"links":       gin.H{"self": selfURL(c, "/v3/regions/"+r.ID)},
		"description": nil,
	}
	if r.Description != "" {
		m["description"] = r.Description
	}
	if r.ParentID != "" {
		m["parent_region_id"] = r.ParentID
	}
	return m
}

// --- services ---

func (h *Hub) listServices(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_services", nil); !ok {
		return
	}
	q := h.DB.Model(&models.Service{}).Order("type, name")
	if v := strings.TrimSpace(c.Query("type")); v != "" {
		q = q.Where("type = ?", v)
	}
	if v := strings.TrimSpace(c.Query("name")); v != "" {
		q = q.Where("name = ?", v)
	}
	var list []models.Service
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, s := range list {
		out = append(out, serviceRef(c, s))
	}
	c.JSON(http.StatusOK, gin.H{"services": out, "links": gin.H{"self": selfURL(c, "/v3/services")}})
}

type serviceJSON struct {
	Service struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     *bool  `json:"enabled"`
	} `json:"service"`
}

func (h *Hub) createService(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_service", nil); !ok {
		return
	}
	var body serviceJSON
	if err := c.ShouldBindJSON(&body); err != nil || body.Service.Type == "" || body.Service.Name == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	en := true
	if body.Service.Enabled != nil {
		en = *body.Service.Enabled
	}
	s := models.Service{
		ID:          uuid.NewString(),
		Type:        body.Service.Type,
		Name:        body.Service.Name,
		Description: body.Service.Description,
		Enabled:     en,
	}
	if err := h.DB.Create(&s).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"service": serviceRef(c, s)})
}

func (h *Hub) getService(c *gin.Context) {
	id := c.Param("service_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_service", map[string]string{"service_id": id}); !ok {
		return
	}
	var s models.Service
	if err := h.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find service: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": serviceRef(c, s)})
}

func (h *Hub) headService(c *gin.Context) {
	id := c.Param("service_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_service", map[string]string{"service_id": id}); !ok {
		return
	}
	var s models.Service
	if err := h.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find service: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Hub) patchService(c *gin.Context) {
	id := c.Param("service_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_service", map[string]string{"service_id": id}); !ok {
		return
	}
	var body serviceJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var s models.Service
	if err := h.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find service: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Service.Type != "" {
		s.Type = body.Service.Type
	}
	if body.Service.Name != "" {
		s.Name = body.Service.Name
	}
	if body.Service.Description != "" {
		s.Description = body.Service.Description
	}
	if body.Service.Enabled != nil {
		s.Enabled = *body.Service.Enabled
	}
	if err := h.DB.Save(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": serviceRef(c, s)})
}

func (h *Hub) deleteService(c *gin.Context) {
	id := c.Param("service_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_service", map[string]string{"service_id": id}); !ok {
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("service_id = ?", id).Delete(&models.Endpoint{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", id).Delete(&models.Service{})
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
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find service: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func serviceRef(c *gin.Context, s models.Service) map[string]any {
	m := map[string]any{
		"id":      s.ID,
		"type":    s.Type,
		"name":    s.Name,
		"enabled": s.Enabled,
		"links":   gin.H{"self": selfURL(c, "/v3/services/"+s.ID)},
	}
	if s.Description != "" {
		m["description"] = s.Description
	}
	return m
}

// --- endpoints ---

func (h *Hub) listEndpoints(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_endpoints", nil); !ok {
		return
	}
	q := h.DB.Model(&models.Endpoint{}).Order("service_id, interface")
	if v := strings.TrimSpace(c.Query("service_id")); v != "" {
		q = q.Where("service_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("region_id")); v != "" {
		q = q.Where("region_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("interface")); v != "" {
		q = q.Where("interface = ?", v)
	}
	var list []models.Endpoint
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, e := range list {
		out = append(out, endpointRef(c, e))
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": out, "links": gin.H{"self": selfURL(c, "/v3/endpoints")}})
}

type endpointJSON struct {
	Endpoint struct {
		ServiceID string `json:"service_id"`
		RegionID  string `json:"region_id"`
		Interface string `json:"interface"`
		URL       string `json:"url"`
		Enabled   *bool  `json:"enabled"`
	} `json:"endpoint"`
}

func (h *Hub) createEndpoint(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_endpoint", nil); !ok {
		return
	}
	var body endpointJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	ep := body.Endpoint
	if ep.ServiceID == "" || ep.RegionID == "" || ep.Interface == "" || ep.URL == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var svc models.Service
	if err := h.DB.Where("id = ?", ep.ServiceID).First(&svc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": 400, "message": "Invalid service id."}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	var reg models.Region
	if err := h.DB.Where("id = ?", ep.RegionID).First(&reg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": 400, "message": "Invalid region id."}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	en := true
	if ep.Enabled != nil {
		en = *ep.Enabled
	}
	e := models.Endpoint{
		ID:        uuid.NewString(),
		ServiceID: ep.ServiceID,
		RegionID:  ep.RegionID,
		Interface: ep.Interface,
		URL:       ep.URL,
		Enabled:   en,
	}
	if err := h.DB.Create(&e).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"endpoint": endpointRef(c, e)})
}

func (h *Hub) getEndpoint(c *gin.Context) {
	id := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_endpoint", map[string]string{"endpoint_id": id}); !ok {
		return
	}
	var e models.Endpoint
	if err := h.DB.Where("id = ?", id).First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find endpoint: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"endpoint": endpointRef(c, e)})
}

func (h *Hub) headEndpoint(c *gin.Context) {
	id := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_endpoint", map[string]string{"endpoint_id": id}); !ok {
		return
	}
	var e models.Endpoint
	if err := h.DB.Where("id = ?", id).First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find endpoint: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Hub) patchEndpoint(c *gin.Context) {
	id := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_endpoint", map[string]string{"endpoint_id": id}); !ok {
		return
	}
	var body endpointJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var e models.Endpoint
	if err := h.DB.Where("id = ?", id).First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find endpoint: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	ep := body.Endpoint
	if ep.ServiceID != "" {
		var svc models.Service
		if err := h.DB.Where("id = ?", ep.ServiceID).First(&svc).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": 400, "message": "Invalid service id."}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		e.ServiceID = ep.ServiceID
	}
	if ep.RegionID != "" {
		var reg models.Region
		if err := h.DB.Where("id = ?", ep.RegionID).First(&reg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": 400, "message": "Invalid region id."}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		e.RegionID = ep.RegionID
	}
	if ep.Interface != "" {
		e.Interface = ep.Interface
	}
	if ep.URL != "" {
		e.URL = ep.URL
	}
	if ep.Enabled != nil {
		e.Enabled = *ep.Enabled
	}
	if err := h.DB.Save(&e).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"endpoint": endpointRef(c, e)})
}

func (h *Hub) deleteEndpoint(c *gin.Context) {
	id := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_endpoint", map[string]string{"endpoint_id": id}); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.Endpoint{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find endpoint: " + id}})
		return
	}
	c.Status(http.StatusNoContent)
}

func endpointRef(c *gin.Context, e models.Endpoint) map[string]any {
	return map[string]any{
		"id":         e.ID,
		"service_id": e.ServiceID,
		"region_id":  e.RegionID,
		"interface":  e.Interface,
		"url":        e.URL,
		"enabled":    e.Enabled,
		"links":      gin.H{"self": selfURL(c, "/v3/endpoints/"+e.ID)},
	}
}
