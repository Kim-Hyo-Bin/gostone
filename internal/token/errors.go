package token

import "errors"

// ErrFernetShadowMissing is returned when a Fernet token decrypts but no auth_tokens
// shadow row exists (required when the manager has a database).
var ErrFernetShadowMissing = errors.New("fernet token metadata missing")
