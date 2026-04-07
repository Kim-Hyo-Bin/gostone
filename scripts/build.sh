#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "${root}/build/bin"

(cd "${root}" && go build -o build/bin/gostone ./cmd/gostone)
echo "Built ${root}/build/bin/gostone"
