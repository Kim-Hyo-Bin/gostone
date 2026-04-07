# gostone

A Go-based reimplementation of OpenStack **Keystone**, the identity service for OpenStack.

## Upstream Keystone

- **Source code**: [openstack/keystone on GitHub](https://github.com/openstack/keystone)
- **Keystone documentation**: [docs.openstack.org — Keystone](https://docs.openstack.org/keystone/latest/)
- **Identity API reference**: [docs.openstack.org — Identity API](https://docs.openstack.org/api-ref/identity/)

The goal is to port Keystone’s **API surface** so that it can run **standalone**—independent of the rest of the OpenStack stack—while matching the behavior needed for those APIs to operate correctly on their own.

## Conventions

- **Go style**: Code follows [Effective Go](https://go.dev/doc/effective_go) as the baseline for idiomatic Go and project-wide conventions.
- **HTTP**: APIs are served with **[Gin](https://github.com/gin-gonic/gin)** on top of `net/http`.
- **Data layer**: Persistence uses **[GORM](https://gorm.io/)** so multiple backends can sit behind one abstraction. Local/dev builds use **[glebarez/sqlite](https://github.com/glebarez/sqlite)** (pure Go, no CGO); other drivers can be wired in as needed.
