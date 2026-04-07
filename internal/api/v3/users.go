package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func registerV3Users(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/users")
	g.GET("", h.listUsers)
	g.POST("", h.createUser)
	g.GET("/:user_id", h.getUser)
	g.HEAD("/:user_id", h.headUser)
	g.PATCH("/:user_id", h.patchUser)
	g.PUT("/:user_id", h.patchUser)
	g.DELETE("/:user_id", h.deleteUser)
}

func (h *Hub) listUsers(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_users", nil); !ok {
		return
	}
	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": http.StatusInternalServerError, "message": err.Error()},
		})
		return
	}
	refs := make([]map[string]any, 0, len(users))
	for _, u := range users {
		refs = append(refs, userRef(c, u))
	}
	c.JSON(http.StatusOK, gin.H{"users": refs, "links": gin.H{"self": selfURL(c, "/v3/users")}})
}

type createUserBody struct {
	User struct {
		Name        string `json:"name"`
		DomainID    string `json:"domain_id"`
		Password    string `json:"password"`
		Enabled     *bool  `json:"enabled"`
		Description string `json:"description"`
	} `json:"user"`
}

func (h *Hub) createUser(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_user", nil); !ok {
		return
	}
	var body createUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	name := strings.TrimSpace(body.User.Name)
	domainID := strings.TrimSpace(body.User.DomainID)
	if name == "" || domainID == "" {
		httperr.BadRequest(c, "user name and domain_id are required")
		return
	}
	if strings.TrimSpace(body.User.Password) == "" {
		httperr.BadRequest(c, "password is required for local users")
		return
	}
	var dom models.Domain
	if err := h.DB.Where("id = ?", domainID).First(&dom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httperr.BadRequest(c, "invalid domain")
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	en := true
	if body.User.Enabled != nil {
		en = *body.User.Enabled
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.User.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	u := models.User{
		ID:           uuid.NewString(),
		DomainID:     domainID,
		Name:         name,
		Enabled:      en,
		PasswordHash: string(hash),
	}
	if err := h.DB.Create(&u).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{"code": http.StatusConflict, "message": "Duplicate user name in domain."},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	ref := userRef(c, u)
	if body.User.Description != "" {
		ref["description"] = body.User.Description
	}
	c.JSON(http.StatusCreated, gin.H{"user": ref})
}

func (h *Hub) getUser(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_user", map[string]string{"user_id": uid}); !ok {
		return
	}
	var u models.User
	if err := h.DB.Where("id = ?", uid).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    http.StatusNotFound,
					"title":   "Not Found",
					"message": "Could not find user: " + uid,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": http.StatusInternalServerError, "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userRef(c, u)})
}

func (h *Hub) headUser(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_user", map[string]string{"user_id": uid}); !ok {
		return
	}
	var n int64
	if err := h.DB.Model(&models.User{}).Where("id = ?", uid).Count(&n).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

type patchUserBody struct {
	User struct {
		Name        string `json:"name"`
		DomainID    string `json:"domain_id"`
		Password    string `json:"password"`
		Enabled     *bool  `json:"enabled"`
		Description string `json:"description"`
	} `json:"user"`
}

func (h *Hub) patchUser(c *gin.Context) {
	uid := c.Param("user_id")
	actx, ok := h.requireAuthPolicy(c, "identity:update_user", map[string]string{"user_id": uid})
	if !ok {
		return
	}
	var body patchUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var u models.User
	if err := h.DB.Where("id = ?", uid).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": http.StatusNotFound, "message": "Could not find user: " + uid},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	isAdmin := actx.HasRole("admin")
	if !isAdmin {
		if strings.TrimSpace(body.User.Name) != "" && body.User.Name != u.Name {
			httperr.Forbidden(c, "You are not authorized to change user name.")
			return
		}
		if strings.TrimSpace(body.User.DomainID) != "" && body.User.DomainID != u.DomainID {
			httperr.Forbidden(c, "You are not authorized to change domain_id.")
			return
		}
	}
	if isAdmin && strings.TrimSpace(body.User.DomainID) != "" && body.User.DomainID != u.DomainID {
		var dom models.Domain
		if err := h.DB.Where("id = ?", body.User.DomainID).First(&dom).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				httperr.BadRequest(c, "invalid domain")
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		u.DomainID = body.User.DomainID
	}
	if strings.TrimSpace(body.User.Name) != "" {
		u.Name = strings.TrimSpace(body.User.Name)
	}
	if body.User.Enabled != nil {
		u.Enabled = *body.User.Enabled
	}
	if strings.TrimSpace(body.User.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.User.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
			return
		}
		u.PasswordHash = string(hash)
	}
	if err := h.DB.Save(&u).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{"code": http.StatusConflict, "message": "Duplicate user name in domain."},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	ref := userRef(c, u)
	if body.User.Description != "" {
		ref["description"] = body.User.Description
	}
	c.JSON(http.StatusOK, gin.H{"user": ref})
}

func (h *Hub) deleteUser(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_user", map[string]string{"user_id": uid}); !ok {
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", uid).Delete(&models.UserProjectRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&models.AuthToken{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", uid).Delete(&models.User{})
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
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": http.StatusNotFound, "message": "Could not find user: " + uid},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}

func userRef(c *gin.Context, u models.User) map[string]any {
	return map[string]any{
		"id":                  u.ID,
		"name":                u.Name,
		"domain_id":           u.DomainID,
		"enabled":             u.Enabled,
		"links":               gin.H{"self": selfURL(c, "/v3/users/"+u.ID)},
		"password_expires_at": nil,
	}
}

func selfURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + path
}
