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

func registerV3ProjectsAPI(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/projects")
	g.GET("", h.listProjects)
	g.POST("", h.createProject)
	g.GET("/:project_id", h.getProject)
	g.PATCH("/:project_id", h.patchProject)
	g.DELETE("/:project_id", h.deleteProject)
}

func (h *Hub) listProjects(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:list_projects", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var list []models.Project
	if err := h.DB.Order("name").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, projectRef(c, p))
	}
	c.JSON(http.StatusOK, gin.H{"projects": out, "links": gin.H{"self": selfURL(c, "/v3/projects")}})
}

type projectJSON struct {
	Project struct {
		Name     string `json:"name"`
		DomainID string `json:"domain_id"`
		Enabled  *bool  `json:"enabled"`
	} `json:"project"`
}

func (h *Hub) createProject(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:create_project", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var body projectJSON
	if err := c.ShouldBindJSON(&body); err != nil || body.Project.Name == "" || body.Project.DomainID == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	en := true
	if body.Project.Enabled != nil {
		en = *body.Project.Enabled
	}
	p := models.Project{ID: uuid.NewString(), Name: body.Project.Name, DomainID: body.Project.DomainID, Enabled: en}
	if err := h.DB.Create(&p).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"project": projectRef(c, p)})
}

func (h *Hub) getProject(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("project_id")
	if !h.Policy.Allow("identity:get_project", actx, map[string]string{"project_id": id}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var p models.Project
	if err := h.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find project: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": projectRef(c, p)})
}

func (h *Hub) patchProject(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("project_id")
	if !h.Policy.Allow("identity:update_project", actx, map[string]string{"project_id": id}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	var body projectJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var p models.Project
	if err := h.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find project: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	if body.Project.Name != "" {
		p.Name = body.Project.Name
	}
	if body.Project.DomainID != "" {
		p.DomainID = body.Project.DomainID
	}
	if body.Project.Enabled != nil {
		p.Enabled = *body.Project.Enabled
	}
	if err := h.DB.Save(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": projectRef(c, p)})
}

func (h *Hub) deleteProject(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("project_id")
	if !h.Policy.Allow("identity:delete_project", actx, map[string]string{"project_id": id}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return
	}
	res := h.DB.Delete(&models.Project{}, "id = ?", id)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": res.Error.Error()}})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find project: " + id}})
		return
	}
	c.Status(http.StatusNoContent)
}

func projectRef(c *gin.Context, p models.Project) map[string]any {
	return map[string]any{
		"id":          p.ID,
		"name":        p.Name,
		"domain_id":   p.DomainID,
		"enabled":     p.Enabled,
		"links":       gin.H{"self": selfURL(c, "/v3/projects/"+p.ID)},
		"description": nil,
		"tags":        []any{},
	}
}
