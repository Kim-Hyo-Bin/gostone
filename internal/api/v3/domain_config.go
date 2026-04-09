package v3

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/gin-gonic/gin"
)

func registerV3DomainConfigRoutes(g *gin.RouterGroup, h *Hub) {
	g.GET("/config/default", h.domainConfigGlobalDefault)
	g.GET("/config/:group/default", h.domainConfigGroupGlobalDefault)
	g.GET("/config/:group/:option/default", h.domainConfigOptionGlobalDefault)

	g.GET("/:domain_id/config/:group/:option", h.getDomainConfigOption)
	g.GET("/:domain_id/config/:group", h.getDomainConfigGroup)
	g.GET("/:domain_id/config", h.listDomainConfig)
	g.PATCH("/:domain_id/config", h.patchDomainConfig)
	g.PUT("/:domain_id/config", h.patchDomainConfig)
	g.DELETE("/:domain_id/config/:group/:option", h.deleteDomainConfigOption)
}

func (h *Hub) domainConfigGlobalDefault(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain_config", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": gin.H{}})
}

func (h *Hub) domainConfigGroupGlobalDefault(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain_config", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": gin.H{}})
}

func (h *Hub) domainConfigOptionGlobalDefault(c *gin.Context) {
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain_config", nil); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": gin.H{}})
}

func (h *Hub) listDomainConfig(c *gin.Context) {
	did := c.Param("domain_id")
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain_config", map[string]string{"domain_id": did}); !ok {
		return
	}
	var rows []models.DomainConfig
	if err := h.DB.Where("domain_id = ?", did).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	cfg := gin.H{}
	for _, r := range rows {
		cfg[r.Group+"/"+r.Option] = r.Value
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func (h *Hub) getDomainConfigGroup(c *gin.Context) {
	did := c.Param("domain_id")
	grp := c.Param("group")
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain_config", map[string]string{"domain_id": did}); !ok {
		return
	}
	var rows []models.DomainConfig
	h.DB.Where("domain_id = ? AND cfg_group = ?", did, grp).Find(&rows)
	cfg := gin.H{}
	for _, r := range rows {
		cfg[r.Option] = r.Value
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func (h *Hub) getDomainConfigOption(c *gin.Context) {
	did := c.Param("domain_id")
	grp := c.Param("group")
	opt := c.Param("option")
	if _, ok := h.requireAuthPolicy(c, "identity:get_domain_config", map[string]string{"domain_id": did}); !ok {
		return
	}
	var r models.DomainConfig
	if err := h.DB.Where("domain_id = ? AND cfg_group = ? AND cfg_option = ?", did, grp, opt).First(&r).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "message": "Not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": gin.H{grp: gin.H{opt: r.Value}}})
}

func (h *Hub) patchDomainConfig(c *gin.Context) {
	did := c.Param("domain_id")
	if _, ok := h.requireAuthPolicy(c, "identity:set_domain_config", map[string]string{"domain_id": did}); !ok {
		return
	}
	var body struct {
		Config map[string]any `json:"config"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.BadRequest(c, "Malformed request body")
		return
	}
	for k, v := range body.Config {
		grp, opt, ok := splitConfigKey(k)
		if !ok {
			continue
		}
		val, _ := v.(string)
		var row models.DomainConfig
		err := h.DB.Where("domain_id = ? AND cfg_group = ? AND cfg_option = ?", did, grp, opt).First(&row).Error
		if err != nil {
			_ = h.DB.Create(&models.DomainConfig{DomainID: did, Group: grp, Option: opt, Value: val})
		} else {
			row.Value = val
			_ = h.DB.Save(&row)
		}
	}
	c.Status(http.StatusNoContent)
}

func splitConfigKey(k string) (group, option string, ok bool) {
	for i := 0; i < len(k); i++ {
		if k[i] == '/' {
			return k[:i], k[i+1:], true
		}
	}
	return "", "", false
}

func (h *Hub) deleteDomainConfigOption(c *gin.Context) {
	did := c.Param("domain_id")
	grp := c.Param("group")
	opt := c.Param("option")
	if _, ok := h.requireAuthPolicy(c, "identity:delete_domain_config", map[string]string{"domain_id": did}); !ok {
		return
	}
	h.DB.Where("domain_id = ? AND cfg_group = ? AND cfg_option = ?", did, grp, opt).Delete(&models.DomainConfig{})
	c.Status(http.StatusNoContent)
}
