#!/usr/bin/env bash
# Smoke-test upstream OpenStack Keystone (Identity v3) started via
# docker-compose.upstream-keystone.yml — same flow as gostone MariaDB integration.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

for cmd in curl python3; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "Missing required command: $cmd" >&2
		exit 1
	fi
done

BASE_URL="${KEYSTONE_UPSTREAM_URL:-http://127.0.0.1:15000}"
BASE_URL="${BASE_URL%/}"
export KEYSTONE_ADMIN_PASSWORD="${KEYSTONE_ADMIN_PASSWORD:-admin}"

# Short wait so this script works right after `docker compose up` (before healthcheck passes).
if ! curl -sf "${BASE_URL}/v3" >/dev/null 2>&1; then
	echo "Waiting for ${BASE_URL}/v3 ..."
	for _ in $(seq 1 60); do
		if curl -sf "${BASE_URL}/v3" >/dev/null 2>&1; then
			break
		fi
		sleep 2
	done
fi
if ! curl -sf "${BASE_URL}/v3" >/dev/null 2>&1; then
	echo "Keystone not reachable at ${BASE_URL}/v3 (start the stack or set KEYSTONE_UPSTREAM_URL)." >&2
	exit 1
fi

echo "POST $BASE_URL/v3/auth/tokens (password, admin @ Default, project admin)..."
# Project-scoped token: upstream Keystone policy denies identity:list_users on unscoped tokens.
AUTH_JSON=$(python3 -c 'import json,os; print(json.dumps({"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":os.environ["KEYSTONE_ADMIN_PASSWORD"],"domain":{"name":"Default"}}}},"scope":{"project":{"name":"admin","domain":{"name":"Default"}}}}}))')

RESP=$(curl -sS -i -X POST "$BASE_URL/v3/auth/tokens" \
	-H 'Content-Type: application/json' \
	-d "$AUTH_JSON")

CODE=$(printf '%s' "$RESP" | head -1 | awk '{print $2}')
TOKEN=$(printf '%s' "$RESP" | sed -n 's/^[Xx]-[Ss]ubject-[Tt]oken: //p' | tr -d '\r')

if [[ "$CODE" != "201" ]]; then
	echo "Expected HTTP 201 from auth, got $CODE" >&2
	printf '%s\n' "$RESP" >&2
	exit 1
fi
if [[ -z "$TOKEN" ]]; then
	echo "Missing X-Subject-Token header" >&2
	printf '%s\n' "$RESP" >&2
	exit 1
fi

echo "GET $BASE_URL/v3/users ..."
TMP_USERS=$(mktemp)
trap 'rm -f "$TMP_USERS"' EXIT
HTTP_USERS=$(curl -sS -o "$TMP_USERS" -w '%{http_code}' "$BASE_URL/v3/users" -H "X-Auth-Token: $TOKEN")
if [[ "$HTTP_USERS" != "200" ]]; then
	echo "Expected HTTP 200 from /v3/users, got $HTTP_USERS" >&2
	cat "$TMP_USERS" >&2
	exit 1
fi

if ! python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); assert any(u.get("name")=="admin" for u in d.get("users",[]))' "$TMP_USERS"; then
	echo "Expected user admin in response:" >&2
	cat "$TMP_USERS" >&2
	exit 1
fi

echo "OK: upstream Keystone password auth and list users succeeded."
