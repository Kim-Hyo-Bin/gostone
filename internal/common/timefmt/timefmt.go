package timefmt

import "time"

// KeystoneUTC formats t as ISO8601 UTC with exactly six fractional digits.
// OpenStack Tempest and many Python clients parse this with strptime %f, which accepts at most six digits.
func KeystoneUTC(t time.Time) string {
	u := t.UTC().Truncate(time.Microsecond)
	return u.Format("2006-01-02T15:04:05.000000") + "Z"
}
