#!/usr/bin/env bash

set -Eeuo pipefail

readonly WS_PATH="/api/media/ws"
readonly VALID_KEY="dGhlIHNhbXBsZSBub25jZQ=="
readonly INVALID_TICKET="00000000000000000000000000000000"

usage() {
  cat <<'EOF'
Check the credential-free media-WebSocket HTTP/pre-upgrade boundary.

Run from the repository root on the real target server:

  NEKO_PUBLIC_BASE_URL=https://neko.example \
  NEKO_PUBLIC_ORIGIN=https://neko.example \
    ./deploy/check-media-websocket-http.sh disabled

  NEKO_PUBLIC_BASE_URL=https://neko.example \
  NEKO_PUBLIC_ORIGIN=https://neko.example \
    ./deploy/check-media-websocket-http.sh enabled

  NEKO_PUBLIC_BASE_URL=http://127.0.0.1:8082 \
  NEKO_PUBLIC_ORIGIN=https://neko.example \
    ./deploy/check-media-websocket-http.sh insecure-denied

The command sends no login credential and uses only a fixed invalid ticket.
It prints status codes but never response bodies.
EOF
}

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

request_status() {
  local path="$1"
  local origin="$2"
  local protocol="$3"
  local args=(
    --http1.1
    --silent
    --show-error
    --output /dev/null
    --write-out '%{http_code}'
    --max-time 15
    --header 'Connection: Upgrade'
    --header 'Upgrade: websocket'
    --header 'Sec-WebSocket-Version: 13'
    --header "Sec-WebSocket-Key: $VALID_KEY"
  )

  if [[ -n "$origin" ]]; then
    args+=(--header "Origin: $origin")
  fi
  if [[ -n "$protocol" ]]; then
    args+=(--header "Sec-WebSocket-Protocol: $protocol")
  fi

  curl "${args[@]}" "${NEKO_PUBLIC_BASE_URL%/}$path"
}

expect_status() {
  local name="$1"
  local expected="$2"
  local path="$3"
  local origin="$4"
  local protocol="$5"
  local actual

  actual="$(request_status "$path" "$origin" "$protocol")"
  if [[ "$actual" != "$expected" ]]; then
    printf 'FAIL %-28s expected=%s actual=%s\n' "$name" "$expected" "$actual" >&2
    return 1
  fi
  printf 'PASS %-28s status=%s\n' "$name" "$actual"
}

main() {
  local mode="${1:-}"
  local failures=0

  [[ $# -eq 1 ]] || { usage >&2; exit 2; }
  command -v curl >/dev/null 2>&1 || fail 'required command not found: curl'
  [[ "${NEKO_PUBLIC_BASE_URL:-}" =~ ^https?://[^/]+$ ]] || \
    fail 'NEKO_PUBLIC_BASE_URL must be one http(s) origin without a path'
  [[ "${NEKO_PUBLIC_ORIGIN:-}" =~ ^https?://[^/]+$ ]] || \
    fail 'NEKO_PUBLIC_ORIGIN must be one http(s) origin without a path'

  case "$mode" in
    disabled)
      expect_status 'disabled route' 404 "$WS_PATH?probe=1" "$NEKO_PUBLIC_ORIGIN" 'neko.media.v1' || failures=$((failures + 1))
      ;;
    enabled)
      expect_status 'query rejected first' 400 "$WS_PATH?probe=1" "$NEKO_PUBLIC_ORIGIN" 'neko.media.v1' || failures=$((failures + 1))
      expect_status 'missing origin' 403 "$WS_PATH" '' 'neko.media.v1' || failures=$((failures + 1))
      expect_status 'wrong origin' 403 "$WS_PATH" 'https://invalid.example' 'neko.media.v1' || failures=$((failures + 1))
      expect_status 'wrong first protocol' 426 "$WS_PATH" "$NEKO_PUBLIC_ORIGIN" 'invalid.protocol' || failures=$((failures + 1))
      expect_status 'missing ticket protocol' 400 "$WS_PATH" "$NEKO_PUBLIC_ORIGIN" 'neko.media.v1' || failures=$((failures + 1))
      expect_status 'malformed ticket' 400 "$WS_PATH" "$NEKO_PUBLIC_ORIGIN" 'neko.media.v1, neko.media.ticket.short' || failures=$((failures + 1))
      expect_status 'unknown ticket' 401 "$WS_PATH" "$NEKO_PUBLIC_ORIGIN" "neko.media.v1, neko.media.ticket.$INVALID_TICKET" || failures=$((failures + 1))
      ;;
    insecure-denied)
      expect_status 'cleartext direct request' 403 "$WS_PATH" "$NEKO_PUBLIC_ORIGIN" 'neko.media.v1' || failures=$((failures + 1))
      ;;
    -h|--help|help)
      usage
      return
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac

  if ((failures > 0)); then
    fail "$failures media-WebSocket boundary check(s) failed"
  fi
}

main "$@"
