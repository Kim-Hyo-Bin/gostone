package token

import (
	"encoding/base64"
	"encoding/binary"
	"strings"
	"time"

	"github.com/fernet/fernet-go"
)

// restoreFernetPadding adds base64 padding stripped by Keystone clients.
func restoreFernetPadding(token string) string {
	mod := len(token) % 4
	if mod == 0 {
		return token
	}
	return token + strings.Repeat("=", 4-mod)
}

// FernetEnvelopeIssuedAt returns the timestamp embedded in the outer Fernet wrapper (UTC).
func FernetEnvelopeIssuedAt(tokenStr string) (time.Time, error) {
	s := restoreFernetPadding(tokenStr)
	buf := make([]byte, base64.URLEncoding.DecodedLen(len(s)))
	n, err := base64.URLEncoding.Decode(buf, []byte(s))
	if err != nil {
		return time.Time{}, err
	}
	if n < 9 {
		return time.Time{}, errInvalidFernet
	}
	ts := binary.BigEndian.Uint64(buf[1:9])
	return time.Unix(int64(ts), 0).UTC(), nil
}

var errInvalidFernet = baseError("invalid fernet token")

type baseError string

func (e baseError) Error() string { return string(e) }

// fernetEncrypt strips '=' padding from the token string like Keystone.
func fernetEncrypt(plain []byte, primary *fernet.Key) (string, error) {
	tok, err := fernet.EncryptAndSign(plain, primary)
	if err != nil {
		return "", err
	}
	s := string(tok)
	return strings.TrimRight(s, "="), nil
}

func fernetDecrypt(tokenStr string, keys []*fernet.Key) ([]byte, error) {
	s := restoreFernetPadding(tokenStr)
	plain := fernet.VerifyAndDecrypt([]byte(s), 0, keys)
	if plain == nil {
		return nil, errInvalidFernet
	}
	return plain, nil
}
