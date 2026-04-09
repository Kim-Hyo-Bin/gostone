package models

// IdentityPolicyDoc is a stored policy blob (JSON rule set) for the Identity API.
type IdentityPolicyDoc struct {
	ID    string `gorm:"primaryKey;size:64"`
	Blob  string `gorm:"not null;type:text"`
	Type  string `gorm:"type:text"` // e.g. application/json
	Extra string `gorm:"type:text"`
}
