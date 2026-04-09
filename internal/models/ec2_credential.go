package models

// EC2Credential stores EC2-style access key material for a user (secret stored for signature checks).
type EC2Credential struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"not null;index;size:64"`
	AccessKey string `gorm:"not null;uniqueIndex;size:64"`
	SecretKey string `gorm:"not null;type:text"`
	TrustID   string `gorm:"type:text"`
}
