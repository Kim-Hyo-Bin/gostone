package v3

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerV3Users(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/users")
	g.GET("", h.listUsers)
	g.POST("", h.stubRoute("POST /v3/users"))
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
