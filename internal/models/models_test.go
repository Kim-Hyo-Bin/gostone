package models

import "testing"

func TestModelFieldRoundTrip(t *testing.T) {
	u := User{ID: "u1", DomainID: "d1", Name: "alice", Enabled: true, PasswordHash: "h"}
	if u.ID != "u1" || !u.Enabled {
		t.Fatalf("%+v", u)
	}
	d := Domain{ID: "d1", Name: "Default", Enabled: true}
	if d.Name != "Default" {
		t.Fatal(d)
	}
}
