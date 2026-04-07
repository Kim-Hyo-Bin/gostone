// Package integration holds optional database integration tests.
//
// They are not run by default. Use:
//
//	go test -tags=integration -v ./internal/integration/...
//
// with GOSTONE_DATABASE_CONNECTION set (see scripts/integration-mariadb.sh).
//
// Upstream OpenStack Keystone (HTTP only, no gostone binary):
//
//	go test -tags=upstream -count=1 -v ./internal/integration/...
//
// with KEYSTONE_UPSTREAM_URL set (e.g. http://127.0.0.1:15000) and the stack running.
// Convenience: scripts/integration-upstream-keystone.sh starts compose, waits for /v3,
// then runs scripts/test-upstream-keystone.sh and the upstream Go tests.
//
// Policy: packaged Keystone usually denies [identity:list_users] on an unscoped token;
// the upstream tests include one case that expects 403 and another that uses a
// project-scoped token (admin project) for a successful list.
package integration
