#!/usr/bin/env bash
# run-baseline.sh — Run a perf-test baseline with all observability enabled.
#
# Usage:
#   ./run-baseline.sh                    # baseline preset, auto-named
#   ./run-baseline.sh --preset light     # different preset
#   ./run-baseline.sh --games 2          # override specific flags
#   ./run-baseline.sh --skip-rebuild     # reuse running stack
#   ./run-baseline.sh --branch feature-x # build from a specific branch
#
# The stack is rebuilt from main by default. Use --skip-rebuild to reuse the
# current running containers, or --branch <name> to build from another branch.
set -euo pipefail

PERF_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$PERF_DIR/scripts/common.sh"

ensure_fresh_stack "$@"
resolve_anon_key
build_default_args

DEFAULT_ARGS+=(
    --output text
    --save-baseline
)

# Strip stack-management flags before forwarding to the Go binary.
FORWARD_ARGS=($(strip_stack_flags "$@"))

# If no --preset and no --games passed, default to baseline preset.
if ! has_flag "--preset" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}" && \
   ! has_flag "--games" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"; then
    DEFAULT_ARGS+=(--preset baseline)
fi

# If no --baseline-name passed, auto-generate from preset or "baseline".
if ! has_flag "--baseline-name" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"; then
    PRESET_NAME=$(extract_flag_value "--preset" "baseline" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}")
    DEFAULT_ARGS+=(--baseline-name "$PRESET_NAME")
fi

# --- Wait for API readiness through Kong ---
source "$PERF_DIR/wait-for-api.sh"

print_banner "Perf Test Baseline" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"

cd "$PERF_DIR"
exec go run ./cmd/loadtest/ "${DEFAULT_ARGS[@]}" "${FORWARD_ARGS[@]+"${FORWARD_ARGS[@]}"}"
