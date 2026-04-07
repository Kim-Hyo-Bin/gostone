package v3

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// headProject mirrors get_project policy and existence checks with an empty body.
func (h *Hub) headProject(c *gin.Context) {
	id := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_project", map[string]string{"project_id": id}); !ok {
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
	c.Status(http.StatusOK)
}

func registerV3ProjectsAPI(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/projects")
	g.GET("/:project_id/users/:user_id/roles", h.listProjectUserRoles)
	g.PUT("/:project_id/users/:user_id/roles/:role_id", h.putProjectUserRole)
	g.DELETE("/:project_id/users/:user_id/roles/:role_id", h.deleteProjectUserRole)
	g.GET("/:project_id/groups/:group_id/roles", h.listGroupProjectRoles)
	g.PUT("/:project_id/groups/:group_id/roles/:role_id", h.putGroupProjectRole)
	g.DELETE("/:project_id/groups/:group_id/roles/:role_id", h.deleteGroupProjectRole)
	g.GET("", h.listProjects)
	g.POST("", h.createProject)
	g.GET("/:project_id", h.getProject)
	g.HEAD("/:project_id", h.headProject)
	g.PATCH("/:project_id", h.patchProject)
	g.DELETE("/:project_id", h.deleteProject)
}

func (h *Hub) listProjects(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_projects", nil); !ok {
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
	if _, ok := h.requireAuthPolicy(c, "identity:create_project", nil); !ok {
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
	id := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_project", map[string]string{"project_id": id}); !ok {
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
	id := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_project", map[string]string{"project_id": id}); !ok {
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
	id := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_project", map[string]string{"project_id": id}); !ok {
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", id).Delete(&models.UserProjectRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.GroupProjectRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.ProjectTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.ProjectEndpointFilter{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.Limit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.EndpointGroupProject{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", id).Delete(&models.Project{})
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
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find project: " + id}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
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
