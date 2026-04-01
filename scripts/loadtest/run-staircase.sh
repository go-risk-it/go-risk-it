#!/usr/bin/env bash
# run-staircase.sh — Run a staircase load test with all observability enabled.
#
# Usage:
#   ./run-staircase.sh                          # default staircase preset
#   ./run-staircase.sh --preset staircase-light # lighter preset
#   ./run-staircase.sh --steps 5,10,20          # custom steps (no preset)
#   ./run-staircase.sh --skip-rebuild           # reuse running stack
#   ./run-staircase.sh --branch feature-x       # build from a specific branch
#
# The stack is rebuilt from main by default. Use --skip-rebuild to reuse the
# current running containers, or --branch <name> to build from another branch.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/common.sh"

ensure_fresh_stack "$@"
resolve_anon_key
build_default_args

DEFAULT_ARGS+=(--save-journal)

# Strip stack-management flags before forwarding to the Go binary.
FORWARD_ARGS=($(strip_stack_flags "$@"))

# If no --preset and no --steps passed, default to staircase preset.
if ! has_flag "--preset" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}" && \
   ! has_flag "--steps" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"; then
    DEFAULT_ARGS+=(--preset staircase)
fi

# If no --journal-name passed, auto-generate from preset or "staircase".
if ! has_flag "--journal-name" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"; then
    PRESET_NAME=$(extract_flag_value "--preset" "staircase" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}")
    DEFAULT_ARGS+=(--journal-name "$PRESET_NAME")
fi

# --- Wait for API readiness through Kong ---
source "$SCRIPT_DIR/wait-for-api.sh"

print_banner "Staircase Load Test" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"

cd "$PROJECT_ROOT"
exec go run ./cmd/loadtest/ "${DEFAULT_ARGS[@]}" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"
