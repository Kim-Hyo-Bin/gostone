package policy

import (
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
)

func TestDefault(t *testing.T) {
	p := Default()
	if p.Rules["identity:list_users"] == "" {
		t.Fatal("missing rule")
	}
}

func TestAllow_listUsers_adminOnly(t *testing.T) {
	p := Default()
	admin := auth.Context{UserID: "a", Roles: []string{"admin"}}
	user := auth.Context{UserID: "u", Roles: []string{"member"}}
	if !p.Allow("identity:list_users", admin, nil) {
		t.Fatal("admin should list")
	}
	if p.Allow("identity:list_users", user, nil) {
		t.Fatal("member should not list")
	}
}

func TestAllow_getUser_selfOrAdmin(t *testing.T) {
	p := Default()
	admin := auth.Context{UserID: "a", Roles: []string{"admin"}}
	self := auth.Context{UserID: "u1", Roles: []string{}}
	other := auth.Context{UserID: "u2", Roles: []string{}}

	if !p.Allow("identity:get_user", admin, map[string]string{"user_id": "any"}) {
		t.Fatal("admin")
	}
	if !p.Allow("identity:get_user", self, map[string]string{"user_id": "u1"}) {
		t.Fatal("self")
	}
	if p.Allow("identity:get_user", other, map[string]string{"user_id": "u1"}) {
		t.Fatal("other user")
	}
}

func TestAllow_unknownAction_authenticated(t *testing.T) {
	p := Default()
	ctx := auth.Context{UserID: "x"}
	if !p.Allow("identity:unknown_action", ctx, nil) {
		t.Fatal("default authenticated")
	}
	if p.Allow("identity:unknown_action", auth.Context{}, nil) {
		t.Fatal("anon")
	}
}

func TestEvalOr_nilTarget(t *testing.T) {
	p := Default()
	if !p.evalOr("user_match or role:admin", auth.Context{UserID: "1", Roles: []string{"admin"}}, nil, 0) {
		t.Fatal("nil target should become empty map")
	}
}
