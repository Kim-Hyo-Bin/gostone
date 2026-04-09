package token

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

// MethodsToInt maps auth method names to Keystone’s bitmask (CONF.auth.methods order).
func MethodsToInt(order []string, methods []string) int {
	sum := 0
	for i, name := range order {
		bit := 1 << i
		for _, m := range methods {
			if m == name {
				sum += bit
				break
			}
		}
	}
	return sum
}

// IntToMethods expands a Keystone method bitmask using the same order slice.
func IntToMethods(order []string, methodInt int) []string {
	if methodInt == 0 {
		return nil
	}
	var out []string
	for i, name := range order {
		bit := 1 << i
		if methodInt&bit != 0 {
			out = append(out, name)
		}
	}
	return out
}

func keystoneExpiresFloat(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}

func keystoneExpiresFromFloat(f float64) time.Time {
	if f <= 0 {
		return time.Time{}
	}
	sec, frac := math.Modf(f)
	ns := int64(math.Round(frac * 1e9))
	return time.Unix(int64(sec), ns).UTC()
}

func auditStringsToBytes(auditIDs []string) ([][]byte, error) {
	out := make([][]byte, 0, len(auditIDs))
	for _, s := range auditIDs {
		b, err := keystoneAuditDecode(s)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func keystoneAuditDecode(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty audit id")
	}
	// Keystone random_urlsafe_str_to_bytes adds padding.
	mod := len(s) % 4
	if mod != 0 {
		s += strings.Repeat("=", 4-mod)
	}
	return base64.URLEncoding.DecodeString(s)
}

func keystoneAuditEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func packUserOrString(userID string) ([]interface{}, error) {
	u, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return []interface{}{false, userID}, nil
	}
	ub, err := u.MarshalBinary()
	if err != nil {
		return []interface{}{false, userID}, nil
	}
	return []interface{}{true, ub}, nil
}

func unpackUserOrString(v interface{}) (string, error) {
	arr, ok := v.([]interface{})
	if !ok || len(arr) < 2 {
		return "", fmt.Errorf("invalid user_id payload")
	}
	flag, ok := arr[0].(bool)
	if !ok {
		return "", fmt.Errorf("invalid user_id flag")
	}
	switch val := arr[1].(type) {
	case []byte:
		if flag {
			u, err := uuid.FromBytes(val)
			if err != nil {
				return "", err
			}
			return u.String(), nil
		}
		return string(val), nil
	case string:
		return val, nil
	default:
		return "", fmt.Errorf("invalid user_id value type")
	}
}

// PackKeystoneFernetUnscoped builds msgpack payload for Keystone UnscopedPayload (version 0).
func PackKeystoneFernetUnscoped(userID string, methods []string, exp time.Time, auditIDs []string, authOrder []string) ([]byte, error) {
	u0, err := packUserOrString(userID)
	if err != nil {
		return nil, err
	}
	audBytes, err := auditStringsToBytes(auditIDs)
	if err != nil {
		return nil, err
	}
	mi := MethodsToInt(authOrder, methods)
	expF := keystoneExpiresFloat(exp)
	// audit_ids as list of binaries
	audList := make([]interface{}, len(audBytes))
	for i, b := range audBytes {
		audList[i] = b
	}
	ver := []interface{}{
		0,
		u0,
		mi,
		expF,
		audList,
	}
	return msgpack.Marshal(ver)
}

// PackKeystoneFernetDomainScoped builds msgpack for Keystone DomainScopedPayload (version 1).
func PackKeystoneFernetDomainScoped(userID, domainID string, methods []string, exp time.Time, auditIDs []string, authOrder []string) ([]byte, error) {
	u0, err := packUserOrString(userID)
	if err != nil {
		return nil, err
	}
	d0, err := packUserOrString(domainID)
	if err != nil {
		return nil, err
	}
	audBytes, err := auditStringsToBytes(auditIDs)
	if err != nil {
		return nil, err
	}
	mi := MethodsToInt(authOrder, methods)
	expF := keystoneExpiresFloat(exp)
	audList := make([]interface{}, len(audBytes))
	for i, b := range audBytes {
		audList[i] = b
	}
	ver := []interface{}{
		1,
		u0,
		mi,
		d0,
		expF,
		audList,
	}
	return msgpack.Marshal(ver)
}

// PackKeystoneFernetProjectScoped builds msgpack for ProjectScopedPayload (version 2).
func PackKeystoneFernetProjectScoped(userID, projectID string, methods []string, exp time.Time, auditIDs []string, authOrder []string) ([]byte, error) {
	u0, err := packUserOrString(userID)
	if err != nil {
		return nil, err
	}
	p0, err := packUserOrString(projectID)
	if err != nil {
		return nil, err
	}
	audBytes, err := auditStringsToBytes(auditIDs)
	if err != nil {
		return nil, err
	}
	mi := MethodsToInt(authOrder, methods)
	expF := keystoneExpiresFloat(exp)
	audList := make([]interface{}, len(audBytes))
	for i, b := range audBytes {
		audList[i] = b
	}
	ver := []interface{}{
		2,
		u0,
		mi,
		p0,
		expF,
		audList,
	}
	return msgpack.Marshal(ver)
}

// FernetDecoded is the scope extracted from a decrypted Keystone Fernet payload (all supported versions).
type FernetDecoded struct {
	Version int
	// Core identity / scope (versions 0–2, 5–7, 9–10)
	UserID        string
	ProjectID     string
	ScopeDomainID string
	Methods       []string
	Exp           time.Time

	TrustID           string
	SystemScope       string
	AppCredID         string
	AccessTokenID     string
	Thumbprint        string
	FederatedGroupIDs []string
	IdentityProvider  string
	ProtocolID        string
}

// UnpackKeystoneFernetPayload decodes Keystone msgpack (version prefix + payload tuple).
// Layout matches openstack/keystone token_formatters.py _PAYLOAD_CLASSES.
func UnpackKeystoneFernetPayload(plain []byte, authOrder []string) (FernetDecoded, error) {
	var raw []interface{}
	if err := msgpack.Unmarshal(plain, &raw); err != nil {
		return FernetDecoded{}, fmt.Errorf("msgpack: %w", err)
	}
	if len(raw) < 2 {
		return FernetDecoded{}, fmt.Errorf("payload too short")
	}
	ver := toInt(raw[0])
	switch ver {
	case 0:
		return unpackFernetV0Unscoped(raw, authOrder)
	case 1:
		return unpackFernetV1Domain(raw, authOrder)
	case 2:
		return unpackFernetV2Project(raw, authOrder)
	case 3:
		return unpackFernetV3Trust(raw, authOrder)
	case 4:
		return unpackFernetV4FederatedUnscoped(raw, authOrder)
	case 5:
		return unpackFederatedScoped(raw, authOrder, true, false)
	case 6:
		return unpackFederatedScoped(raw, authOrder, false, true)
	case 7:
		return unpackFernetV7OAuth(raw, authOrder)
	case 8:
		return unpackFernetV8System(raw, authOrder)
	case 9:
		return unpackFernetV9AppCred(raw, authOrder)
	case 10:
		return unpackFernetV10OAuth2MTLS(raw, authOrder)
	default:
		return FernetDecoded{}, fmt.Errorf("unsupported keystone fernet payload version %d", ver)
	}
}

func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int8:
		return int(x)
	case int16:
		return int(x)
	case int32:
		return int(x)
	case int64:
		return int(x)
	case uint:
		return int(x)
	case uint8:
		return int(x)
	case uint16:
		return int(x)
	case uint32:
		return int(x)
	case uint64:
		return int(x)
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return 0
	}
}
