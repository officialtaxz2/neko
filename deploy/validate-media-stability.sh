#!/usr/bin/env bash

# Target server only. Prepare one exact reviewed commit; never deploy/restart it.
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
[[ $# -eq 1 || ( $# -eq 2 && ( "${2:-}" == --capture-ordering-repair || "${2:-}" == --hls-member-access ) ) ]] || {
  printf 'Usage: bash deploy/validate-media-stability.sh NEW_OUTPUT_DIR [--capture-ordering-repair|--hls-member-access]\n' >&2; exit 2;
}
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || {
  printf 'A clean, reviewed testing commit is required.\n' >&2; exit 1;
}
[[ ! -e "$1" && ! -L "$1" ]] || { printf 'Preserve existing evidence; use a new output directory.\n' >&2; exit 1; }
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
readonly selected_commit="$NEKO_VALIDATION_COMMIT"
readonly image_tag="hls-${NEKO_VALIDATION_COMMIT:0:12}"
readonly repository="$(pwd -P)"
readonly validation_mode="${2:-}"
readonly ordering_base=f03bc4bcf68be81a76e65f9195580daad76df7b2
readonly member_access_base=8d8126c92903dab58056fa316b626b38dcffb376
if [[ "$validation_mode" == --capture-ordering-repair ]]; then
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
if [[ "$validation_mode" == --hls-member-access ]]; then
  # Inherit only the supplied exact-8d backend/race/ordering evidence. Fresh
  # client tests/type/build and Base/Brave images remain required below.
  git merge-base --is-ancestor "$member_access_base" "$selected_commit"
  member_access_paths="$(git diff --no-renames --name-only "$member_access_base" "$selected_commit")"
  while IFS= read -r changed_path; do
    case "$changed_path" in
      '') ;;
      client/src/neko/index.ts|client/src/neko/hls/controller.ts|client/src/components/settings.vue|client/tests/hls-controller.test.mjs|deploy/validate-media-stability.sh|AGENTS.md|README.md|docs/*) ;;
      *) printf 'Member-access gate cannot inherit backend evidence after changing %s.\n' "$changed_path" >&2; exit 1 ;;
    esac
  done <<<"$member_access_paths"
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
if [[ "$validation_mode" == --capture-ordering-repair ]]; then
  git show "$ordering_base:server/internal/capture/media.go" >"$output/ordering-baseline.go"
  printf 'client_evidence=inherited exact-f03; 80 tests/type/build; unchanged client source\n'
fi
if [[ "$validation_mode" == --hls-member-access ]]; then
  printf 'backend_evidence=inherited exact-8d; unchanged backend/runtime/codec/config sources; no fresh backend/race/native/fuzz result\n'
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
  if [[ "$validation_mode" != --capture-ordering-repair ]]; then
    docker compose -f docker-compose.validation.yaml run --rm -T client-checks </dev/null
  fi
  if [[ "$validation_mode" != --hls-member-access ]]; then
    docker compose -f docker-compose.validation.yaml build server-checks </dev/null
  fi
  if [[ "$validation_mode" == --capture-ordering-repair ]]; then
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
  if [[ "$validation_mode" != --hls-member-access ]]; then
    docker compose -f docker-compose.validation.yaml run --rm -T server-checks sh -ec '
  go test ./pkg/utils ./pkg/drop ./pkg/types ./pkg/auth ./internal/config ./internal/capture ./internal/media ./internal/mediahls ./internal/mediaws ./internal/member/multiuser ./internal/session ./internal/http ./internal/http/legacy ./internal/websocket ./internal/websocket/handler ./internal/webrtc
  go test -race ./pkg/utils ./internal/config ./internal/capture ./internal/http/legacy ./internal/websocket ./internal/websocket/handler ./internal/webrtc
  ./build
' </dev/null
  fi
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
if [[ "$validation_mode" == --capture-ordering-repair ]]; then
  printf 'ORDERING GATE PASSED: old defect reproduced; repaired subscriptions passed 100 race repetitions; backend/race/server/Base/Brave checks passed; client evidence inherited from exact-f03.\n'
fi
if [[ "$validation_mode" == --hls-member-access ]]; then
  printf 'MEMBER ACCESS PREPARATION PASSED: fresh client tests/type/build and Base/Brave images; backend evidence inherited from unchanged exact-8d; live role/device gates pending.\n'
fi
printf 'PREPARATION PASSED; live service retained; device/comparison acceptance pending. Evidence: %s\n' "$output"
