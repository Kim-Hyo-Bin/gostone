package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func registerV3ProjectTags(v3 *gin.RouterGroup, h *Hub) {
	p := v3.Group("/projects/:project_id/tags")
	p.GET("", h.listProjectTags)
	p.PUT("/:value", h.putProjectTag)
	p.DELETE("/:value", h.deleteProjectTag)
}

func (h *Hub) listProjectTags(c *gin.Context) {
	pid := c.Param("project_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_project_tags", map[string]string{"project_id": pid}); !ok {
		return
	}
	var tags []models.ProjectTag
	if err := h.DB.Where("project_id = ?", pid).Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	vals := make([]string, 0, len(tags))
	for _, t := range tags {
		vals = append(vals, t.Value)
	}
	c.JSON(http.StatusOK, gin.H{"tags": vals})
}

func (h *Hub) putProjectTag(c *gin.Context) {
	pid := c.Param("project_id")
	val := c.Param("value")
	if _, ok := h.requireAuthPolicy(c, "identity:add_project_tag", map[string]string{"project_id": pid}); !ok {
		return
	}
	if val == "" {
		httperr.BadRequest(c, "Invalid tag")
		return
	}
	var ex models.ProjectTag
	if err := h.DB.Where("project_id = ? AND value = ?", pid, val).First(&ex).Error; err == nil {
		c.Status(http.StatusNoContent)
		return
	}
	if err := h.DB.Create(&models.ProjectTag{ProjectID: pid, Value: val}).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) deleteProjectTag(c *gin.Context) {
	pid := c.Param("project_id")
	val := c.Param("value")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_project_tag", map[string]string{"project_id": pid}); !ok {
		return
	}
	res := h.DB.Where("project_id = ? AND value = ?", pid, val).Delete(&models.ProjectTag{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Tag not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func registerV3UserMisc(v3 *gin.RouterGroup, h *Hub) {
	u := v3.Group("/users/:user_id")
	u.PATCH("/password", h.patchUserPassword)
	u.GET("/groups", h.listUserGroups)
	u.GET("/projects", h.listUserProjects)
}

func (h *Hub) patchUserPassword(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:change_user_password", map[string]string{"user_id": uid}); !ok {
		return
	}
	var body struct {
		User struct {
			Password         string `json:"password"`
			OriginalPassword string `json:"original_password"`
		} `json:"user"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.User.Password) == "" {
		httperr.BadRequest(c, "Malformed request body")
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
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if actx.UserID == uid && !actx.HasRole("admin") {
		if body.User.OriginalPassword == "" {
			httperr.BadRequest(c, "original_password required")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.User.OriginalPassword)) != nil {
			httperr.Unauthorized(c, "Invalid original password")
			return
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.User.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	u.PasswordHash = string(hash)
	if err := h.DB.Save(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) listUserGroups(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_groups_for_user", map[string]string{"user_id": uid}); !ok {
		return
	}
	var gids []string
	if err := h.DB.Model(&models.GroupMember{}).Where("user_id = ?", uid).Pluck("group_id", &gids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	groups := make([]map[string]any, 0, len(gids))
	for _, id := range gids {
		groups = append(groups, map[string]any{"id": id})
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Hub) listUserProjects(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_projects_for_user", map[string]string{"user_id": uid}); !ok {
		return
	}
	var pids []string
	if err := h.DB.Model(&models.UserProjectRole{}).Where("user_id = ?", uid).Distinct("project_id").Pluck("project_id", &pids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	projs := make([]map[string]any, 0, len(pids))
	for _, id := range pids {
		projs = append(projs, map[string]any{"id": id})
	}
	c.JSON(http.StatusOK, gin.H{"projects": projs})
}
