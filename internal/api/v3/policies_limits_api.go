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

func registerV3PoliciesBundle(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/policies")
	g.GET("/:policy_id/OS-ENDPOINT-POLICY/endpoints", h.listPolicyEndpoints)
	g.PUT("/:policy_id/OS-ENDPOINT-POLICY/endpoints/:endpoint_id", h.addPolicyEndpoint)
	g.DELETE("/:policy_id/OS-ENDPOINT-POLICY/endpoints/:endpoint_id", h.removePolicyEndpoint)
	g.PUT("/:policy_id/OS-ENDPOINT-POLICY/services/:service_id", h.addPolicyService)
	g.DELETE("/:policy_id/OS-ENDPOINT-POLICY/services/:service_id/regions/:region_id", h.removePolicyServiceRegion)

	g.GET("", h.listPolicyDocs)
	g.POST("", h.createPolicyDoc)
	g.GET("/:policy_id", h.getPolicyDoc)
	g.HEAD("/:policy_id", h.headPolicyDoc)
	g.PATCH("/:policy_id", h.patchPolicyDoc)
	g.PUT("/:policy_id", h.patchPolicyDoc)
	g.DELETE("/:policy_id", h.deletePolicyDoc)

	ep := v3.Group("/endpoints")
	ep.GET("/:endpoint_id/OS-ENDPOINT-POLICY/policy", h.getEndpointPolicyBinding)
	ep.PUT("/:endpoint_id/OS-ENDPOINT-POLICY/policy", h.setEndpointPolicyBinding)
	ep.DELETE("/:endpoint_id/OS-ENDPOINT-POLICY/policy", h.clearEndpointPolicyBinding)

	rl := v3.Group("/registered_limits")
	rl.GET("", h.listRegisteredLimits)
	rl.POST("", h.createRegisteredLimit)
	rl.GET("/:registered_limit_id", h.getRegisteredLimit)
	rl.PATCH("/:registered_limit_id", h.patchRegisteredLimit)
	rl.PUT("/:registered_limit_id", h.patchRegisteredLimit)
	rl.DELETE("/:registered_limit_id", h.deleteRegisteredLimit)

	lim := v3.Group("/limits")
	lim.GET("", h.listLimits)
	lim.POST("", h.createLimit)
	lim.GET("/:limit_id", h.getLimit)
	lim.PATCH("/:limit_id", h.patchLimit)
	lim.PUT("/:limit_id", h.patchLimit)
	lim.DELETE("/:limit_id", h.deleteLimit)

	v3.GET("/limits/model", h.getLimitsModel)
}

func (h *Hub) listPolicyEndpoints(c *gin.Context) {
	pid := c.Param("policy_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_policy_endpoints", map[string]string{"policy_id": pid}); !ok {
		return
	}
	var links []models.EndpointPolicyLink
	if err := h.DB.Where("policy_id = ?", pid).Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(links))
	for _, l := range links {
		out = append(out, map[string]any{"id": l.EndpointID})
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": out})
}

func (h *Hub) addPolicyEndpoint(c *gin.Context) {
	pid := c.Param("policy_id")
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_policy_endpoint", map[string]string{"policy_id": pid}); !ok {
		return
	}
	if err := h.DB.Create(&models.EndpointPolicyLink{EndpointID: eid, PolicyID: pid}).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) removePolicyEndpoint(c *gin.Context) {
	pid := c.Param("policy_id")
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_policy_endpoint", map[string]string{"policy_id": pid}); !ok {
		return
	}
	h.DB.Where("policy_id = ? AND endpoint_id = ?", pid, eid).Delete(&models.EndpointPolicyLink{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) addPolicyService(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:add_policy_service", nil); !ok {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) removePolicyServiceRegion(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:remove_policy_service_region", nil); !ok {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) getEndpointPolicyBinding(c *gin.Context) {
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_endpoint", map[string]string{"endpoint_id": eid}); !ok {
		return
	}
	var l models.EndpointPolicyLink
	if err := h.DB.Where("endpoint_id = ?", eid).First(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "No policy on endpoint"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policy": gin.H{"id": l.PolicyID}})
}

func (h *Hub) setEndpointPolicyBinding(c *gin.Context) {
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_policy_endpoint", nil); !ok {
		return
	}
	var body struct {
		Policy struct {
			ID string `json:"id"`
		} `json:"policy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Policy.ID == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	h.DB.Where("endpoint_id = ?", eid).Delete(&models.EndpointPolicyLink{})
	if err := h.DB.Create(&models.EndpointPolicyLink{EndpointID: eid, PolicyID: body.Policy.ID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) clearEndpointPolicyBinding(c *gin.Context) {
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_policy_endpoint", nil); !ok {
		return
	}
	h.DB.Where("endpoint_id = ?", eid).Delete(&models.EndpointPolicyLink{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listPolicyDocs(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_policies", nil); !ok {
		return
	}
	var list []models.IdentityPolicyDoc
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, policyDocRef(c, p))
	}
	c.JSON(http.StatusOK, gin.H{"policies": out, "links": gin.H{"self": selfURL(c, "/v3/policies")}})
}

type policyDocBody struct {
	Policy struct {
		Blob string `json:"blob"`
		Type string `json:"type"`
	} `json:"policy"`
}

func (h *Hub) createPolicyDoc(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_policy", nil); !ok {
		return
	}
	var body policyDocBody
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Policy.Blob) == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	p := models.IdentityPolicyDoc{ID: uuid.NewString(), Blob: body.Policy.Blob, Type: body.Policy.Type}
	if err := h.DB.Create(&p).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"policy": policyDocRef(c, p)})
}

func (h *Hub) getPolicyDoc(c *gin.Context) {
	id := c.Param("policy_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_policy", map[string]string{"policy_id": id}); !ok {
		return
	}
	var p models.IdentityPolicyDoc
	if err := h.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find policy: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policy": policyDocRef(c, p)})
}

func (h *Hub) headPolicyDoc(c *gin.Context) {
	id := c.Param("policy_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_policy", map[string]string{"policy_id": id}); !ok {
		return
	}
	var p models.IdentityPolicyDoc
	if err := h.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find policy: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Hub) patchPolicyDoc(c *gin.Context) {
	id := c.Param("policy_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_policy", map[string]string{"policy_id": id}); !ok {
		return
	}
	var body policyDocBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var p models.IdentityPolicyDoc
	if err := h.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find policy: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Policy.Blob != "" {
		p.Blob = body.Policy.Blob
	}
	if body.Policy.Type != "" {
		p.Type = body.Policy.Type
	}
	if err := h.DB.Save(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policy": policyDocRef(c, p)})
}

func (h *Hub) deletePolicyDoc(c *gin.Context) {
	id := c.Param("policy_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_policy", map[string]string{"policy_id": id}); !ok {
		return
	}
	h.DB.Where("policy_id = ?", id).Delete(&models.EndpointPolicyLink{})
	res := h.DB.Where("id = ?", id).Delete(&models.IdentityPolicyDoc{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find policy: " + id}})
		return
	}
	c.Status(http.StatusNoContent)
}

func policyDocRef(c *gin.Context, p models.IdentityPolicyDoc) map[string]any {
	return map[string]any{
		"id":    p.ID,
		"blob":  p.Blob,
		"type":  p.Type,
		"links": gin.H{"self": selfURL(c, "/v3/policies/"+p.ID)},
	}
}

func (h *Hub) listRegisteredLimits(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_registered_limits", nil); !ok {
		return
	}
	var list []models.RegisteredLimit
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, r := range list {
		out = append(out, registeredLimitRef(c, r))
	}
	c.JSON(http.StatusOK, gin.H{"registered_limits": out})
}

func (h *Hub) createRegisteredLimit(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_registered_limit", nil); !ok {
		return
	}
	var body struct {
		RegisteredLimit struct {
			ServiceID    string `json:"service_id"`
			RegionID     string `json:"region_id"`
			ResourceName string `json:"resource_name"`
			DefaultLimit int64  `json:"default"`
		} `json:"registered_limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.RegisteredLimit.ResourceName == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	r := models.RegisteredLimit{
		ID: uuid.NewString(), ServiceID: body.RegisteredLimit.ServiceID, RegionID: body.RegisteredLimit.RegionID,
		ResourceName: body.RegisteredLimit.ResourceName, Default: body.RegisteredLimit.DefaultLimit,
	}
	if err := h.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"registered_limit": registeredLimitRef(c, r)})
}

func (h *Hub) getRegisteredLimit(c *gin.Context) {
	id := c.Param("registered_limit_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_registered_limit", map[string]string{}); !ok {
		return
	}
	var r models.RegisteredLimit
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registered_limit": registeredLimitRef(c, r)})
}

func (h *Hub) patchRegisteredLimit(c *gin.Context) {
	id := c.Param("registered_limit_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_registered_limit", nil); !ok {
		return
	}
	var body struct {
		RegisteredLimit struct {
			DefaultLimit int64 `json:"default"`
		} `json:"registered_limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var r models.RegisteredLimit
	if err := h.DB.Where("id = ?", id).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	r.Default = body.RegisteredLimit.DefaultLimit
	if err := h.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registered_limit": registeredLimitRef(c, r)})
}

func (h *Hub) deleteRegisteredLimit(c *gin.Context) {
	id := c.Param("registered_limit_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_registered_limit", nil); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.RegisteredLimit{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func registeredLimitRef(c *gin.Context, r models.RegisteredLimit) map[string]any {
	return map[string]any{
		"id":            r.ID,
		"service_id":    r.ServiceID,
		"region_id":     r.RegionID,
		"resource_name": r.ResourceName,
		"default":       r.Default,
		"links":         gin.H{"self": selfURL(c, "/v3/registered_limits/"+r.ID)},
	}
}

func (h *Hub) listLimits(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_limits", nil); !ok {
		return
	}
	q := h.DB.Model(&models.Limit{})
	if p := c.Query("project_id"); p != "" {
		q = q.Where("project_id = ?", p)
	}
	var list []models.Limit
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, l := range list {
		out = append(out, limitRef(c, l))
	}
	c.JSON(http.StatusOK, gin.H{"limits": out})
}

func (h *Hub) createLimit(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_limit", nil); !ok {
		return
	}
	var body struct {
		Limit struct {
			ProjectID         string `json:"project_id"`
			RegisteredLimitID string `json:"registered_limit_id"`
			ResourceLimit     int64  `json:"resource_limit"`
		} `json:"limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Limit.ProjectID == "" || body.Limit.RegisteredLimitID == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	l := models.Limit{ID: uuid.NewString(), ProjectID: body.Limit.ProjectID, RegisteredLimitID: body.Limit.RegisteredLimitID, ResourceLimit: body.Limit.ResourceLimit}
	if err := h.DB.Create(&l).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"limit": limitRef(c, l)})
}

func (h *Hub) getLimit(c *gin.Context) {
	id := c.Param("limit_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_limit", nil); !ok {
		return
	}
	var l models.Limit
	if err := h.DB.Where("id = ?", id).First(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"limit": limitRef(c, l)})
}

func (h *Hub) patchLimit(c *gin.Context) {
	id := c.Param("limit_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_limit", nil); !ok {
		return
	}
	var body struct {
		Limit struct {
			ResourceLimit int64 `json:"resource_limit"`
		} `json:"limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var l models.Limit
	if err := h.DB.Where("id = ?", id).First(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	l.ResourceLimit = body.Limit.ResourceLimit
	if err := h.DB.Save(&l).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"limit": limitRef(c, l)})
}

func (h *Hub) deleteLimit(c *gin.Context) {
	id := c.Param("limit_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_limit", nil); !ok {
		return
	}
	res := h.DB.Where("id = ?", id).Delete(&models.Limit{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func limitRef(c *gin.Context, l models.Limit) map[string]any {
	return map[string]any{
		"id":                  l.ID,
		"project_id":          l.ProjectID,
		"registered_limit_id": l.RegisteredLimitID,
		"resource_limit":      l.ResourceLimit,
		"links":               gin.H{"self": selfURL(c, "/v3/limits/"+l.ID)},
	}
}

func (h *Hub) getLimitsModel(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_limits_model", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"limits_model": gin.H{
			"endpoint": selfURL(c, "/v3/limits/model"),
			"schema":   gin.H{},
		},
	})
}
