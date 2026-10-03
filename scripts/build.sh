#!/usr/bin/env bash
# scripts/build.sh — compile all pj binaries into ./bin
#
# Manager   -> bin/pj       (with version metadata baked in)
# Plugins   -> bin/pj-*     (plain build)
#
# Usage:
#     ./scripts/build.sh
#
# Requires: go 1.22+ on PATH.

set -euo pipefail

cd "$(dirname "$0")/.."

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Note: VERSION from `git describe` never contains spaces, so we can
# embed it directly without extra quoting.
MANAGER_LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${DATE}"

OUT="bin"
mkdir -p "$OUT"

build() {
	local pkg="$1" name="$2" ldflags="${3:-}"
	printf '  %-20s -> %s\n' "$pkg" "$OUT/$name"
	if [ -n "$ldflags" ]; then
		CGO_ENABLED=0 go build -trimpath -ldflags "$ldflags" -o "$OUT/$name" "./$pkg"
	else
		CGO_ENABLED=0 go build -trimpath -o "$OUT/$name" "./$pkg"
	fi
}

echo "pj build — ${VERSION}"

build cmd/pj pj "$MANAGER_LDFLAGS"

shopt -s nullglob
for d in cmd/pj-*; do
	[ -d "$d" ] || continue
	build "$d" "$(basename "$d")"
done
shopt -u nullglob

echo
echo "output:"
ls -lh "$OUT"