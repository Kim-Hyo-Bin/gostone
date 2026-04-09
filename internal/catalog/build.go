// Package catalog builds Keystone-style service catalog JSON from the database.
package catalog

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// Build returns the "catalog" array for a token response (OpenStack Identity v3 shape).
func Build(db *gorm.DB) ([]any, error) {
	var services []models.Service
	if err := db.Where("enabled = ?", true).Order("type, name").Find(&services).Error; err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return []any{}, nil
	}

	var endpoints []models.Endpoint
	if err := db.Where("enabled = ?", true).Find(&endpoints).Error; err != nil {
		return nil, err
	}
	epByService := make(map[string][]models.Endpoint)
	for _, e := range endpoints {
		epByService[e.ServiceID] = append(epByService[e.ServiceID], e)
	}

	out := make([]any, 0, len(services))
	for _, s := range services {
		eps := epByService[s.ID]
		epObjs := make([]map[string]any, 0, len(eps))
		for _, e := range eps {
			epObjs = append(epObjs, map[string]any{
				"id":        e.ID,
				"interface": e.Interface,
				"region_id": e.RegionID,
				// Keystone and Tempest expect "region" (same value as region_id) for catalog filtering.
				"region": e.RegionID,
				"url":    e.URL,
			})
		}
		out = append(out, map[string]any{
			"id":          s.ID,
			"type":        s.Type,
			"name":        s.Name,
			"enabled":     s.Enabled,
			"endpoints":   epObjs,
			"description": s.Description,
		})
	}
	return out, nil
}
