package bootstrap

import (
	"errors"
	"fmt"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Options configures keystone-manage bootstrap–style initial identity data.
type Options struct {
	AdminPassword string
	AdminUsername string
	DomainName    string
	ProjectName   string
	RegionID      string
	AdminRoleName string
}

// DefaultBootstrapOptions matches common Keystone bootstrap defaults.
// RegionID is empty so catalog seeding can use [service] region_id from config (else RegionOne).
func DefaultBootstrapOptions() Options {
	return Options{
		AdminUsername: "admin",
		DomainName:    "Default",
		ProjectName:   "admin",
		RegionID:      "",
		AdminRoleName: "admin",
	}
}

// RunIfEmpty seeds domain, project, role, and admin user only when the user table is empty.
// Used on server startup via FromEnv; duplicate calls are no-ops.
func RunIfEmpty(db *gorm.DB, o Options) error {
	if o.AdminPassword == "" {
		return nil
	}
	var n int64
	if err := db.Model(&models.User{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return runBootstrapTx(db, o)
}

// RunAdmin seeds the initial admin when no users exist; returns an error if any user already exists
// (keystone-manage bootstrap refuses a non-empty database).
func RunAdmin(db *gorm.DB, o Options) error {
	if o.AdminPassword == "" {
		return errors.New("bootstrap admin password is empty")
	}
	var n int64
	if err := db.Model(&models.User{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("refusing bootstrap: database already has %d user(s)", n)
	}
	return runBootstrapTx(db, o)
}

func runBootstrapTx(db *gorm.DB, o Options) error {
	if o.AdminUsername == "" {
		o.AdminUsername = "admin"
	}
	if o.DomainName == "" {
		o.DomainName = "Default"
	}
	if o.ProjectName == "" {
		o.ProjectName = "admin"
	}
	if o.AdminRoleName == "" {
		o.AdminRoleName = "admin"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(o.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	domainID := uuid.NewString()
	roleID := uuid.NewString()
	projectID := uuid.NewString()
	userID := uuid.NewString()

	dom := models.Domain{ID: domainID, Name: o.DomainName, Enabled: true}
	role := models.Role{ID: roleID, Name: o.AdminRoleName, DomainID: ""}
	proj := models.Project{ID: projectID, DomainID: domainID, Name: o.ProjectName, Enabled: true}
	user := models.User{
		ID: userID, DomainID: domainID, Name: o.AdminUsername, Enabled: true,
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
