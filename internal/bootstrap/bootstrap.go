package bootstrap

import (
	"os"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// FromEnv seeds an initial domain, project, admin user, and role when the DB is empty
// and GOSTONE_BOOTSTRAP_ADMIN_PASSWORD is set (development / first boot).
func FromEnv(db *gorm.DB) error {
	pw := os.Getenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD")
	if pw == "" {
		return nil
	}
	var n int64
	if err := db.Model(&models.User{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	domainID := uuid.NewString()
	roleID := uuid.NewString()
	projectID := uuid.NewString()
	userID := uuid.NewString()

	dom := models.Domain{ID: domainID, Name: "Default", Enabled: true}
	role := models.Role{ID: roleID, Name: "admin", DomainID: ""}
	proj := models.Project{ID: projectID, DomainID: domainID, Name: "admin", Enabled: true}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{
		ID: userID, DomainID: domainID, Name: "admin", Enabled: true,
		PasswordHash: string(hash),
	}
	assign := models.UserProjectRole{UserID: userID, RoleID: roleID, ProjectID: projectID}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&dom).Error; err != nil {
			return err
		}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		if err := tx.Create(&proj).Error; err != nil {
			return err
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Create(&assign).Error
	})
}
