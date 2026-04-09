package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/Kim-Hyo-Bin/gostone/internal/catalog"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func registerV3OSRevokeOAuthCert(v3 *gin.RouterGroup, h *Hub) {
	rev := v3.Group("/OS-REVOKE")
	rev.GET("/events", h.listRevokeEvents)
	rev.POST("/events", h.createRevokeEvent)

	o2 := v3.Group("/OS-OAUTH2")
	o2.POST("/token", h.postOAuth2Token)

	o1 := v3.Group("/OS-OAUTH1")
	o1.POST("/request_token", h.postOAuth1RequestToken)
	o1.POST("/access_token", h.postOAuth1AccessToken)
	o1.GET("/authorize/:request_token_id", h.oauth1AuthorizeStub)

	con := o1.Group("/consumers")
	con.GET("", h.listOAuthConsumers)
	con.POST("", h.createOAuthConsumer)
	con.GET("/:consumer_id", h.getOAuthConsumer)
	con.PATCH("/:consumer_id", h.patchOAuthConsumer)
	con.PUT("/:consumer_id", h.patchOAuthConsumer)
	con.DELETE("/:consumer_id", h.deleteOAuthConsumer)

	cert := v3.Group("/OS-SIMPLE-CERT")
	cert.GET("/ca", h.getSimpleCA)
	cert.GET("/certificates", h.listSimpleCerts)
}

func (h *Hub) listRevokeEvents(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_revoke_events", nil); !ok {
		return
	}
	var list []models.RevokeEvent
	if err := h.DB.Order("id desc").Limit(500).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, e := range list {
		out = append(out, map[string]any{"audit_id": e.AuditID, "domain_id": e.DomainID})
	}
	c.JSON(http.StatusOK, gin.H{"events": out})
}

func (h *Hub) createRevokeEvent(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_revoke_event", nil); !ok {
		return
	}
	var body struct {
		Event struct {
			AuditID  string `json:"audit_id"`
			DomainID string `json:"domain_id"`
			Reason   string `json:"reason"`
		} `json:"event"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	e := models.RevokeEvent{AuditID: body.Event.AuditID, DomainID: body.Event.DomainID, Reason: body.Event.Reason}
	if err := h.DB.Create(&e).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"event": gin.H{"id": e.ID}})
}

func (h *Hub) postOAuth2Token(c *gin.Context) {
	var req struct {
		GrantType string `json:"grant_type" form:"grant_type"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.GrantType == "" {
		_ = c.ShouldBind(&req)
	}
	if strings.TrimSpace(req.GrantType) != "password" {
		httperr.BadRequest(c, "unsupported_grant_type")
		return
	}
	httperr.NotImplemented(c, "Use /v3/auth/tokens for password grants.")
}

func (h *Hub) postOAuth1RequestToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"oauth_token": uuid.NewString(), "oauth_token_secret": uuid.NewString()})
}

func (h *Hub) postOAuth1AccessToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"oauth_token": uuid.NewString(), "oauth_token_secret": uuid.NewString()})
}

func (h *Hub) oauth1AuthorizeStub(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"authorized": true})
}

func (h *Hub) listOAuthConsumers(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_oauth_consumers", nil); !ok {
		return
	}
	var list []models.OAuthConsumer
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.ID, "description": x.Description})
	}
	c.JSON(http.StatusOK, gin.H{"consumers": out})
}

func (h *Hub) createOAuthConsumer(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_oauth_consumer", nil); !ok {
		return
	}
	var body struct {
		Consumer struct {
			Description string `json:"description"`
		} `json:"consumer"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	sec := uuid.NewString()
	x := models.OAuthConsumer{ID: uuid.NewString(), Secret: sec, Description: body.Consumer.Description}
	if err := h.DB.Create(&x).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"consumer": gin.H{"id": x.ID, "secret": sec}})
}

func (h *Hub) getOAuthConsumer(c *gin.Context) {
	id := c.Param("consumer_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_oauth_consumer", nil); !ok {
		return
	}
	var x models.OAuthConsumer
	if err := h.DB.Where("id = ?", id).First(&x).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"consumer": gin.H{"id": x.ID, "description": x.Description}})
}

func (h *Hub) patchOAuthConsumer(c *gin.Context) {
	id := c.Param("consumer_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_oauth_consumer", nil); !ok {
		return
	}
	var body struct {
		Consumer struct {
			Description string `json:"description"`
		} `json:"consumer"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var x models.OAuthConsumer
	if err := h.DB.Where("id = ?", id).First(&x).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	if body.Consumer.Description != "" {
		x.Description = body.Consumer.Description
	}
	_ = h.DB.Save(&x)
	c.JSON(http.StatusOK, gin.H{"consumer": gin.H{"id": x.ID, "description": x.Description}})
}

func (h *Hub) deleteOAuthConsumer(c *gin.Context) {
	id := c.Param("consumer_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_oauth_consumer", nil); !ok {
		return
	}
	h.DB.Where("id = ?", id).Delete(&models.OAuthConsumer{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) getSimpleCA(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_simple_ca", nil); !ok {
		return
	}
	c.Data(http.StatusOK, "application/x-pem-file", []byte("-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJAKExampleGostone\n-----END CERTIFICATE-----\n"))
}

func (h *Hub) listSimpleCerts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"certificates": []any{}})
}

func registerV3OSEPFilter(v3 *gin.RouterGroup, h *Hub) {
	epf := v3.Group("/OS-EP-FILTER")
	epf.GET("/endpoints/:endpoint_id/projects", h.listEndpointProjects)
	epf.GET("/projects/:project_id/endpoints", h.listProjectFilteredEndpoints)
	epf.PUT("/projects/:project_id/endpoints/:endpoint_id", h.addProjectEndpointFilter)
	epf.DELETE("/projects/:project_id/endpoints/:endpoint_id", h.removeProjectEndpointFilter)
	epf.GET("/projects/:project_id/endpoint_groups", h.listEndpointGroups)
	epf.GET("/endpoint_groups/:endpoint_group_id/endpoints", h.listEndpointGroupEndpoints)
	epf.PUT("/endpoint_groups/:endpoint_group_id/endpoints/:endpoint_id", h.addEndpointGroupMember)
	epf.GET("/endpoint_groups/:endpoint_group_id/projects", h.listEndpointGroupProjects)
	epf.PUT("/endpoint_groups/:endpoint_group_id/projects/:project_id", h.addEndpointGroupProject)
}

func (h *Hub) listEndpointProjects(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_project_endpoint_filters", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": []any{}})
}

func (h *Hub) listProjectFilteredEndpoints(c *gin.Context) {
	pid := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_project_endpoint_filters", map[string]string{"project_id": pid}); !ok {
		return
	}
	var list []models.ProjectEndpointFilter
	h.DB.Where("project_id = ?", pid).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.EndpointID})
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": out})
}

func (h *Hub) addProjectEndpointFilter(c *gin.Context) {
	pid := c.Param("project_id")
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_project_endpoint_filter", map[string]string{"project_id": pid}); !ok {
		return
	}
	_ = h.DB.Create(&models.ProjectEndpointFilter{ProjectID: pid, EndpointID: eid})
	c.Status(http.StatusNoContent)
}

func (h *Hub) removeProjectEndpointFilter(c *gin.Context) {
	pid := c.Param("project_id")
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_project_endpoint_filter", map[string]string{"project_id": pid}); !ok {
		return
	}
	h.DB.Where("project_id = ? AND endpoint_id = ?", pid, eid).Delete(&models.ProjectEndpointFilter{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listEndpointGroups(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_endpoint_groups", nil); !ok {
		return
	}
	var list []models.EndpointGroup
	h.DB.Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, g := range list {
		out = append(out, map[string]any{"id": g.ID, "name": g.Name})
	}
	c.JSON(http.StatusOK, gin.H{"endpoint_groups": out})
}

func (h *Hub) listEndpointGroupEndpoints(c *gin.Context) {
	gid := c.Param("endpoint_group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_endpoint_group_endpoints", nil); !ok {
		return
	}
	var list []models.EndpointGroupMember
	h.DB.Where("endpoint_group_id = ?", gid).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.EndpointID})
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": out})
}

func (h *Hub) addEndpointGroupMember(c *gin.Context) {
	gid := c.Param("endpoint_group_id")
	eid := c.Param("endpoint_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_endpoint_group_endpoint", nil); !ok {
		return
	}
	_ = h.DB.Create(&models.EndpointGroupMember{EndpointGroupID: gid, EndpointID: eid})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listEndpointGroupProjects(c *gin.Context) {
	gid := c.Param("endpoint_group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_endpoint_group_projects", nil); !ok {
		return
	}
	var list []models.EndpointGroupProject
	h.DB.Where("endpoint_group_id = ?", gid).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.ProjectID})
	}
	c.JSON(http.StatusOK, gin.H{"projects": out})
}

func (h *Hub) addEndpointGroupProject(c *gin.Context) {
	gid := c.Param("endpoint_group_id")
	pid := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_endpoint_group_project", nil); !ok {
		return
	}
	_ = h.DB.Create(&models.EndpointGroupProject{EndpointGroupID: gid, ProjectID: pid})
	c.Status(http.StatusNoContent)
}

func registerV3InheritAndSystem(v3 *gin.RouterGroup, h *Hub) {
	inh := v3.Group("/OS-INHERIT")
	inh.GET("/domains/:domain_id/groups/:group_id/roles/:role_id/inherited_to_projects", h.emptyRolesJSON)
	inh.GET("/domains/:domain_id/groups/:group_id/roles/inherited_to_projects", h.emptyRolesJSON)
	inh.GET("/domains/:domain_id/users/:user_id/roles/:role_id/inherited_to_projects", h.emptyRolesJSON)
	inh.GET("/domains/:domain_id/users/:user_id/roles/inherited_to_projects", h.emptyRolesJSON)
	inh.GET("/projects/:project_id/users/:user_id/roles/:role_id/inherited_to_projects", h.emptyRolesJSON)
	inh.GET("/projects/:project_id/groups/:group_id/roles/:role_id/inherited_to_projects", h.emptyRolesJSON)

	sys := v3.Group("/system")
	sys.GET("/users/:user_id/roles", h.listSystemUserRoles)
	sys.PUT("/users/:user_id/roles/:role_id", h.putSystemUserRole)
	sys.DELETE("/users/:user_id/roles/:role_id", h.deleteSystemUserRole)
	sys.GET("/groups/:group_id/roles", h.listSystemGroupRoles)
	sys.PUT("/groups/:group_id/roles/:role_id", h.putSystemGroupRole)
	sys.DELETE("/groups/:group_id/roles/:role_id", h.deleteSystemGroupRole)
}

func (h *Hub) emptyRolesJSON(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_inherited_roles", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": []any{}, "links": gin.H{"self": selfURL(c, c.Request.URL.Path)}})
}

func (h *Hub) listSystemUserRoles(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_system_user_roles", nil); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	h.DB.Table("user_system_roles").
		Select("user_system_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = user_system_roles.role_id").
		Where("user_system_roles.user_id = ?", uid).
		Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"id": r.RoleID, "name": r.RoleName})
	}
	c.JSON(http.StatusOK, gin.H{"roles": out})
}

func (h *Hub) putSystemUserRole(c *gin.Context) {
	uid := c.Param("user_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:assign_system_user_role", nil); !ok {
		return
	}
	_ = h.DB.Create(&models.UserSystemRole{UserID: uid, RoleID: rid})
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteSystemUserRole(c *gin.Context) {
	uid := c.Param("user_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_system_user_role", nil); !ok {
		return
	}
	h.DB.Where("user_id = ? AND role_id = ?", uid, rid).Delete(&models.UserSystemRole{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listSystemGroupRoles(c *gin.Context) {
	gid := c.Param("group_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_system_group_roles", nil); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	h.DB.Table("group_system_roles").
		Select("group_system_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = group_system_roles.role_id").
		Where("group_system_roles.group_id = ?", gid).
		Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"id": r.RoleID, "name": r.RoleName})
	}
	c.JSON(http.StatusOK, gin.H{"roles": out})
}

func (h *Hub) putSystemGroupRole(c *gin.Context) {
	gid := c.Param("group_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:assign_system_group_role", nil); !ok {
		return
	}
	_ = h.DB.Create(&models.GroupSystemRole{GroupID: gid, RoleID: rid})
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteSystemGroupRole(c *gin.Context) {
	gid := c.Param("group_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_system_group_role", nil); !ok {
		return
	}
	h.DB.Where("group_id = ? AND role_id = ?", gid, rid).Delete(&models.GroupSystemRole{})
	c.Status(http.StatusNoContent)
}

func registerV3AuthResources(v3 *gin.RouterGroup, h *Hub) {
	v3.GET("/auth/catalog", h.getAuthCatalog)
	v3.GET("/auth/projects", h.listAuthProjects)
	v3.GET("/auth/domains", h.listAuthDomains)
	v3.GET("/auth/system", h.getAuthSystem)
	v3.GET("/OS-FEDERATION/projects", h.listAuthProjects)
	v3.GET("/OS-FEDERATION/domains", h.listAuthDomains)
}

func (h *Hub) getAuthCatalog(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_auth_catalog", nil); !ok {
		return
	}
	cat, err := catalog.Build(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"catalog": cat})
}

func (h *Hub) listAuthProjects(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:list_auth_projects", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	type row struct {
		ID   string
		Name string
	}
	var rows []row
	h.DB.Model(&models.UserProjectRole{}).
		Select("projects.id AS id, projects.name AS name").
		Joins("JOIN projects ON projects.id = user_project_roles.project_id").
		Where("user_project_roles.user_id = ?", actx.UserID).
		Group("projects.id, projects.name").
		Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"id": r.ID, "name": r.Name})
	}
	c.JSON(http.StatusOK, gin.H{"projects": out})
}

func (h *Hub) listAuthDomains(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:list_auth_domains", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var d models.Domain
	if err := h.DB.Where("id = ?", actx.DomainID).First(&d).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"domains": []any{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"domains": []map[string]any{{"id": d.ID, "name": d.Name}}})
}

func (h *Hub) getAuthSystem(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_auth_system", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"system": gin.H{"name": "gostone"}})
}
