package token

import (
	"errors"
	"fmt"
	"strings"
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

// Issue returns a signed token, the issued-at instant (same as embedded "iat"), and expiry.
func (j *JWT) Issue(s TokenSubject) (string, time.Time, time.Time, error) {
	if err := s.validate(); err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	if len(j.Secret) == 0 {
		return "", time.Time{}, time.Time{}, errors.New("token secret is empty")
	}
	// Match jwt.RegisteredClaims encoding (jwt.NewNumericDate truncates to jwt.TimePrecision, default 1s).
	prec := jwt.TimePrecision
	now := time.Now().UTC().Truncate(prec)
	exp := now.Add(j.TTL).Truncate(prec)
	methods := s.normalizedMethods()
	jti := strings.TrimSpace(s.JTI)
	if jti == "" {
		jti = uuid.NewString()
	}
	claims := Claims{
		UserID:             s.UserID,
		DomainID:           s.DomainID,
		ProjectID:          s.ProjectID,
		ScopeDomainID:      s.ScopeDomainID,
		Roles:              s.Roles,
		Methods:            methods,
		TrustID:            s.TrustID,
		SystemScope:        s.SystemScope,
		AppCredID:          s.AppCredID,
		AccessTokenID:      s.AccessTokenID,
		Thumbprint:         s.Thumbprint,
		IdentityProviderID: s.IdentityProviderID,
		ProtocolID:         s.ProtocolID,
		FederatedGroupIDs:  append([]string(nil), s.FederatedGroupIDs...),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.Issuer,
			ID:        jti,
		},
	}
	jwtTok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	tokStr, err := jwtTok.SignedString(j.Secret)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	return tokStr, now, exp, nil
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
