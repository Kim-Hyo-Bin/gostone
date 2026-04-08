#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "${root}/build/bin"

VERSION="${VERSION:-dev}"
COMMIT="$(git -C "${root}" rev-parse --short HEAD 2>/dev/null || echo unknown)"
(cd "${root}" && go build \
  -ldflags "-X github.com/Kim-Hyo-Bin/gostone/internal/buildinfo.Version=${VERSION} -X github.com/Kim-Hyo-Bin/gostone/internal/buildinfo.Commit=${COMMIT}" \
  -o build/bin/gostone ./cmd/gostone)
echo "Built ${root}/build/bin/gostone"
