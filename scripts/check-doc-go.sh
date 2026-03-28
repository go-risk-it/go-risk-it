#!/usr/bin/env bash
set -euo pipefail

# check-doc-go.sh — Verify doc.go existence meets the baseline minimum.
# Mirrors TestArch_DocGoExists logic for use as a pre-commit hook.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

BASELINE_FILE="$ROOT_DIR/internal/arch_baseline.json"
MIN_COUNT=$(python3 -c "import json; print(json.load(open('$BASELINE_FILE'))['minDocGoCount'])")

# Wiring roots — excluded from doc.go requirements (matches arch_test.go wiringRoots)
WIRING_ROOTS="|logic|logic/game|logic/game/move|logic/game/move/service|logic/lobby|data|data/game|data/lobby|web|web/game|web/lobby|"

is_wiring_root() {
  [[ "$WIRING_ROOTS" == *"|$1|"* ]]
}

MODULE_PREFIX="github.com/go-risk-it/go-risk-it/internal/"
doc_count=0
missing=()

while IFS= read -r pkg_path; do
  suffix="${pkg_path#"$MODULE_PREFIX"}"

  # Skip generated packages
  if [[ "$suffix" == *"/sqlc"* ]] || [[ "$suffix" == *"/mocks"* ]]; then
    continue
  fi

  # Skip internal root (prefix stripping didn't match)
  if [[ "$suffix" == "$pkg_path" ]]; then
    continue
  fi

  # Skip empty suffix (internal root itself)
  if [[ -z "$suffix" ]]; then
    continue
  fi

  # Skip wiring roots
  if is_wiring_root "$suffix"; then
    continue
  fi

  pkg_dir="$ROOT_DIR/internal/$suffix"

  # Check for doc.go or package comment
  if [[ -f "$pkg_dir/doc.go" ]]; then
    doc_count=$((doc_count + 1))
    continue
  fi

  # Check first .go file for package comment
  found_comment=false
  for f in "$pkg_dir"/*.go; do
    [[ -f "$f" ]] || continue
    [[ "$f" == *_test.go ]] && continue
    if head -5 "$f" | grep -q "^// Package "; then
      doc_count=$((doc_count + 1))
      found_comment=true
    fi
    break
  done

  if ! $found_comment; then
    missing+=("$suffix")
  fi
done < <(go list ./internal/... 2>/dev/null)

if (( doc_count < MIN_COUNT )); then
  echo "FAIL: doc.go count $doc_count is below baseline minimum $MIN_COUNT"
  echo ""
  echo "Missing doc.go in ${#missing[@]} packages:"
  for m in "${missing[@]}"; do
    echo "  - internal/$m"
  done
  echo ""
  echo "Run 'make new-package' to scaffold doc.go for a new package,"
  echo "or manually create doc.go following docs/doc-go-spec.md."
  exit 1
fi

echo "OK: doc.go coverage $doc_count packages (minimum: $MIN_COUNT)"
