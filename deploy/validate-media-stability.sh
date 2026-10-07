#!/usr/bin/env bash

# Target server only. Prepare one exact reviewed commit; never deploy/restart it.
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
[[ $# -eq 1 || ( $# -eq 2 && "${2:-}" == --capture-ordering-repair ) ]] || {
  printf 'Usage: bash deploy/validate-media-stability.sh NEW_OUTPUT_DIR [--capture-ordering-repair]\n' >&2; exit 2;
}
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || {
  printf 'A clean, reviewed testing commit is required.\n' >&2; exit 1;
}
[[ ! -e "$1" && ! -L "$1" ]] || { printf 'Preserve existing evidence; use a new output directory.\n' >&2; exit 1; }
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
readonly selected_commit="$NEKO_VALIDATION_COMMIT"
readonly image_tag="hls-${NEKO_VALIDATION_COMMIT:0:12}"
readonly repository="$(pwd -P)"
readonly ordering_mode="${2:-}"
readonly ordering_base=f03bc4bcf68be81a76e65f9195580daad76df7b2
if [[ -n "$ordering_mode" ]]; then
  # The operator supplied 80 passing client tests/type/build at this exact base.
  # Reuse them only for the bounded provider/test/helper/docs repair scope.
  git merge-base --is-ancestor "$ordering_base" "$selected_commit"
  ordering_paths="$(git diff --no-renames --name-only "$ordering_base" "$selected_commit")"
  while IFS= read -r changed_path; do
    case "$changed_path" in
      '') ;;
      server/internal/capture/media.go|server/internal/capture/media_test.go|deploy/validate-media-stability.sh|AGENTS.md|README.md|docs/*) ;;
      *) printf 'Ordering-only gate cannot inherit client evidence after changing %s.\n' "$changed_path" >&2; exit 1 ;;
    esac
  done <<<"$ordering_paths"
fi
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
if [[ -n "$ordering_mode" ]]; then
  git show "$ordering_base:server/internal/capture/media.go" >"$output/ordering-baseline.go"
  printf 'client_evidence=inherited exact-f03; 80 tests/type/build; unchanged client source\n'
fi
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
  if [[ -z "$ordering_mode" ]]; then
    docker compose -f docker-compose.validation.yaml run --rm -T client-checks </dev/null
  fi
  docker compose -f docker-compose.validation.yaml build server-checks </dev/null
  if [[ -n "$ordering_mode" ]]; then
    docker compose -f docker-compose.validation.yaml run --rm -T \
      -v "$output/ordering-baseline.go:/baseline-media.go:ro" server-checks sh -ec '
      cp internal/capture/media.go /tmp/repaired-media.go
      cp /baseline-media.go internal/capture/media.go
      old_status=0
      go test ./internal/capture -run "^TestMediaSubscriptionGenerationTransitionAfterFormatHandoff$" -count=1 \
        >/tmp/old-ordering.txt 2>&1 || old_status=$?
      if [ "$old_status" -ne 1 ] ||
         ! grep -Fq -- "--- FAIL: TestMediaSubscriptionGenerationTransitionAfterFormatHandoff" /tmp/old-ordering.txt ||
         ! grep -Fq "want generation 2 source_restart discontinuity" /tmp/old-ordering.txt; then
        cat /tmp/old-ordering.txt
        printf "Expected old format-handoff ordering failure was not reproduced.\n" >&2
        exit 1
      fi
      printf "PASS old provider: selected format loses its generation discontinuity\n"
      cp /tmp/repaired-media.go internal/capture/media.go
      go test -race ./internal/capture -run "^TestMediaSubscription" -count=100
      printf "PASS repaired subscription ordering/lifecycle: 100 repetitions under race\n"
    ' </dev/null
  fi
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
if [[ -n "$ordering_mode" ]]; then
  printf 'ORDERING GATE PASSED: old defect reproduced; repaired subscriptions passed 100 race repetitions; backend/race/server/Base/Brave checks passed; client evidence inherited from exact-f03.\n'
fi
printf 'PREPARATION PASSED; live service retained; device/comparison acceptance pending. Evidence: %s\n' "$output"
