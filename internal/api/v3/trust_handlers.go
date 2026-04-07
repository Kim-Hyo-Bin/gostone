package v3

import (
	"net/http"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func registerV3TrustHandlers(trust *gin.RouterGroup, h *Hub) {
	g := trust.Group("/trusts")
	g.GET("", h.listTrusts)
	g.POST("", h.createTrust)
	g.GET("/:trust_id", h.getTrust)
	g.DELETE("/:trust_id", h.deleteTrust)
	g.GET("/:trust_id/roles", h.listTrustRoles)
	g.PUT("/:trust_id/roles/:role_id", h.addTrustRole)
	g.DELETE("/:trust_id/roles/:role_id", h.removeTrustRole)
}

func (h *Hub) listTrusts(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_trusts", nil); !ok {
		return
	}
	var list []models.Trust
	if err := h.DB.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, t := range list {
		out = append(out, trustRef(c, t))
	}
	c.JSON(http.StatusOK, gin.H{"trusts": out})
}

func (h *Hub) createTrust(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_trust", nil); !ok {
		return
	}
	var body struct {
		Trust struct {
			TrustorUserID string `json:"trustor_user_id"`
			TrusteeUserID string `json:"trustee_user_id"`
			Impersonation bool   `json:"impersonation"`
			ProjectID     string `json:"project_id"`
		} `json:"trust"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Trust.TrustorUserID == "" || body.Trust.TrusteeUserID == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	exp := time.Now().UTC().Add(24 * time.Hour)
	t := models.Trust{
		ID: uuid.NewString(), TrustorUserID: body.Trust.TrustorUserID, TrusteeUserID: body.Trust.TrusteeUserID,
		Impersonation: body.Trust.Impersonation, ProjectID: body.Trust.ProjectID, ExpiresAt: &exp,
	}
	if err := h.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"trust": trustRef(c, t)})
}

func (h *Hub) getTrust(c *gin.Context) {
	id := c.Param("trust_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_trust", map[string]string{}); !ok {
		return
	}
	var t models.Trust
	if err := h.DB.Where("id = ?", id).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trust": trustRef(c, t)})
}

func (h *Hub) deleteTrust(c *gin.Context) {
	id := c.Param("trust_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_trust", nil); !ok {
		return
	}
	h.DB.Where("trust_id = ?", id).Delete(&models.TrustRole{})
	res := h.DB.Where("id = ?", id).Delete(&models.Trust{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}

func trustRef(c *gin.Context, t models.Trust) map[string]any {
	m := map[string]any{
		"id":                 t.ID,
		"trustor_user_id":    t.TrustorUserID,
		"trustee_user_id":    t.TrusteeUserID,
		"impersonation":      t.Impersonation,
		"allow_redelegation": t.AllowRedelegation,
		"links":              gin.H{"self": selfURL(c, "/v3/OS-TRUST/trusts/"+t.ID)},
	}
	if t.ProjectID != "" {
		m["project_id"] = t.ProjectID
	}
	if t.ExpiresAt != nil {
		m["expires_at"] = t.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}
	return m
}

func (h *Hub) listTrustRoles(c *gin.Context) {
	tid := c.Param("trust_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_trust_roles", nil); !ok {
		return
	}
	type row struct {
		RoleID   string
		RoleName string
	}
	var rows []row
	err := h.DB.Table("trust_roles").
		Select("trust_roles.role_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = trust_roles.role_id").
		Where("trust_roles.trust_id = ?", tid).
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"id": r.RoleID, "name": r.RoleName})
	}
	c.JSON(http.StatusOK, gin.H{"roles": out})
}

func (h *Hub) addTrustRole(c *gin.Context) {
	tid := c.Param("trust_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:add_trust_role", nil); !ok {
		return
	}
	if err := h.DB.Create(&models.TrustRole{TrustID: tid, RoleID: rid}).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) removeTrustRole(c *gin.Context) {
	tid := c.Param("trust_id")
	rid := c.Param("role_id")
	if _, ok := h.requireAuthPolicy(c, "identity:remove_trust_role", nil); !ok {
		return
	}
	h.DB.Where("trust_id = ? AND role_id = ?", tid, rid).Delete(&models.TrustRole{})
	c.Status(http.StatusNoContent)
}
