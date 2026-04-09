package db

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate creates or updates schema for core Identity tables.
func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.Domain{},
		&models.DomainConfig{},
		&models.Project{},
		&models.ProjectTag{},
		&models.User{},
		&models.Role{},
		&models.UserProjectRole{},
		&models.UserDomainRole{},
		&models.GroupProjectRole{},
		&models.GroupDomainRole{},
		&models.UserSystemRole{},
		&models.GroupSystemRole{},
		&models.AuthToken{},
		&models.JWTRevocation{},
		&models.Region{},
		&models.Service{},
		&models.Endpoint{},
		&models.Group{},
		&models.GroupMember{},
		&models.IdentityPolicyDoc{},
		&models.RegisteredLimit{},
		&models.Limit{},
		&models.Credential{},
		&models.EC2Credential{},
		&models.Trust{},
		&models.TrustRole{},
		&models.IdentityProvider{},
		&models.FederationMapping{},
		&models.ServiceProvider{},
		&models.OAuthConsumer{},
		&models.OAuth1AccessToken{},
		&models.ApplicationCredential{},
		&models.ApplicationCredentialRole{},
		&models.AccessRule{},
		&models.RevokeEvent{},
		&models.EndpointPolicyLink{},
		&models.ProjectEndpointFilter{},
		&models.EndpointGroup{},
		&models.EndpointGroupMember{},
		&models.EndpointGroupProject{},
	)
}
