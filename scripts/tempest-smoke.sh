#!/usr/bin/env bash
# Run a minimal Tempest smoke against a local gostone (SQLite, JWT).
# Requires: Python 3, pip packages from ci/requirements-tempest.txt
#
# This script isolates listen URL and Tempest uri_v3 from a developer shell:
# - Does NOT inherit GOSTONE_HTTP_ADDR / GOSTONE_PUBLIC_URL (they must match TEMPEST_SMOKE_PORT).
# - Uses -config-file with a tiny INI so ./gostone.conf and GOSTONE_CONFIG are not loaded.
# - Clears GOSTONE_LISTEN_* / GOSTONE_CONFIG_DIR so multi-listen does not hijack the port.
#
# Port 5000 is often already taken (e.g. other Keystone, AirPlay Receiver). If gostone exits with
# "address already in use", use:  TEMPEST_SMOKE_PORT=15007 ./scripts/tempest-smoke.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ADMIN_PW="${GOSTONE_BOOTSTRAP_ADMIN_PASSWORD:-tempest-smoke-pass}"
PORT="${TEMPEST_SMOKE_PORT:-5000}"

# No trailing $ — Tempest appends [id-…] to test names in discovery.
REGEX="${TEMPEST_SMOKE_REGEX:-^tempest\.api\.identity\.v3\.test_tokens\.TokensV3Test\.(test_create_token|test_validate_token)}"

BIN="$ROOT/build/gostone"
DBFILE="$(mktemp "${TMPDIR:-/tmp}/gostone-tempest-smoke-XXXXXX.db")"
TEMPEST_WS="$(mktemp -d "${TMPDIR:-/tmp}/gostone-tempest-ws-XXXXXX")"
SMOKE_INI="$(mktemp "${TMPDIR:-/tmp}/gostone-tempest-smoke-XXXXXX.ini")"
# Empty conf.d prevents merging /etc/gostone/gostone.conf.d when -c is used.
EMPTY_CONF_D="$(mktemp -d "${TMPDIR:-/tmp}/gostone-tempest-empty-conf.d-XXXXXX")"
PID=""
cleanup() {
  if [[ -n "${PID}" ]]; then kill "${PID}" 2>/dev/null || true; wait "${PID}" 2>/dev/null || true; fi
  rm -f "$DBFILE" "$SMOKE_INI"
  rm -rf "$TEMPEST_WS" "$EMPTY_CONF_D"
}
trap cleanup EXIT

go build -o "$BIN" ./cmd/gostone

cat >"$SMOKE_INI" <<EOF
[service]
listen = 127.0.0.1:${PORT}
public_url = http://127.0.0.1:${PORT}
EOF

# Drop env that would override the above or pull in extra config/listeners.
unset GOSTONE_CONFIG GOSTONE_CONFIG_DIR GOSTONE_HTTP_ADDR GOSTONE_PUBLIC_URL \
  GOSTONE_LISTEN_PUBLIC GOSTONE_LISTEN_ADMIN GOSTONE_LISTEN_INTERNAL \
  GOSTONE_SQLITE_DSN 2>/dev/null || true
export GOSTONE_CONFIG_DIR="$EMPTY_CONF_D"

export GOSTONE_DATABASE_CONNECTION="file:${DBFILE}"
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD="$ADMIN_PW"
export GOSTONE_TOKEN_PROVIDER=jwt
export GOSTONE_TOKEN_SECRET="${GOSTONE_TOKEN_SECRET:-tempest-smoke-jwt-secret-32chars-min!!}"

"$BIN" -c "$SMOKE_INI" > "${TMPDIR:-/tmp}/gostone-tempest-smoke.log" 2>&1 &
PID=$!
sleep 1
if ! kill -0 "$PID" 2>/dev/null; then
  echo "gostone exited immediately (often: address already in use on ${PORT}). Log:" >&2
  tail -120 "${TMPDIR:-/tmp}/gostone-tempest-smoke.log" >&2 || true
  exit 1
fi

ok=0
BASE="http://127.0.0.1:${PORT}"
for _ in $(seq 1 90); do
  # Another daemon on the same port can return 200 /health; require gostone's JSON.
  if resp=$(curl -sfS "${BASE}/health" 2>/dev/null) && printf '%s' "$resp" | grep -q gostone; then
    ok=1
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "gostone died while waiting for /health. Log:" >&2
    tail -120 "${TMPDIR:-/tmp}/gostone-tempest-smoke.log" >&2 || true
    exit 1
  fi
  sleep 0.3
done
if [[ "$ok" -ne 1 ]]; then
  echo "gostone /health did not look like this service on ${BASE} (wrong process on port ${PORT}?)" >&2
  echo "Last /health body (if any):" >&2
  curl -sS "${BASE}/health" 2>&1 | head -c 400 >&2 || true
  echo "" >&2
  echo "gostone log:" >&2
  tail -80 "${TMPDIR:-/tmp}/gostone-tempest-smoke.log" >&2 || true
  exit 1
fi

export PATH="${HOME}/.local/bin:${PATH}"
if ! command -v tempest >/dev/null 2>&1; then
  echo "Installing Tempest (pip install -r ci/requirements-tempest.txt)..." >&2
  python3 -m pip install --user -q -r ci/requirements-tempest.txt
fi

CONF_TMP="${TMPDIR:-/tmp}/tempest-smoke-$$.conf"
# Avoid sed breaking on & or / in passwords; match listen port for uri_v3.
ROOT="$ROOT" PORT="$PORT" CONF_TMP="$CONF_TMP" ADMIN_PW="$ADMIN_PW" python3 <<'PY'
import os
from pathlib import Path

root = Path(os.environ["ROOT"])
port = os.environ["PORT"]
pw = os.environ["ADMIN_PW"]
out = Path(os.environ["CONF_TMP"])
text = (root / "ci" / "tempest.conf").read_text(encoding="utf-8")
text = text.replace("__ADMIN_PASSWORD__", pw)
text = text.replace("http://127.0.0.1:5000", f"http://127.0.0.1:{port}")
out.write_text(text, encoding="utf-8")
PY
tempest init "$TEMPEST_WS"
install -m 644 "$CONF_TMP" "$TEMPEST_WS/etc/tempest.conf"
rm -f "$CONF_TMP"

echo "Running Tempest smoke (regex: ${REGEX}) against ${BASE} ..." >&2
cd "$TEMPEST_WS"
tempest run --config-file etc/tempest.conf --regex "${REGEX}"
