#!/usr/bin/env bash
set -euo pipefail

# new-package.sh — Scaffold a new Go package with a doc.go from the
# lightweight template. Usage: scripts/new-package.sh <package-path> <layer>
#
# Example:
#   scripts/new-package.sh internal/logic/game/forfeit Logic

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <package-path> <layer-name>"
  echo ""
  echo "  package-path  Path relative to repo root (e.g., internal/logic/game/forfeit)"
  echo "  layer-name    Architecture layer (e.g., Logic, Infrastructure, Web, Data, API)"
  echo ""
  echo "Available layers:"
  echo "  API, Infrastructure, Ctx, Data, Events, Events-domain,"
  echo "  Logic, Shared, Web, Test"
  exit 1
fi

PKG_PATH="$1"
LAYER="$2"
PKG_NAME="$(basename "$PKG_PATH")"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FULL_PATH="$ROOT_DIR/$PKG_PATH"

if [[ -d "$FULL_PATH" ]]; then
  echo "ERROR: Directory already exists: $PKG_PATH" >&2
  exit 1
fi

mkdir -p "$FULL_PATH"

cat > "$FULL_PATH/doc.go" << EOF
// Package ${PKG_NAME} TODO: add package summary.
//
// # Layer
//
// ${LAYER} — TODO: describe role.
package ${PKG_NAME}
EOF

echo "Created $PKG_PATH/doc.go"
echo ""
echo "Next steps:"
echo "  1. Edit $PKG_PATH/doc.go — replace TODO placeholders"
echo "  2. Add your .go files"
echo "  3. Run: go test ./internal/ -run TestArch -count=1"
