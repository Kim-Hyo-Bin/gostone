package bootstrap

import (
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
)

func TestEnsureIdentityCatalog_adminInternalEndpoints(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureIdentityCatalog(gdb,
		"http://pub.example:5000",
		"http://adm.example:35357",
		"http://int.example:5000",
		"R1",
	); err != nil {
		t.Fatal(err)
	}
	var eps []models.Endpoint
	if err := gdb.Order("interface").Find(&eps).Error; err != nil {
		t.Fatal(err)
	}
	if len(eps) != 3 {
		t.Fatalf("endpoints %d", len(eps))
	}
	want := map[string]string{
		"admin":    "http://adm.example:35357/v3",
		"internal": "http://int.example:5000/v3",
		"public":   "http://pub.example:5000/v3",
	}
	for _, e := range eps {
		u, ok := want[e.Interface]
		if !ok || e.URL != u {
			t.Fatalf("%s -> %q", e.Interface, e.URL)
		}
	}
}
