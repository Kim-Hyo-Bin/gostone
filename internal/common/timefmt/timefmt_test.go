package timefmt

import (
	"testing"
	"time"
)

func TestKeystoneUTC_truncatesToMicroseconds(t *testing.T) {
	in := time.Date(2026, 4, 10, 6, 48, 22, 967221009, time.UTC)
	got := KeystoneUTC(in)
	want := "2026-04-10T06:48:22.967221Z"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
