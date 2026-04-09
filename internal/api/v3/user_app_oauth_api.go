package v3

import (
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func registerV3UserAppCredentials(v3 *gin.RouterGroup, h *Hub) {
	ac := v3.Group("/users/:user_id/application_credentials")
	ac.GET("", h.listAppCredentials)
	ac.POST("", h.createAppCredential)
	ac.GET("/:application_credential_id", h.getAppCredential)
	ac.DELETE("/:application_credential_id", h.deleteAppCredential)

	ar := v3.Group("/users/:user_id/access_rules")
	ar.GET("", h.listAccessRules)
	ar.POST("", h.createAccessRule)
	ar.GET("/:access_rule_id", h.getAccessRule)
	ar.DELETE("/:access_rule_id", h.deleteAccessRule)

	o1 := v3.Group("/users/:user_id/OS-OAUTH1/access_tokens")
	o1.GET("", h.listUserOAuth1Tokens)
	o1.POST("", h.createUserOAuth1Token)
	o1.GET("/:access_token_id", h.getUserOAuth1Token)
	o1.DELETE("/:access_token_id", h.deleteUserOAuth1Token)
	o1.GET("/:access_token_id/roles", h.listOAuth1TokenRoles)
	o1.PUT("/:access_token_id/roles/:role_id", h.addOAuth1TokenRole)
	o1.DELETE("/:access_token_id/roles/:role_id", h.removeOAuth1TokenRole)
}

func (h *Hub) listAppCredentials(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_application_credentials", map[string]string{"user_id": uid}); !ok {
		return
	}
	var list []models.ApplicationCredential
	h.DB.Where("user_id = ?", uid).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.ID, "name": x.Name, "description": x.Description})
	}
	c.JSON(http.StatusOK, gin.H{"application_credentials": out})
}

func (h *Hub) createAppCredential(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:create_application_credential", map[string]string{"user_id": uid}); !ok {
		return
	}
	var body struct {
		ApplicationCredential struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Secret      string `json:"secret"`
		} `json:"application_credential"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ApplicationCredential.Name == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	sec := body.ApplicationCredential.Secret
	if sec == "" {
		sec = uuid.NewString()
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(sec), bcrypt.MinCost)
	x := models.ApplicationCredential{ID: uuid.NewString(), UserID: uid, Name: body.ApplicationCredential.Name, Description: body.ApplicationCredential.Description, SecretHash: string(hash)}
	if err := h.DB.Create(&x).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"application_credential": gin.H{"id": x.ID, "name": x.Name, "secret": sec}})
}

func (h *Hub) getAppCredential(c *gin.Context) {
	uid := c.Param("user_id")
	id := c.Param("application_credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_application_credential", map[string]string{"user_id": uid}); !ok {
		return
	}
	var x models.ApplicationCredential
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&x).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"application_credential": gin.H{"id": x.ID, "name": x.Name}})
}

func (h *Hub) deleteAppCredential(c *gin.Context) {
	uid := c.Param("user_id")
	id := c.Param("application_credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_application_credential", map[string]string{"user_id": uid}); !ok {
		return
	}
	h.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.ApplicationCredential{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listAccessRules(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_access_rules", map[string]string{"user_id": uid}); !ok {
		return
	}
	var list []models.AccessRule
	h.DB.Where("user_id = ?", uid).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.ID, "path": x.Path, "method": x.Method})
	}
	c.JSON(http.StatusOK, gin.H{"access_rules": out})
}

func (h *Hub) createAccessRule(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:create_access_rule", map[string]string{"user_id": uid}); !ok {
		return
	}
	var body struct {
		AccessRule struct {
			Path   string `json:"path"`
			Method string `json:"method"`
		} `json:"access_rule"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.AccessRule.Path == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	x := models.AccessRule{ID: uuid.NewString(), UserID: uid, Path: body.AccessRule.Path, Method: strings.ToUpper(body.AccessRule.Method)}
	if err := h.DB.Create(&x).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"access_rule": gin.H{"id": x.ID}})
}

func (h *Hub) getAccessRule(c *gin.Context) {
	uid := c.Param("user_id")
	id := c.Param("access_rule_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_access_rule", map[string]string{"user_id": uid}); !ok {
		return
	}
	var x models.AccessRule
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&x).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_rule": gin.H{"id": x.ID, "path": x.Path, "method": x.Method}})
}

func (h *Hub) deleteAccessRule(c *gin.Context) {
	uid := c.Param("user_id")
	id := c.Param("access_rule_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_access_rule", map[string]string{"user_id": uid}); !ok {
		return
	}
	h.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.AccessRule{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listUserOAuth1Tokens(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_oauth1_access_tokens", map[string]string{"user_id": uid}); !ok {
		return
	}
	var list []models.OAuth1AccessToken
	h.DB.Where("user_id = ?", uid).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, x := range list {
		out = append(out, map[string]any{"id": x.ID})
	}
	c.JSON(http.StatusOK, gin.H{"access_tokens": out})
}

func (h *Hub) createUserOAuth1Token(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:create_oauth1_access_token", map[string]string{"user_id": uid}); !ok {
		return
	}
	var body struct {
		AccessToken struct {
			ConsumerID string `json:"consumer_id"`
		} `json:"access_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.AccessToken.ConsumerID == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	x := models.OAuth1AccessToken{ID: uuid.NewString(), UserID: uid, ConsumerID: body.AccessToken.ConsumerID, Secret: uuid.NewString()}
	if err := h.DB.Create(&x).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"access_token": gin.H{"id": x.ID}})
}

func (h *Hub) getUserOAuth1Token(c *gin.Context) {
	uid := c.Param("user_id")
	id := c.Param("access_token_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_oauth1_access_token", map[string]string{"user_id": uid}); !ok {
		return
	}
	var x models.OAuth1AccessToken
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&x).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": gin.H{"id": x.ID}})
}

func (h *Hub) deleteUserOAuth1Token(c *gin.Context) {
	uid := c.Param("user_id")
	id := c.Param("access_token_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_oauth1_access_token", map[string]string{"user_id": uid}); !ok {
		return
	}
	h.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&models.OAuth1AccessToken{})
	c.Status(http.StatusNoContent)
}

func (h *Hub) listOAuth1TokenRoles(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_oauth1_token_roles", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": []any{}})
}

func (h *Hub) addOAuth1TokenRole(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:assign_oauth1_token_role", nil); !ok {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Hub) removeOAuth1TokenRole(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:remove_oauth1_token_role", nil); !ok {
		return
	}
	c.Status(http.StatusNoContent)
}
