#!/usr/bin/env bash

# Real target server only; explicitly invoked baseline/enable/rollback.
set -Eeuo pipefail
[[ ( $# -eq 2 || $# -eq 3 ) && ( "$1" == baseline || "$1" == enable || "$1" == rollback ) ]] || {
  printf 'Usage: bash deploy/deploy-hls-media.sh baseline|enable|rollback OUTPUT_DIR [REPOSITORY]\n' >&2
  exit 2
}
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "${3:-$(dirname -- "$helper")/..}"
readonly action="$1"
readonly commit="$(git rev-parse HEAD)"
readonly image_tag="hls-${commit:0:12}"
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
bash deploy/collect-hls-media.sh init "$2"
readonly output="$(realpath -m -- "$2")"
umask 077
exec > >(tee "$output/$action.log") 2>&1
printf 'application_commit=%s\noperator_deployer_blob=%s\n' "$commit" "$(git hash-object -- "$helper")"

dc() {
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
    -f docker-compose.webcodecs-ws.yaml "$@"
}
hls_dc() { dc -f docker-compose.hls.yaml "$@"; }
deploy_dc() {
  if [[ "$action" == enable ]]; then hls_dc "$@"; else dc "$@"; fi
}

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
  # Compare against the exact preparation record before touching the service.
  image_record="$(grep -E "^image=(docker.io/)?my-neko/brave:$image_tag id=sha256:[0-9a-f]{64} created=" "$output/images.txt")"
  [[ "$image_record" != *$'\n'* ]]
  prepared_image_id="${image_record#* id=}"
  prepared_image_id="${prepared_image_id%% *}"
  test "$(docker image inspect --format '{{.Id}}' "my-neko/brave:$image_tag")" = "$prepared_image_id"
  NEKO_IMAGE="my-neko/brave:$image_tag" deploy_dc config --quiet
  if ! docker image inspect "$rollback_image" >/dev/null 2>&1; then
    container="$(dc ps -q neko)"
    test -n "$container"
    docker image tag "$(docker inspect -f '{{.Image}}' "$container")" "$rollback_image"
  fi
  printf '%s\n' "$rollback_image" >"$output/rollback-image.txt"
  dc stop neko
  if NEKO_IMAGE="my-neko/brave:$image_tag" deploy_dc up -d --force-recreate --wait --wait-timeout 180; then
    deploy_dc ps
  else
    dc stop neko
    NEKO_IMAGE="$rollback_image" dc up -d --force-recreate --wait --wait-timeout 180
    printf '%s start failed; restored the saved prior image without the HLS overlay.\n' "$action" >&2
    exit 1
  fi
fi
bash deploy/collect-hls-media.sh snapshot "$output" "$action-idle"
printf '%s completed. Check public HTTP security and browser/media behavior next.\n' "$action"
