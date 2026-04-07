#!/usr/bin/env bash
# Bring up packaged OpenStack Keystone (docker-compose.upstream-keystone.yml), wait until
# the API answers, then run shell + Go smoke tests (same as manual test-upstream-keystone.sh).
#
# Usage:
#   ./scripts/integration-upstream-keystone.sh
#   ./scripts/integration-upstream-keystone.sh --fresh   # down -v first
#
# Requires: docker compose, curl, python3, go
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
COMPOSE="$ROOT/docker-compose.upstream-keystone.yml"

for cmd in docker curl python3 go; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "Missing required command: $cmd" >&2
		exit 1
	fi
done

FRESH=0
if [[ "${1:-}" == "--fresh" ]]; then
	FRESH=1
fi

if [[ "$FRESH" -eq 1 ]]; then
	echo "Removing upstream Keystone stack and volumes (--fresh)..."
	docker compose -f "$COMPOSE" down -v 2>/dev/null || true
fi

# Fixed container_name in the compose file can survive a project-name change; remove stale ones.
docker rm -f keystone-upstream-mariadb keystone-upstream-api 2>/dev/null || true

echo "Starting upstream MariaDB + Keystone..."
docker compose -f "$COMPOSE" up -d --build

BASE_URL="${KEYSTONE_UPSTREAM_URL:-http://127.0.0.1:15000}"
BASE_URL="${BASE_URL%/}"

echo "Waiting for Identity API at ${BASE_URL}/v3 ..."
for i in $(seq 1 90); do
	if curl -sf "${BASE_URL}/v3" >/dev/null 2>&1; then
		echo "Keystone responded."
		break
	fi
	if [[ "$i" -eq 90 ]]; then
		echo "Keystone did not become ready in time." >&2
		docker compose -f "$COMPOSE" logs keystone_upstream 2>&1 | tail -120 >&2
		exit 1
	fi
	sleep 2
done

export KEYSTONE_UPSTREAM_URL="$BASE_URL"
export KEYSTONE_ADMIN_PASSWORD="${KEYSTONE_ADMIN_PASSWORD:-admin}"

echo "Running scripts/test-upstream-keystone.sh ..."
"$ROOT/scripts/test-upstream-keystone.sh"

echo "Running go test -tags=upstream ./internal/integration/..."
go test -tags=upstream -count=1 -v ./internal/integration/...

echo "OK: upstream Keystone integration checks passed."
