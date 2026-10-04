#!/usr/bin/env bash

# Real target server only; explicitly invoked enable/rollback with prepared images.
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
[[ $# -eq 2 && ( "$1" == enable || "$1" == rollback ) ]] || {
  printf 'Usage: bash deploy/deploy-hls-media.sh enable|rollback OUTPUT_DIR\n' >&2
  exit 2
}
readonly action="$1"
readonly commit="$(git rev-parse HEAD)"
readonly image_tag="hls-${commit:0:12}"
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
bash deploy/collect-hls-media.sh init "$2"
readonly output="$(realpath -m -- "$2")"
umask 077
exec > >(tee "$output/$action.log") 2>&1

dc() {
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
    -f docker-compose.webcodecs-ws.yaml "$@"
}
hls_dc() { dc -f docker-compose.hls.yaml "$@"; }

if [[ "$action" == rollback ]]; then
  rollback_image="$(cat "$output/rollback-image.txt")"
  [[ "$rollback_image" =~ ^my-neko/brave:rollback-hls-[0-9a-f]{12}$ ]]
  docker image inspect "$rollback_image" >/dev/null
  dc stop neko
  NEKO_IMAGE="$rollback_image" dc up -d --force-recreate --wait --wait-timeout 180
  dc ps
else
  rollback_image="my-neko/brave:rollback-$image_tag"
  test "$(cat "$output/validation-commit.txt")" = "$commit"
  docker image inspect "my-neko/brave:$image_tag" >/dev/null
  hls_dc config --quiet
  if ! docker image inspect "$rollback_image" >/dev/null 2>&1; then
    container="$(dc ps -q neko)"
    test -n "$container"
    docker image tag "$(docker inspect -f '{{.Image}}' "$container")" "$rollback_image"
  fi
  printf '%s\n' "$rollback_image" >"$output/rollback-image.txt"
  dc stop neko
  if NEKO_IMAGE="my-neko/brave:$image_tag" hls_dc up -d --force-recreate --wait --wait-timeout 180; then
    hls_dc ps
  else
    dc stop neko
    NEKO_IMAGE="$rollback_image" dc up -d --force-recreate --wait --wait-timeout 180
    printf 'HLS start failed; restored the previous image without the HLS overlay.\n' >&2
    exit 1
  fi
fi
bash deploy/collect-hls-media.sh snapshot "$output" "$action-idle"
printf '%s completed. Check public HTTP security and browser/media behavior next.\n' "$action"
