package v3

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func registerV3CredentialsAPI(v3 *gin.RouterGroup, h *Hub) {
	g := v3.Group("/credentials")
	g.GET("", h.listCredentials)
	g.POST("", h.createCredential)
	g.GET("/:credential_id", h.getCredential)
	g.PATCH("/:credential_id", h.patchCredential)
	g.PUT("/:credential_id", h.patchCredential)
	g.DELETE("/:credential_id", h.deleteCredential)
}

func (h *Hub) listCredentials(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:list_credentials", nil); !ok {
		return
	}
	q := h.DB.Model(&models.Credential{})
	if u := c.Query("user_id"); u != "" {
		q = q.Where("user_id = ?", u)
	}
	var list []models.Credential
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	var ec2 []models.EC2Credential
	if err := h.DB.Find(&ec2).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list)+len(ec2))
	for _, cr := range list {
		out = append(out, credentialRef(c, cr))
	}
	for _, e := range ec2 {
		if u := c.Query("user_id"); u != "" && e.UserID != u {
			continue
		}
		out = append(out, ec2CredentialRef(c, e))
	}
	c.JSON(http.StatusOK, gin.H{"credentials": out})
}

type credBody struct {
	Credential struct {
		UserID    string `json:"user_id"`
		ProjectID string `json:"project_id"`
		Type      string `json:"type"`
		Blob      string `json:"blob"`
	} `json:"credential"`
}

func (h *Hub) createCredential(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:create_credential", nil); !ok {
		return
	}
	var body credBody
	if err := c.ShouldBindJSON(&body); err != nil || body.Credential.UserID == "" || body.Credential.Type == "" {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	cr := models.Credential{ID: uuid.NewString(), UserID: body.Credential.UserID, ProjectID: body.Credential.ProjectID, Type: body.Credential.Type, Blob: body.Credential.Blob}
	if err := h.DB.Create(&cr).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"credential": credentialRef(c, cr)})
}

func (h *Hub) getCredential(c *gin.Context) {
	id := c.Param("credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_credential", map[string]string{"credential_id": id}); !ok {
		return
	}
	var cr models.Credential
	if err := h.DB.Where("id = ?", id).First(&cr).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"credential": credentialRef(c, cr)})
		return
	}
	var ec2 models.EC2Credential
	if err := h.DB.Where("id = ?", id).First(&ec2).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find credential"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"credential": ec2CredentialRef(c, ec2)})
}

func (h *Hub) patchCredential(c *gin.Context) {
	id := c.Param("credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:update_credential", map[string]string{"credential_id": id}); !ok {
		return
	}
	var body credBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	var cr models.Credential
	if err := h.DB.Where("id = ?", id).First(&cr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find credential"}})
		return
	}
	if body.Credential.Blob != "" {
		cr.Blob = body.Credential.Blob
	}
	if body.Credential.Type != "" {
		cr.Type = body.Credential.Type
	}
	if err := h.DB.Save(&cr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"credential": credentialRef(c, cr)})
}

func (h *Hub) deleteCredential(c *gin.Context) {
	id := c.Param("credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_credential", map[string]string{"credential_id": id}); !ok {
		return
	}
	if res := h.DB.Where("id = ?", id).Delete(&models.Credential{}); res.RowsAffected > 0 {
		c.Status(http.StatusNoContent)
		return
	}
	if res := h.DB.Where("id = ?", id).Delete(&models.EC2Credential{}); res.RowsAffected > 0 {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Could not find credential"}})
}

func credentialRef(c *gin.Context, cr models.Credential) map[string]any {
	return map[string]any{
		"id":         cr.ID,
		"user_id":    cr.UserID,
		"project_id": cr.ProjectID,
		"type":       cr.Type,
		"blob":       cr.Blob,
		"links":      gin.H{"self": selfURL(c, "/v3/credentials/"+cr.ID)},
	}
}

func ec2CredentialRef(c *gin.Context, e models.EC2Credential) map[string]any {
	inner, _ := json.Marshal(map[string]string{"access": e.AccessKey})
	blob := base64.StdEncoding.EncodeToString(inner)
	return map[string]any{
		"id":      e.ID,
		"user_id": e.UserID,
		"type":    "ec2",
		"blob":    blob,
		"links":   gin.H{"self": selfURL(c, "/v3/credentials/"+e.ID)},
	}
}

func registerV3UserEC2Credentials(v3 *gin.RouterGroup, h *Hub) {
	u := v3.Group("/users/:user_id/credentials/OS-EC2")
	u.GET("", h.listUserEC2Credentials)
	u.POST("", h.createUserEC2Credential)
	u.GET("/:credential_id", h.getUserEC2Credential)
	u.DELETE("/:credential_id", h.deleteUserEC2Credential)
}

func (h *Hub) listUserEC2Credentials(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:list_credentials_for_user", map[string]string{"user_id": uid}); !ok {
		return
	}
	var list []models.EC2Credential
	if err := h.DB.Where("user_id = ?", uid).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, e := range list {
		out = append(out, ec2CredentialRef(c, e))
	}
	c.JSON(http.StatusOK, gin.H{"credentials": out})
}

func (h *Hub) createUserEC2Credential(c *gin.Context) {
	uid := c.Param("user_id")
	if _, ok := h.requireAuthPolicy(c, "identity:create_ec2_credential", map[string]string{"user_id": uid}); !ok {
		return
	}
	var body struct {
		Credential struct {
			Blob string `json:"blob"`
		} `json:"credential"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(body.Credential.Blob))
	if err != nil {
		httperr.BadRequest(c, "Invalid blob")
		return
	}
	var parsed struct {
		Access string `json:"access"`
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		httperr.BadRequest(c, "Invalid blob JSON")
		return
	}
	if parsed.Access == "" || parsed.Secret == "" {
		httperr.BadRequest(c, "access and secret required in blob")
		return
	}
	e := models.EC2Credential{ID: uuid.NewString(), UserID: uid, AccessKey: parsed.Access, SecretKey: parsed.Secret}
	if err := h.DB.Create(&e).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": 409, "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"credential": ec2CredentialRef(c, e)})
}

func (h *Hub) getUserEC2Credential(c *gin.Context) {
	uid := c.Param("user_id")
	cid := c.Param("credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_ec2_credential", map[string]string{"user_id": uid}); !ok {
		return
	}
	var e models.EC2Credential
	if err := h.DB.Where("id = ? AND user_id = ?", cid, uid).First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"credential": ec2CredentialRef(c, e)})
}

func (h *Hub) deleteUserEC2Credential(c *gin.Context) {
	uid := c.Param("user_id")
	cid := c.Param("credential_id")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_ec2_credential", map[string]string{"user_id": uid}); !ok {
		return
	}
	res := h.DB.Where("id = ? AND user_id = ?", cid, uid).Delete(&models.EC2Credential{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.Status(http.StatusNoContent)
}
