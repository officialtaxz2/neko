#!/usr/bin/env bash

# Read-only LL-HLS readiness check for the target server. Run it from the
# repository root on the Neko host. It changes nothing: no Compose, Caddy or
# container action. Output contains only fixed verdicts, counts and timings;
# no cookies, lease paths, client addresses or raw configuration.
set -Eeuo pipefail

usage() {
  cat <<'EOF'
Usage: bash deploy/check-ll-hls.sh https://PUBLIC-ORIGIN [WATCH_SECONDS]

  PUBLIC-ORIGIN   the public Neko origin, e.g. https://neko.example
  WATCH_SECONDS   optional (10..600). Play an LL-HLS stream in a browser
                  during this window; the check then reports how blocking
                  playlist reloads and part requests behaved end to end.

Environment overrides:
  NEKO_METRICS_URL  default http://127.0.0.1:8082/metrics
  CADDYFILE         default /etc/caddy/Caddyfile
EOF
}

[[ $# -ge 1 && $# -le 2 ]] || { usage >&2; exit 2; }
origin="${1%/}"
watch="${2:-0}"
[[ "$origin" =~ ^https://[A-Za-z0-9.-]+(:[0-9]+)?$ ]] || { echo "error: origin must be https://host[:port]" >&2; exit 2; }
[[ "$watch" =~ ^[0-9]+$ ]] && { [[ "$watch" -eq 0 ]] || [[ "$watch" -ge 10 && "$watch" -le 600 ]]; } || \
  { echo "error: WATCH_SECONDS must be 10..600" >&2; exit 2; }
metrics_url="${NEKO_METRICS_URL:-http://127.0.0.1:8082/metrics}"
caddyfile="${CADDYFILE:-/etc/caddy/Caddyfile}"

failures=0
warnings=0
pass() { printf 'PASS  %s\n' "$*"; }
warn() { printf 'WARN  %s\n' "$*"; warnings=$((warnings + 1)); }
fail() { printf 'FAIL  %s\n' "$*"; failures=$((failures + 1)); }
info() { printf 'INFO  %s\n' "$*"; }

for command in docker curl awk grep sed; do
  command -v "$command" >/dev/null || { echo "error: required command not found: $command" >&2; exit 2; }
done

printf '== 1. Neko configuration\n'
container="$(docker compose -f docker-compose.yaml ps -q neko 2>/dev/null || true)"
if [[ -z "$container" ]]; then
  fail 'neko container is not running'
else
  # Only the two HLS switches are read from the container environment.
  env_lines="$(docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$container" | grep -E '^NEKO_MEDIA_HLS_(ENABLED|MODES)=' || true)"
  enabled="$(sed -n 's/^NEKO_MEDIA_HLS_ENABLED=//p' <<<"$env_lines")"
  modes="$(sed -n 's/^NEKO_MEDIA_HLS_MODES=//p' <<<"$env_lines")"
  [[ "$enabled" == true ]] && pass 'HLS is enabled' || fail 'HLS is not enabled (docker-compose.hls.yaml missing from COMPOSE_FILE?)'
  if [[ " $modes " == *" ll-hls "* ]]; then
    pass "advertised HLS modes: $modes"
  else
    fail "LL-HLS is not advertised (NEKO_MEDIA_HLS_MODES='${modes}'); set 'hls ll-hls' and recreate neko"
  fi
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container")"
  [[ "$health" == healthy ]] && pass 'container is healthy' || warn "container health: $health"
fi

printf '\n== 2. Public HTTP protocol\n'
# LL-HLS keeps two blocking playlist reloads open per viewer. HTTP/2 or 3
# multiplexes them with part requests on one connection.
version="$(curl -sS -o /dev/null --http2 --max-time 10 -w '%{http_version}' "$origin/" 2>/dev/null || true)"
case "$version" in
  2|3) pass "public origin negotiates HTTP/$version" ;;
  1.1) fail 'public origin only offers HTTP/1.1; LL-HLS needs HTTP/2 or HTTP/3 at the proxy' ;;
  *) fail 'public origin is not reachable from this host' ;;
esac
if curl -sS -o /dev/null -D - --http2 --max-time 10 "$origin/" 2>/dev/null | grep -qi '^alt-svc:.*h3'; then
  info 'HTTP/3 is advertised as well (alt-svc)'
fi

printf '\n== 3. Caddy reverse proxy\n'
if command -v caddy >/dev/null && [[ -r "$caddyfile" ]]; then
  adapted="$(caddy adapt --config "$caddyfile" --adapter caddyfile 2>/dev/null || true)"
  if [[ -z "$adapted" ]]; then
    warn "could not adapt $caddyfile; check it with caddy validate"
  else
    # Blocking reloads hold a request for up to seven seconds before the full
    # playlist is written. Any proxy/server timeout must exceed that.
    short=0
    while IFS=: read -r key value; do
      [[ -z "$key" ]] && continue
      if (( value > 0 && value < 10000000000 )); then
        warn "Caddy ${key//\"/} is $((value / 1000000)) ms; LL-HLS blocking reloads need at least 10 s"
        short=1
      fi
    done < <(grep -oE '"(read_timeout|read_header_timeout|write_timeout|response_header_timeout|dial_timeout)":[0-9]+' <<<"$adapted" | \
      grep -vE '"(read_header_timeout|dial_timeout)"' || true)
    (( short == 0 )) && pass 'no Caddy read/write/response-header timeout below 10 s'
    if grep -qE '"(request_buffers|response_buffers)":[1-9]' <<<"$adapted"; then
      warn 'Caddy request/response buffering is configured; remove it for the Neko route'
    else
      pass 'no Caddy request/response buffering'
    fi
    if grep -q '"flush_interval"' <<<"$adapted"; then
      info 'flush_interval is set; not needed for LL-HLS (Neko writes complete playlists/parts)'
    fi
    if grep -q '"encode"' <<<"$adapted"; then
      info 'a Caddy encode handler exists; it must not compress the Neko route (MP4 parts stay uncompressed)'
    fi
  fi
else
  info "caddy or $caddyfile not available here; skipped proxy review"
fi

printf '\n== 4. Viewer round-trip time (HTTPS connections on :443)\n'
# LL-HLS with one-second parts needs path RTT <= 333 ms (part target >= 3 x RTT).
if command -v ss >/dev/null; then
  rtts="$(ss -Htin state established '( sport = :443 )' 2>/dev/null | grep -oE ' rtt:[0-9.]+' | sed 's/ rtt://' | sort -n || true)"
  count="$(grep -c . <<<"$rtts" || true)"
  if [[ "$count" -eq 0 ]]; then
    info 'no established HTTPS connections; run again while viewers are connected'
  else
    summary="$(awk '{v[NR]=$1} END {p=int(NR*0.95+0.999); if (p<1) p=1; printf "%d connections, median %.0f ms, p95 %.0f ms, max %.0f ms", NR, v[int((NR+1)/2)], v[p], v[NR]}' <<<"$rtts")"
    p95="$(awk '{v[NR]=$1} END {p=int(NR*0.95+0.999); if (p<1) p=1; printf "%.0f", v[p]}' <<<"$rtts")"
    if (( p95 <= 333 )); then pass "RTT $summary"; else warn "RTT $summary; LL-HLS may stall on the slowest paths, conventional HLS stays available"; fi
  fi
else
  info 'ss not available; skipped RTT sample'
fi

if [[ "$watch" -gt 0 ]]; then
  printf '\n== 5. LL-HLS requests over %s s (play LL-HLS in a browser now)\n' "$watch"
  snapshot() { curl -sS --max-time 10 "$metrics_url" | grep -E '^neko_media_hls_(requests_total|request_duration_seconds_(sum|count)|bootstrap_total)\{' | grep 'mode="ll-hls"' || true; }
  before="$(snapshot)"
  sleep "$watch"
  after="$(snapshot)"
  deltas="$(awk 'NR==FNR {old[$1]=$2; next} {d=$2-old[$1]; if (d>0) print $1, d}' <(printf '%s\n' "$before") <(printf '%s\n' "$after"))"
  value() { awk -v key="$1" 'index($1, key) {s+=$2} END {printf "%g", s+0}' <<<"$deltas"; }
  bootstrap="$(value 'bootstrap_total{mode="ll-hls",result="success"')"
  playlist_ok="$(value 'requests_total{mode="ll-hls",resource="playlist",result="success"')"
  playlist_timeout="$(value 'requests_total{mode="ll-hls",resource="playlist",result="timeout"')"
  playlist_all="$(value 'requests_total{mode="ll-hls",resource="playlist"')"
  part_ok="$(value 'requests_total{mode="ll-hls",resource="part",result="success"')"
  part_missing="$(value 'requests_total{mode="ll-hls",resource="part",result="not_found"')"
  duration_sum="$(value 'request_duration_seconds_sum{mode="ll-hls",resource="playlist"')"
  duration_count="$(value 'request_duration_seconds_count{mode="ll-hls",resource="playlist"')"
  info "bootstraps $bootstrap, playlist requests $playlist_all (success $playlist_ok, timeout $playlist_timeout), parts success $part_ok, parts not found $part_missing"
  if awk -v n="$playlist_ok" 'BEGIN {exit !(n == 0)}'; then
    fail 'no successful LL-HLS playlist request in the window; was an LL-HLS stream playing?'
  else
    mean="$(awk -v s="$duration_sum" -v c="$duration_count" 'BEGIN {if (c > 0) printf "%.2f", s / c; else print 0}')"
    # A held reload returns when the next one-second part is published, so a
    # healthy mean is a fraction of a second up to about one second.
    if awk -v m="$mean" 'BEGIN {exit !(m >= 0.2 && m <= 2)}'; then
      pass "blocking playlist reloads are held and released (mean ${mean} s)"
    else
      warn "mean LL-HLS playlist duration ${mean} s; expected about 0.2–2 s for held reloads"
    fi
    if awk -v t="$playlist_timeout" -v a="$playlist_all" 'BEGIN {exit !(a > 0 && t / a > 0.05)}'; then
      warn 'more than 5% of blocking reloads timed out (media publication stalled?)'
    fi
    if awk -v n="$part_ok" 'BEGIN {exit !(n > 0)}'; then pass 'parts are delivered'; else warn 'no part requests succeeded; the player may be loading whole segments only'; fi
    if awk -v n="$part_missing" 'BEGIN {exit !(n > 0)}'; then warn 'some listed parts returned 404 (they stop the player)'; fi
  fi
fi

printf '\nResult: %d failure(s), %d warning(s)\n' "$failures" "$warnings"
(( failures == 0 ))
