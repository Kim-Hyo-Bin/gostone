package token

import (
	"fmt"

	"github.com/google/uuid"
)

func unpackFernetV0Unscoped(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 5 {
		return FernetDecoded{}, fmt.Errorf("unscoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	exp := keystoneExpiresFromFloat(toFloat(raw[3]))
	return FernetDecoded{Version: 0, UserID: userID, Methods: methods, Exp: exp}, nil
}

func unpackFernetV1Domain(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 6 {
		return FernetDecoded{}, fmt.Errorf("domain-scoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	scopeDom, err := unpackDomainIDField(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	exp := keystoneExpiresFromFloat(toFloat(raw[4]))
	return FernetDecoded{Version: 1, UserID: userID, ScopeDomainID: scopeDom, Methods: methods, Exp: exp}, nil
}

func unpackFernetV2Project(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 6 {
		return FernetDecoded{}, fmt.Errorf("project payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	projectID, err := unpackUserOrString(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	exp := keystoneExpiresFromFloat(toFloat(raw[4]))
	return FernetDecoded{Version: 2, UserID: userID, ProjectID: projectID, Methods: methods, Exp: exp}, nil
}

func unpackFernetV3Trust(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 7 {
		return FernetDecoded{}, fmt.Errorf("trust-scoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	projectID, err := unpackUserOrString(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	exp := keystoneExpiresFromFloat(toFloat(raw[4]))
	trustID, err := uuidBytesToString(raw[6])
	if err != nil {
		return FernetDecoded{}, fmt.Errorf("trust_id: %w", err)
	}
	return FernetDecoded{
		Version: 3, UserID: userID, ProjectID: projectID, Methods: methods, Exp: exp,
		TrustID: trustID,
	}, nil
}

func unpackFernetV4FederatedUnscoped(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 8 {
		return FernetDecoded{}, fmt.Errorf("federated unscoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	groups, err := unpackFederatedGroupList(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	idp, err := unpackUserOrString(raw[4])
	if err != nil {
		return FernetDecoded{}, err
	}
	proto := protocolString(raw[5])
	exp := keystoneExpiresFromFloat(toFloat(raw[6]))
	return FernetDecoded{
		Version: 4, UserID: userID, Methods: methods, Exp: exp,
		FederatedGroupIDs: groups, IdentityProvider: idp, ProtocolID: proto,
	}, nil
}

func unpackFederatedScoped(raw []interface{}, authOrder []string, project, domain bool) (FernetDecoded, error) {
	if len(raw) < 9 {
		return FernetDecoded{}, fmt.Errorf("federated scoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	scopeID, err := unpackUserOrString(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	groups, err := unpackFederatedGroupList(raw[4])
	if err != nil {
		return FernetDecoded{}, err
	}
	idp, err := unpackUserOrString(raw[5])
	if err != nil {
		return FernetDecoded{}, err
	}
	proto := protocolString(raw[6])
	exp := keystoneExpiresFromFloat(toFloat(raw[7]))
	ver := 5
	if domain {
		ver = 6
	}
	fd := FernetDecoded{
		Version: ver, UserID: userID, Methods: methods, Exp: exp,
		FederatedGroupIDs: groups, IdentityProvider: idp, ProtocolID: proto,
	}
	if project {
		fd.ProjectID = scopeID
	} else {
		fd.ScopeDomainID = scopeID
	}
	return fd, nil
}

func unpackFernetV7OAuth(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 7 {
		return FernetDecoded{}, fmt.Errorf("oauth1 scoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	projectID, err := unpackUserOrString(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	accessID, err := unpackUserOrString(raw[4])
	if err != nil {
		return FernetDecoded{}, err
	}
	exp := keystoneExpiresFromFloat(toFloat(raw[5]))
	return FernetDecoded{
		Version: 7, UserID: userID, ProjectID: projectID, Methods: methods, Exp: exp,
		AccessTokenID: accessID,
	}, nil
}

func unpackFernetV8System(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 6 {
		return FernetDecoded{}, fmt.Errorf("system-scoped payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	sys := systemString(raw[3])
	exp := keystoneExpiresFromFloat(toFloat(raw[4]))
	return FernetDecoded{
		Version: 8, UserID: userID, Methods: methods, Exp: exp, SystemScope: sys,
	}, nil
}

func unpackFernetV9AppCred(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 7 {
		return FernetDecoded{}, fmt.Errorf("application credential payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	projectID, err := unpackUserOrString(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	exp := keystoneExpiresFromFloat(toFloat(raw[4]))
	appCred, err := unpackUserOrString(raw[6])
	if err != nil {
		return FernetDecoded{}, err
	}
	return FernetDecoded{
		Version: 9, UserID: userID, ProjectID: projectID, Methods: methods, Exp: exp,
		AppCredID: appCred,
	}, nil
}

func unpackFernetV10OAuth2MTLS(raw []interface{}, authOrder []string) (FernetDecoded, error) {
	if len(raw) < 8 {
		return FernetDecoded{}, fmt.Errorf("oauth2 mTLS payload too short")
	}
	userID, err := unpackUserOrString(raw[1])
	if err != nil {
		return FernetDecoded{}, err
	}
	methods := IntToMethods(authOrder, toInt(raw[2]))
	projectID, err := unpackUserOrString(raw[3])
	if err != nil {
		return FernetDecoded{}, err
	}
	domainID, err := unpackUserOrString(raw[4])
	if err != nil {
		return FernetDecoded{}, err
	}
	exp := keystoneExpiresFromFloat(toFloat(raw[5]))
	thumb, err := unpackUserOrString(raw[7])
	if err != nil {
		return FernetDecoded{}, err
	}
	return FernetDecoded{
		Version: 10, UserID: userID, ProjectID: projectID, ScopeDomainID: domainID,
		Methods: methods, Exp: exp, Thumbprint: thumb,
	}, nil
}

func unpackDomainIDField(v interface{}) (string, error) {
	// Keystone: UUID bytes or default-domain string.
	switch x := v.(type) {
	case []byte:
		if len(x) == 16 {
			u, err := uuid.FromBytes(x)
			if err == nil {
				return u.String(), nil
			}
		}
		return string(x), nil
	case string:
		return x, nil
	default:
		return unpackUserOrString(v)
	}
}

func uuidBytesToString(v interface{}) (string, error) {
	b, ok := v.([]byte)
	if !ok || len(b) != 16 {
		return "", fmt.Errorf("expected 16-byte trust uuid")
	}
	u, err := uuid.FromBytes(b)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func unpackFederatedGroupList(v interface{}) ([]string, error) {
	arr, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("federated groups not a list")
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s, err := unpackUserOrString(item)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func protocolString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(x)
	}
}

func systemString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(x)
	}
}
