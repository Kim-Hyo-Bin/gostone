package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

func packUserBytes(id string) []interface{} {
	u := uuid.MustParse(id)
	b, _ := u.MarshalBinary()
	return []interface{}{true, b}
}

func TestUnpackKeystoneFernet_v3Trust(t *testing.T) {
	uid := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	pid := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	tid := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	raw := []interface{}{
		3,
		packUserBytes(uid),
		MethodsToInt(DefaultAuthMethods(), []string{"password"}),
		packUserBytes(pid),
		keystoneExpiresFloat(time.Now().Add(time.Hour)),
		[]interface{}{},
		tid[:],
	}
	plain, err := msgpack.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := UnpackKeystoneFernetPayload(plain, DefaultAuthMethods())
	if err != nil {
		t.Fatal(err)
	}
	if fd.Version != 3 || fd.UserID != uid || fd.ProjectID != pid || fd.TrustID != tid.String() {
		t.Fatalf("%+v", fd)
	}
}

func TestUnpackKeystoneFernet_v8System(t *testing.T) {
	uid := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	raw := []interface{}{
		8,
		packUserBytes(uid),
		MethodsToInt(DefaultAuthMethods(), []string{"password"}),
		"all",
		keystoneExpiresFloat(time.Now().Add(time.Hour)),
		[]interface{}{},
	}
	plain, err := msgpack.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := UnpackKeystoneFernetPayload(plain, DefaultAuthMethods())
	if err != nil {
		t.Fatal(err)
	}
	if fd.Version != 8 || fd.UserID != uid || fd.SystemScope != "all" {
		t.Fatalf("%+v", fd)
	}
}

func TestUnpackKeystoneFernet_v10OAuth2MTLS(t *testing.T) {
	uid := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	prj := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	dom := "11111111-1111-1111-1111-111111111111"
	raw := []interface{}{
		10,
		packUserBytes(uid),
		MethodsToInt(DefaultAuthMethods(), []string{"password"}),
		packUserBytes(prj),
		packUserBytes(dom),
		keystoneExpiresFloat(time.Now().Add(time.Hour)),
		[]interface{}{},
		[]interface{}{false, "sha256-thumb"},
	}
	plain, err := msgpack.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := UnpackKeystoneFernetPayload(plain, DefaultAuthMethods())
	if err != nil {
		t.Fatal(err)
	}
	if fd.Version != 10 || fd.Thumbprint != "sha256-thumb" || fd.ProjectID != prj || fd.ScopeDomainID != dom {
		t.Fatalf("%+v", fd)
	}
}
