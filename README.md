# gostone

A Go-based reimplementation of OpenStack **Keystone**, the identity service for OpenStack.

## Upstream Keystone

- **Source code**: [openstack/keystone on GitHub](https://github.com/openstack/keystone)
- **Keystone documentation**: [docs.openstack.org — Keystone](https://docs.openstack.org/keystone/latest/)
- **Identity API reference**: [docs.openstack.org — Identity API](https://docs.openstack.org/api-ref/identity/)

The goal is to port Keystone’s **API surface** so that it can run **standalone**—independent of the rest of the OpenStack stack—while matching the behavior needed for those APIs to operate correctly on their own.

## Conventions

- **Go style**: [Effective Go](https://go.dev/doc/effective_go) as the baseline; layout is also shaped like Keystone’s service boundaries (HTTP wiring vs. API handlers vs. auth).
- **HTTP**: [Gin](https://github.com/gin-gonic/gin) on top of `net/http`.
- **Data layer**: [GORM](https://gorm.io/) against **SQLite** ([glebarez/sqlite](https://github.com/glebarez/sqlite), no CGO), **MySQL/MariaDB**, or **PostgreSQL**, using the same **`[database] connection`** idea as Keystone (SQLAlchemy-style URLs or native DSNs).

## Layout

| Path | Role |
|------|------|
| `cmd/gostone` | Binary entrypoint |
| `internal/app` | Composes DB, JWT, hub, and Gin engine |
| `internal/server` | Registers routes: health, version discovery (`/`, `/v3`), Identity v3 mount |
| `internal/api/v3` | Identity API v3 handlers |
| `internal/api/discovery` | Version discovery JSON (Keystone-compatible) |
| `internal/auth` | `X-Auth-Token` middleware; `internal/auth/password` — password grant helpers |
| `internal/common/httperr` | Keystone-style HTTP error bodies |
| `internal/conf` | INI config (paths, flags, env overrides) |
| `internal/bootstrap` | Optional first-boot seeding from environment |
| `config/gostone.conf.example` | Example configuration |

## Build

```bash
./scripts/build.sh
# or
go build -o build/bin/gostone ./cmd/gostone
```

## Configuration

INI files follow a Keystone/oslo-style split: main file plus optional `*.conf` drop-ins.

- **Flags**: `--config-file` / `-c`, `--config-dir`
- **Paths**: `GOSTONE_CONFIG`, or `/etc/gostone/gostone.conf`, or `./gostone.conf`, plus `gostone.conf.d` when using default layout (see `internal/conf/paths.go`).
- **Environment overrides** (when set, applied after file config):
  - `GOSTONE_DATABASE_CONNECTION` — overrides `[database] connection` (Keystone-style URL or DSN)
  - `GOSTONE_SQLITE_DSN` — deprecated alias for the same field (SQLite-only name)
  - `GOSTONE_HTTP_ADDR` — `listen` address
  - `GOSTONE_PUBLIC_URL` — advertised public base URL for catalog bootstrap (e.g. `http://controller:5000`)
  - `GOSTONE_TOKEN_PROVIDER` — `uuid` (default) or `jwt`
  - `GOSTONE_TOKEN_SECRET` — required when `provider=jwt`; optional for `uuid`

Copy `config/gostone.conf.example` and set **`[token] secret`** if you use `provider=jwt`. For **`provider=uuid`** (default, OpenStack-friendly), the secret is not used.

## First boot (development)

When the database has **no users** yet, you can seed a default domain, project, `admin` user, and role by setting:

```bash
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD='your-dev-password'
```

Then authenticate with Identity API password auth: user `admin`, domain `Default`, and that password.

## Quick run

```bash
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD='admin'
export GOSTONE_HTTP_ADDR=':5000'
export GOSTONE_PUBLIC_URL='http://127.0.0.1:5000'
# Default [token] provider is uuid — no secret needed. For jwt, set GOSTONE_TOKEN_SECRET and provider=jwt.
./build/bin/gostone -c config/gostone.conf.example
```

Smoke checks:

```bash
curl -sS http://127.0.0.1:5000/health
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:5000/   # expect 300
curl -sS http://127.0.0.1:5000/v3 | head -c 200; echo

TOKEN=$(curl -sS -X POST http://127.0.0.1:5000/v3/auth/tokens \
  -H 'Content-Type: application/json' \
  -d '{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"admin","domain":{"name":"Default"}}}}}}' \
  -D- -o /dev/null | sed -n 's/^[Xx]-[Ss]ubject-[Tt]oken: //p' | tr -d '\r')

curl -sS http://127.0.0.1:5000/v3/users -H "X-Auth-Token: $TOKEN" | head -c 300; echo
```

## Implemented vs. stub

**Working toward full Keystone replacement:** default **UUID tokens** (persisted, revocable via `DELETE /v3/auth/tokens`), **service catalog** from DB (identity public endpoint seeded on bootstrap), **domains**, **projects**, **roles**, **role assignments** (list), **users** (list/get), and auth token flows. Many extensions (federation, trust, application credentials, LDAP, etc.) still return **501**.

**Not yet equivalent to production Keystone:** **Fernet** tokens (common default in OpenStack), full assignment/grant APIs, catalog editing APIs, revocation lists, and most OS-* extensions. Use A/B tests against real Keystone to close gaps.

## Tests

```bash
go test ./...
go vet ./...
```

Merged coverage across module packages (single profile; target **≥ 60%** for CI):

```bash
go test ./... -coverprofile=coverage.out -coverpkg=./...
go tool cover -func=coverage.out
```

Or run `./scripts/coverage.sh` for the same check and a total line at the end.
