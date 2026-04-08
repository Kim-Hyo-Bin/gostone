package bootstrap

import (
	"os"

	"gorm.io/gorm"
)

// FromEnv seeds an initial domain, project, admin user, and role when the DB is empty
// and GOSTONE_BOOTSTRAP_ADMIN_PASSWORD is set (development / first boot).
func FromEnv(db *gorm.DB) error {
	pw := os.Getenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD")
	if pw == "" {
		return nil
	}
	o := DefaultBootstrapOptions()
	o.AdminPassword = pw
	return RunIfEmpty(db, o)
}
