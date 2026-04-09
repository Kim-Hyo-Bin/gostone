#!/usr/bin/env bash
# Start MariaDB (docker compose), wait until healthy, run integration tests against DB.
# Usage:
#   ./scripts/integration-mariadb.sh              # reuse existing data volume
#   ./scripts/integration-mariadb.sh --fresh      # docker compose down -v first
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FRESH=0
if [[ "${1:-}" == "--fresh" ]]; then
  FRESH=1
fi

if [[ "$FRESH" -eq 1 ]]; then
  echo "Removing stack and volumes (--fresh)..."
  docker compose down -v 2>/dev/null || true
fi

echo "Starting MariaDB..."
docker compose up -d mariadb

echo "Waiting for MariaDB health..."
for i in $(seq 1 60); do
  if docker compose exec -T mariadb healthcheck.sh --connect --innodb_initialized 2>/dev/null; then
    break
  fi
  if [[ "$i" -eq 60 ]]; then
    echo "MariaDB did not become healthy in time." >&2
    docker compose logs mariadb | tail -80 >&2
    exit 1
  fi
  sleep 2
done

# Integration DB (empty each --fresh run; otherwise tests tolerate existing admin)
export GOSTONE_DATABASE_CONNECTION="${GOSTONE_DATABASE_CONNECTION:-mysql+pymysql://keystone:keystonepass@127.0.0.1:13306/gostone_integration}"
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD="${GOSTONE_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
export GOSTONE_TOKEN_PROVIDER="${GOSTONE_TOKEN_PROVIDER:-jwt}"
export GOSTONE_TOKEN_SECRET="${GOSTONE_TOKEN_SECRET:-integration-test-jwt-secret}"

echo "Running integration tests with:"
echo "  GOSTONE_DATABASE_CONNECTION=$GOSTONE_DATABASE_CONNECTION"

go test -tags=integration -count=1 -v ./internal/integration/...

echo "OK: integration tests passed."
