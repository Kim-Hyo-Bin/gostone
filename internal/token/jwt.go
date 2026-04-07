package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT issues and parses bearer tokens (Keystone-shaped claims; not Fernet yet).
type JWT struct {
	Secret []byte
	Issuer string
	TTL    time.Duration
}

// Claims carried inside the JWT.
type Claims struct {
	UserID    string   `json:"uid"`
	DomainID  string   `json:"dom"`
	ProjectID string   `json:"prj"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

// Issue returns a signed token and its expiry instant.
func (j *JWT) Issue(userID, domainID, projectID string, roles []string) (string, time.Time, error) {
	if len(j.Secret) == 0 {
		return "", time.Time{}, errors.New("token secret is empty")
	}
	now := time.Now()
	exp := now.Add(j.TTL)
	claims := Claims{
		UserID:    userID,
		DomainID:  domainID,
		ProjectID: projectID,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.Issuer,
			ID:        uuid.NewString(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(j.Secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return s, exp, nil
}

// Parse validates a token string and returns claims.
func (j *JWT) Parse(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return j.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
