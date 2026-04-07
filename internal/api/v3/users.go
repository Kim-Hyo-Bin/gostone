package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
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
	g.HEAD("/:user_id", h.stubRoute("HEAD /v3/users/:user_id"))
	g.PATCH("/:user_id", h.stubRoute("PATCH /v3/users/:user_id"))
	g.PUT("/:user_id", h.stubRoute("PUT /v3/users/:user_id"))
	g.DELETE("/:user_id", h.stubRoute("DELETE /v3/users/:user_id"))
}

func (h *Hub) listUsers(c *gin.Context) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:list_users", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
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
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	if !h.Policy.Allow("identity:create_user", actx, nil) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
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
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return
	}
	uid := c.Param("user_id")
	if !h.Policy.Allow("identity:get_user", actx, map[string]string{"user_id": uid}) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
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
