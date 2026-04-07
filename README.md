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
- **Data layer**: [GORM](https://gorm.io/) with [glebarez/sqlite](https://github.com/glebarez/sqlite) for local/dev (pure Go, no CGO).

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
  - `GOSTONE_SQLITE_DSN` — database `connection`
  - `GOSTONE_HTTP_ADDR` — `listen` address
  - `GOSTONE_TOKEN_SECRET` — JWT HMAC secret (**required** in practice; must match `[token] secret` if both are used)

Copy `config/gostone.conf.example` and adjust at least `[token] secret`.

## First boot (development)

When the database has **no users** yet, you can seed a default domain, project, `admin` user, and role by setting:

```bash
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD='your-dev-password'
```

Then authenticate with Identity API password auth: user `admin`, domain `Default`, and that password.

## Quick run

```bash
export GOSTONE_TOKEN_SECRET='dev-secret-at-least-32-chars-recommended'
export GOSTONE_BOOTSTRAP_ADMIN_PASSWORD='admin'
export GOSTONE_HTTP_ADDR=':8080'
./build/bin/gostone -c config/gostone.conf.example
```

Smoke checks:

```bash
curl -sS http://127.0.0.1:8080/health
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8080/   # expect 300
curl -sS http://127.0.0.1:8080/v3 | head -c 200; echo

TOKEN=$(curl -sS -X POST http://127.0.0.1:8080/v3/auth/tokens \
  -H 'Content-Type: application/json' \
  -d '{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"admin","domain":{"name":"Default"}}}}}}' \
  -D- -o /dev/null | sed -n 's/^[Xx]-[Ss]ubject-[Tt]oken: //p' | tr -d '\r')

curl -sS http://127.0.0.1:8080/v3/users -H "X-Auth-Token: $TOKEN" | head -c 300; echo
```

## Implemented vs. stub

Many v3 routes return **501 Not Implemented**; working pieces include version discovery, `/health`, `POST/GET/HEAD/DELETE /v3/auth/tokens` (JWT), and a subset of user listing/detail with policy checks.

## Tests

```bash
go test ./...
go vet ./...
```
