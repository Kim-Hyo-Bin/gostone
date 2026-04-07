#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"
go test ./... -count=1 -coverprofile=coverage.out -coverpkg=./...
go tool cover -func=coverage.out | tail -n 5
