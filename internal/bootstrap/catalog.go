package bootstrap

import (
	"os"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EnsureIdentityCatalog seeds a minimal identity service with public (and optional admin/internal)
// endpoints if the catalog is empty (idempotent).
// publicBaseURL is optional (scheme://host:port); if empty, GOSTONE_PUBLIC_URL or a default is used.
// adminBaseURL and internalBaseURL add catalog endpoints with those interfaces when non-empty.
// regionID defaults to RegionOne when empty.
func EnsureIdentityCatalog(db *gorm.DB, publicBaseURL, adminBaseURL, internalBaseURL, regionID string) error {
	var n int64
	if err := db.Model(&models.Service{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return seedIdentityCatalogTx(tx, publicBaseURL, adminBaseURL, internalBaseURL, regionID)
	})
}

func seedIdentityCatalogTx(tx *gorm.DB, publicBaseURL, adminBaseURL, internalBaseURL, regionID string) error {
	if regionID == "" {
		regionID = "RegionOne"
	}
	public := strings.TrimSpace(publicBaseURL)
	if public == "" {
		public = strings.TrimSpace(os.Getenv("GOSTONE_PUBLIC_URL"))
	}
	if public == "" {
		public = "http://127.0.0.1:5000"
	}
	public = strings.TrimRight(public, "/")

	var n int64
	if err := tx.Model(&models.Service{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	region := models.Region{ID: regionID, Description: "Default region"}
	svcID := uuid.NewString()
	svc := models.Service{
		ID:      svcID,
		Type:    "identity",
		Name:    "keystone",
		Enabled: true,
	}
	if err := tx.Create(&region).Error; err != nil {
		return err
	}
	if err := tx.Create(&svc).Error; err != nil {
		return err
	}

	type ifaceURL struct {
		iface string
		base  string
	}
	var pairs []ifaceURL
	pairs = append(pairs, ifaceURL{"public", public})
	if b := strings.TrimSpace(adminBaseURL); b != "" {
		pairs = append(pairs, ifaceURL{"admin", strings.TrimRight(b, "/")})
	}
	if b := strings.TrimSpace(internalBaseURL); b != "" {
		pairs = append(pairs, ifaceURL{"internal", strings.TrimRight(b, "/")})
	}
	for _, p := range pairs {
		if p.base == "" {
			continue
		}
		ep := models.Endpoint{
			ID:        uuid.NewString(),
			ServiceID: svcID,
			RegionID:  region.ID,
			Interface: p.iface,
			URL:       p.base + "/v3",
			Enabled:   true,
		}
		if err := tx.Create(&ep).Error; err != nil {
			return err
		}
	}
	return nil
}
