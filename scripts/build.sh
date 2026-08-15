#!/usr/bin/env bash
set -euo pipefail

# Build the domainsearch binary for the current platform into ./bin.
# Usage: ./scripts/build.sh

cd "$(dirname "$0")/.."

OUT="bin/domainsearch"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
	OUT="${OUT}.exe"
fi

echo "Building ${OUT}..."
go build -o "${OUT}" ./cmd/domainsearch

echo "Done. Run: ./${OUT}"
