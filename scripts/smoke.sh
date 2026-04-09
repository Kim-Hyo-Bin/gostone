#!/usr/bin/env bash
# Local deployment smoke: build binary, temp SQLite DB, start server, curl health/ready/auth.
# Do not run `gostone` with no arguments in a foreground shell without env — it starts the HTTP server.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

BIN="$(mktemp "${TMPDIR:-/tmp}/gostone-smoke-bin-XXXXXX")"
DBFILE="$(mktemp "${TMPDIR:-/tmp}/gostone-smoke-db-XXXXXX.db")"
PORT="${GOSTONE_SMOKE_PORT:-18080}"
PID=""
cleanup() {
  rm -f "$BIN"
  if [[ -n "${PID}" ]]; then
    kill "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
  fi
  rm -f "$DBFILE"
}
trap cleanup EXIT

go build -o "$BIN" ./cmd/gostone

export GOSTONE_DATABASE_CONNECTION="file:${DBFILE}"
export GOSTONE_HTTP_ADDR="127.0.0.1:${PORT}"
export GOSTONE_PUBLIC_URL="http://127.0.0.1:${PORT}"
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD="${GOSTONE_BOOTSTRAP_ADMIN_PASSWORD:-smoke-admin}"
export GOSTONE_TOKEN_PROVIDER="${GOSTONE_TOKEN_PROVIDER:-jwt}"
export GOSTONE_TOKEN_SECRET="${GOSTONE_TOKEN_SECRET:-smoke-test-jwt-secret-min-32-chars!!}"

"$BIN" &
PID=$!
# Server prints "listening" after DB migrate + bootstrap; avoid racing the first curl.
sleep 0.5

ok=0
for _ in $(seq 1 60); do
  if curl -sfS "http://127.0.0.1:${PORT}/health" >/dev/null; then
    ok=1
    break
  fi
  sleep 0.25
done
if [[ "$ok" -ne 1 ]]; then
  echo "smoke: /health did not become ready" >&2
  exit 1
fi

curl -sfS "http://127.0.0.1:${PORT}/ready" >/dev/null

HDRS="$(mktemp "${TMPDIR:-/tmp}/gostone-smoke-hdr-XXXXXX")"
curl -sfS -D "$HDRS" -o /dev/null "http://127.0.0.1:${PORT}/health" || true
if ! grep -qi '^X-OpenStack-Request-Id:' "$HDRS"; then
  echo "smoke: missing X-OpenStack-Request-Id on /health" >&2
  exit 1
fi
rm -f "$HDRS"

LOGIN=$(printf '%s' '{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"'"${GOSTONE_BOOTSTRAP_ADMIN_PASSWORD}"'","domain":{"name":"Default"}}}}}}')
OUT="$(mktemp "${TMPDIR:-/tmp}/gostone-smoke-auth-XXXXXX")"
CODE=$(curl -sS -o "$OUT" -w '%{http_code}' -X POST "http://127.0.0.1:${PORT}/v3/auth/tokens" \
  -H 'Content-Type: application/json' \
  -d "$LOGIN")
if [[ "$CODE" != "201" ]]; then
  echo "smoke: POST /v3/auth/tokens expected 201, got ${CODE}: $(cat "$OUT")" >&2
  exit 1
fi
rm -f "$OUT"

echo "smoke: OK (health, ready, request id header, password auth)"
