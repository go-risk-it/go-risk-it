#!/usr/bin/env bash
# wait-for-api.sh — Block until the go-risk-it API is reachable through Kong.
#
# Polls GET /api/v1/games (through Kong) with the anon key.
# Accepts any HTTP response that isn't 502/503/504 or a connection failure.
# Times out after 30 attempts (1s apart) with a non-zero exit.
#
# Expected variables (set by the caller):
#   ANON_KEY       — Supabase anon key for Authorization header
#   KONG_HTTP_PORT — Kong port (default: 8000)

KONG_HTTP_PORT="${KONG_HTTP_PORT:-8000}"
API_URL="http://localhost:${KONG_HTTP_PORT}/api/v1/games"
MAX_ATTEMPTS=30

for attempt in $(seq 1 $MAX_ATTEMPTS); do
    status=$(curl -s -o /dev/null -w '%{http_code}' \
        -H "Authorization: Bearer ${ANON_KEY}" \
        -H "apikey: ${ANON_KEY}" \
        "$API_URL" 2>/dev/null) || status="000"

    if [[ "$status" != "000" && "$status" != "502" && "$status" != "503" && "$status" != "504" ]]; then
        echo "API ready (HTTP $status)"
        return 0 2>/dev/null || exit 0
    fi

    echo "Waiting for API... (attempt $attempt/$MAX_ATTEMPTS, HTTP $status)"
    sleep 1
done

echo "ERROR: API not reachable after $MAX_ATTEMPTS attempts" >&2
return 1 2>/dev/null || exit 1
