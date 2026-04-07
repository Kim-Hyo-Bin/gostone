// Package testutil holds shared helpers for tests (DB seeding, etc.).
package testutil

import (
	"fmt"

	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Fixture holds IDs for a seeded admin user (domain Default, user admin, role admin).
type Fixture struct {
	DomainID  string
	ProjectID string
	UserID    string
	RoleID    string
}

// OpenMemory opens an isolated in-memory SQLite DB (unique DSN per call — avoids shared `file::memory:` collisions in parallel tests).
func OpenMemory() (*gorm.DB, error) {
	dsn := fmt.Sprintf("file:gostone_%s?mode=memory&cache=shared", uuid.NewString())
	gdb, err := db.Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(gdb); err != nil {
		return nil, err
	}
	return gdb, nil
}

// SeedAdmin creates domain "Default", project "admin", user "admin", role "admin", and assignment.
func SeedAdmin(gdb *gorm.DB, plainPassword string) (Fixture, error) {
	domainID := uuid.NewString()
	roleID := uuid.NewString()
	projectID := uuid.NewString()
	userID := uuid.NewString()

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	if err != nil {
		return Fixture{}, err
	}

	dom := models.Domain{ID: domainID, Name: "Default", Enabled: true}
	role := models.Role{ID: roleID, Name: "admin", DomainID: ""}
	proj := models.Project{ID: projectID, DomainID: domainID, Name: "admin", Enabled: true}
	user := models.User{
		ID: userID, DomainID: domainID, Name: "admin", Enabled: true,
		PasswordHash: string(hash),
	}
	assign := models.UserProjectRole{UserID: userID, RoleID: roleID, ProjectID: projectID}

	err = gdb.Transaction(func(tx *gorm.DB) error {
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
	if err != nil {
		return Fixture{}, err
	}
	return Fixture{
		DomainID:  domainID,
		ProjectID: projectID,
		UserID:    userID,
		RoleID:    roleID,
	}, nil
}
