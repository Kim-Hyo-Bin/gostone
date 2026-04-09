package conf

import "testing"

func TestListenBindings_singleDefault(t *testing.T) {
	s := &Service{Listen: ":5000"}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1 || b[0].Name != "listen" || b[0].Addr != ":5000" {
		t.Fatalf("%+v", b)
	}
}

func TestListenBindings_singleEmptyListen(t *testing.T) {
	s := &Service{}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1 || b[0].Addr != ":5000" {
		t.Fatalf("%+v", b)
	}
}

func TestListenBindings_multiPublicOnly(t *testing.T) {
	s := &Service{ListenPublic: ":5000"}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1 || b[0].Name != "public" || b[0].Addr != ":5000" {
		t.Fatalf("%+v", b)
	}
}

func TestListenBindings_multiPublicAdmin(t *testing.T) {
	s := &Service{ListenPublic: ":5000", ListenAdmin: "127.0.0.1:35357"}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 2 {
		t.Fatalf("%+v", b)
	}
	if b[0].Name != "public" || b[0].Addr != ":5000" {
		t.Fatalf("%+v", b[0])
	}
	if b[1].Name != "admin" || b[1].Addr != "127.0.0.1:35357" {
		t.Fatalf("%+v", b[1])
	}
}

func TestListenBindings_multiFallbackPublicFromListen(t *testing.T) {
	s := &Service{Listen: ":5000", ListenAdmin: "127.0.0.1:35357"}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 2 || b[0].Name != "public" || b[0].Addr != ":5000" || b[1].Name != "admin" {
		t.Fatalf("%+v", b)
	}
}

func TestListenBindings_multiAdminOnlyNoPublic(t *testing.T) {
	s := &Service{ListenAdmin: ":35357"}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1 || b[0].Name != "admin" {
		t.Fatalf("%+v", b)
	}
}

func TestListenBindings_multiInternalOnly(t *testing.T) {
	s := &Service{ListenInternal: "10.0.0.1:5000"}
	b, err := ListenBindings(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1 || b[0].Name != "internal" {
		t.Fatalf("%+v", b)
	}
}
