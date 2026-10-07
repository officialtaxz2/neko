#!/usr/bin/env bash

# Target server only. Prepare one exact reviewed commit; never deploy/restart it.
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
[[ $# -eq 1 ]] || { printf 'Usage: bash deploy/validate-media-stability.sh NEW_OUTPUT_DIR\n' >&2; exit 2; }
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || {
  printf 'A clean, reviewed testing commit is required.\n' >&2; exit 1;
}
[[ ! -e "$1" && ! -L "$1" ]] || { printf 'Preserve existing evidence; use a new output directory.\n' >&2; exit 1; }
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
readonly selected_commit="$NEKO_VALIDATION_COMMIT"
readonly image_tag="hls-${NEKO_VALIDATION_COMMIT:0:12}"
readonly repository="$(pwd -P)"
assert_source() {
  [[ "$(git -C "$repository" rev-parse HEAD)" == "$selected_commit" &&
     "$(git -C "$repository" branch --show-current)" == testing &&
     -z "$(git -C "$repository" status --porcelain=v1)" ]] || {
    printf 'Checkout changed during preparation; do not deploy these artifacts.\n' >&2; exit 1;
  }
}
bash deploy/collect-hls-media.sh init "$1"
readonly output="$(realpath -m -- "$1")"
umask 077
printf 'PENDING\n' >"$output/validation-commit.txt"
exec > >(tee "$output/validation.log") 2>&1
printf 'validation_commit=%s\nimage_tag=%s\n' "$NEKO_VALIDATION_COMMIT" "$image_tag"
readonly live_container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$live_container" =~ ^[0-9a-f]{12,64}$ ]] || { printf 'Exactly one retained live Neko container is required.\n' >&2; exit 1; }
readonly live_image="$(docker inspect --format '{{.Image}}' "$live_container")"

bash -n deploy/collect-hls-media.sh deploy/validate-media-stability.sh
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
assert_source
mkdir -m 700 -- "$output/source"
git archive "$selected_commit" | tar -x -C "$output/source"
# Build only this immutable selected source, never ignored dist/runtime files.
(
  cd -- "$output/source"
  unset COMPOSE_FILE COMPOSE_PROJECT_NAME COMPOSE_PROFILES
  docker compose -f docker-compose.validation.yaml run --rm -T client-checks </dev/null
  docker compose -f docker-compose.validation.yaml build server-checks </dev/null
  # Shared-plane lifecycle changed. Recheck authorization/delivery packages too,
  # but do not repeat unchanged codec/fuzz fixtures or claim their fresh acceptance.
  docker compose -f docker-compose.validation.yaml run --rm -T server-checks sh -ec '
  go test ./pkg/utils ./pkg/drop ./pkg/types ./pkg/auth ./internal/config ./internal/capture ./internal/media ./internal/mediahls ./internal/mediaws ./internal/member/multiuser ./internal/session ./internal/http ./internal/http/legacy ./internal/websocket ./internal/websocket/handler ./internal/webrtc
  go test -race ./pkg/utils ./internal/config ./internal/capture ./internal/http/legacy ./internal/websocket ./internal/websocket/handler ./internal/webrtc
  ./build
' </dev/null
  assert_source
  # Root build accepts publishing/prebuilt/flavor/platform options from the shell.
  # This gate always uses native architecture, source-built client and local tags.
  unset PUSH CLIENT_DIST BASE_IMAGE FLAVOR PLATFORM REPOSITORY TAG TAGS APPLICATION YES NO_CACHE USE_BUILDX RUNTIME_DOCKERFILE
  USE_BUILDX=0 bash -o pipefail ./build "my-neko/base:$image_tag" -b "my-neko/base:$image_tag" -y </dev/null
  assert_source
  USE_BUILDX=0 bash -o pipefail ./build "my-neko/brave:$image_tag" -b "my-neko/base:$image_tag" -y </dev/null
  docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
    --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
)
cd -- "$repository"
assert_source
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$live_container" &&
   "$(docker inspect --format '{{.Image}}' "$live_container")" == "$live_image" ]] || {
  printf 'Live container/image changed during preparation; marker remains PENDING.\n' >&2; exit 1;
}
bash deploy/collect-hls-media.sh snapshot "$output" preparation-live-retained
printf '%s\n' "$NEKO_VALIDATION_COMMIT" >"$output/validation-commit.txt"
printf 'PREPARATION PASSED; live service retained; device/comparison acceptance pending. Evidence: %s\n' "$output"
