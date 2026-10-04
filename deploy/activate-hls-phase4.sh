#!/usr/bin/env bash

# Target server only. This operator helper can be extracted with git show into
# the private evidence directory, leaving the tested application checkout intact.
set -Eeuo pipefail
[[ $# -eq 2 ]] || { printf 'Usage: bash activate-hls-phase4.sh REPOSITORY OUTPUT_DIR\n' >&2; exit 2; }
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
umask 077
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }
[[ $EUID -eq 0 ]] || fail 'run as root for the host Caddy service'
for command in docker git caddy systemctl journalctl curl dirname mktemp rm realpath stat cp cmp date sleep tee; do
  command -v "$command" >/dev/null || fail "required command not found: $command"
done
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
readonly commit="$(git rev-parse HEAD)"
readonly image_tag="hls-${commit:0:12}"
bash deploy/collect-hls-media.sh init "$2"
readonly output="$(realpath -m -- "$2")"
test "$(cat "$output/validation-commit.txt")" = "$commit"
exec > >(tee -a "$output/activation.log") 2>&1
printf 'application_commit=%s\noperator_helper_blob=%s\n' "$commit" "$(git hash-object -- "$helper")"
printf 'runtime_error_redaction=pending\n' >"$output/caddy-log-check.txt"
docker image inspect "my-neko/brave:$image_tag" python:3.12-alpine >/dev/null
export NEKO_MEDIA_HLS_MODES=hls
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
readonly active=/etc/caddy/Caddyfile
readonly backup="$output/Caddyfile.before-hls"
readonly candidate="$output/Caddyfile.hls"
readonly merger="$(dirname -- "$helper")/merge-hls-caddy.py"
[[ -f "$active" && ! -L "$active" ]] || fail 'active Caddyfile must be a regular file'
[[ -f "$merger" && ! -L "$merger" ]] || fail 'extract the matching merge-hls-caddy.py beside this helper'
printf 'operator_merger_blob=%s\n' "$(git hash-object -- "$merger")"
printf '{"merge":"pending"}\n' >"$output/caddy-merge-check.json"
cp -- "$active" "$output/Caddyfile.merge-source"
caddy adapt --config "$active" --adapter caddyfile \
  >"$output/Caddyfile.before-merge.json" 2>"$output/caddy-merge-adapt.log" || \
  fail 'active Caddy adaptation failed; details remain private'
merge_check() {
  docker run --rm --network none --read-only \
    --mount "type=bind,src=$output,dst=/evidence" \
    --mount "type=bind,src=$merger,dst=/merge.py,readonly" \
    python:3.12-alpine python /merge.py "$1" /evidence
}
merge_check write
# Preserve relative import resolution by adapting a temporary candidate in
# the active file's directory. Neither this file nor raw JSON is printed.
candidate_temp="$(mktemp "$(dirname -- "$active")/.neko-hls-XXXXXXXX")"
trap 'rm -f -- "$candidate_temp"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
cp -- "$candidate" "$candidate_temp"
caddy adapt --config "$candidate_temp" --adapter caddyfile \
  >"$output/Caddyfile.hls.json" 2>>"$output/caddy-merge-adapt.log" || \
  fail 'candidate adaptation failed; no service change, details remain private'
merge_check verify
caddy validate --config "$candidate_temp" --adapter caddyfile \
  >"$output/caddy-merge-validate.log" 2>&1 || \
  fail 'merged Caddy validation failed; no service change, details remain private'
rm -f -- "$candidate_temp"
trap - EXIT INT TERM
cmp --silent -- "$output/Caddyfile.merge-source" "$active" || \
  fail 'active Caddyfile changed during preparation; no service change'
caddy adapt --config "$active" --adapter caddyfile \
  >"$output/Caddyfile.before-merge.recheck.json" 2>>"$output/caddy-merge-adapt.log" || \
  fail 'active Caddy recheck failed; no service change, details remain private'
cmp --silent -- "$output/Caddyfile.before-merge.json" "$output/Caddyfile.before-merge.recheck.json" || \
  fail 'active adapted configuration changed during preparation; no service change'
printf 'PASS complete merged Caddyfile validation; backup/reload follows.\n'
if [[ -e "$backup" ]]; then
  cmp --silent -- "$backup" "$active" || fail 'existing Caddy backup differs; preserve it and review'
else
  cp -a -- "$active" "$backup"
fi
systemctl is-active --quiet caddy
readonly container="$(docker compose -f docker-compose.yaml ps -q neko)"
test -n "$container"
readonly old_image="$(docker inspect -f '{{.Image}}' "$container")"
readonly rollback_image="my-neko/brave:rollback-$image_tag"
if docker image inspect "$rollback_image" >/dev/null 2>&1; then
  test "$(docker image inspect -f '{{.Id}}' "$rollback_image")" = "$old_image"
else
  docker image tag "$old_image" "$rollback_image"
fi
printf '%s\n' "$rollback_image" >"$output/rollback-image.txt"

caddy_changed=0
neko_stopped=0
recover() {
  local status="$1" recovery_failed=0
  trap - EXIT INT TERM
  if [[ "$status" -ne 0 ]]; then
    if [[ "$neko_stopped" -eq 1 ]]; then
      bash deploy/deploy-hls-media.sh rollback "$output" || recovery_failed=1
    fi
    if [[ "$caddy_changed" -eq 1 ]]; then
      cp -- "$backup" "$active" && systemctl reload caddy || recovery_failed=1
      systemctl is-active --quiet caddy || recovery_failed=1
    fi
    if [[ "$recovery_failed" -eq 1 ]]; then
      printf 'Recovery incomplete: inspect the private evidence and host services.\n' >&2
    elif [[ "$caddy_changed" -eq 1 || "$neko_stopped" -eq 1 ]]; then
      printf 'Activation failed; previous Caddyfile/Neko deployment retained or restored.\n' >&2
    fi
  fi
  exit "$status"
}
trap 'recover $?' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
cmp --silent -- "$output/Caddyfile.merge-source" "$active" || \
  fail 'active Caddyfile changed before reload; no service change'
caddy_changed=1
cp -- "$candidate" "$active"
systemctl reload caddy
systemctl is-active --quiet caddy
printf 'Caddy proposal applied; testing runtime error redaction before valid HLS leases.\n'

# Use the required deployment stop to produce an actual proxy error, rather
# than stopping the service again later. These credentials/IDs are synthetic.
readonly since="$(date -u '+%Y-%m-%d %H:%M:%S UTC')"
readonly marker="NEKO_HLS_PROBE_$(date -u +%s%N)"
readonly lease=AAAAAAAAAAAAAAAAAAAAAA
neko_stopped=1
docker compose -f docker-compose.yaml stop neko
status="$(curl --disable --noproxy '*' --silent --show-error --max-time 15 \
  --output /dev/null --write-out '%{http_code}' \
  --header "Authorization: Bearer $marker" \
  --header "Cookie: __Secure-neko-hls=$marker" \
  --header "Referer: https://neko.taxzvps.de/api/media/hls/$lease/master.m3u8?probe=$marker" \
  "https://neko.taxzvps.de/api/media/hls/$lease/master.m3u8?probe=$marker")"
[[ "$status" == 502 ]] || fail 'synthetic stopped-backend probe did not return 502'
sleep 1
journalctl --unit caddy --since "$since" --no-pager --output cat --lines 200 \
  >"$output/caddy-error-probe.jsonl"
docker run --rm -i --network none --read-only \
  --mount "type=bind,src=$output/caddy-error-probe.jsonl,dst=/evidence.jsonl,readonly" \
  python:3.12-alpine python - "$marker" "$lease" <<'PY'
import json
import pathlib
import sys

raw = pathlib.Path('/evidence.jsonl').read_text()
if sys.argv[1] in raw or '/api/media/hls/' + sys.argv[2] + '/' in raw:
    sys.exit('FAIL Caddy error probe: raw synthetic URI/header data remains; journal stays private')
matches = []
for line in raw.splitlines():
    try:
        item = json.loads(line)
    except ValueError:
        continue
    if not isinstance(item, dict):
        continue
    request = item.get('request')
    if isinstance(request, dict) and item.get('level') == 'error' and request.get('uri') == '/api/media/hls/:redacted':
        if 'headers' in request or 'resp_headers' in item:
            sys.exit('FAIL Caddy error probe: request/response header fields remain')
        matches.append(item)
if not matches:
    sys.exit('FAIL Caddy error probe: no redacted runtime error observed')
print('PASS Caddy runtime error: HLS URI redacted, header fields absent, synthetic marker absent')
PY
printf 'runtime_error_redaction=passed\n' >"$output/caddy-log-check.txt"

bash deploy/deploy-hls-media.sh enable "$output"
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks enabled </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T \
  -e NEKO_PUBLIC_BASE_URL=http://127.0.0.1:8082 hls-http-checks insecure-denied </dev/null
printf 'ACTIVATION/INVALID-INPUT GATE PASSED at %s; valid playback and grouped acceptance remain pending.\n' "$commit"
