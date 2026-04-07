package bootstrap

import (
	"os"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EnsureIdentityCatalog seeds a minimal identity public endpoint if the catalog is empty (idempotent).
func EnsureIdentityCatalog(db *gorm.DB) error {
	var n int64
	if err := db.Model(&models.Service{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return seedIdentityCatalog(db)
}

func seedIdentityCatalog(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return seedIdentityCatalogTx(tx)
	})
}

func seedIdentityCatalogTx(tx *gorm.DB) error {
	public := strings.TrimSpace(os.Getenv("GOSTONE_PUBLIC_URL"))
	if public == "" {
		public = "http://127.0.0.1:5000"
	}
	public = strings.TrimRight(public, "/")
	apiURL := public + "/v3"

	var n int64
	if err := tx.Model(&models.Service{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	region := models.Region{ID: "RegionOne", Description: "Default region"}
	svcID := uuid.NewString()
	svc := models.Service{
		ID:      svcID,
		Type:    "identity",
		Name:    "keystone",
		Enabled: true,
	}
	ep := models.Endpoint{
		ID:        uuid.NewString(),
		ServiceID: svcID,
		RegionID:  region.ID,
		Interface: "public",
		URL:       apiURL,
		Enabled:   true,
	}
	if err := tx.Create(&region).Error; err != nil {
		return err
	}
	if err := tx.Create(&svc).Error; err != nil {
		return err
	}
	return tx.Create(&ep).Error
}
