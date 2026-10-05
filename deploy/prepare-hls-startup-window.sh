#!/usr/bin/env bash

# Target server only: verify/build the bounded conventional-HLS startup change.
# This prepares candidate tags and evidence; it never restarts the live service.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 4 ]] || fail 'Usage: bash prepare-hls-startup-window.sh REPOSITORY BASE_OUTPUT OUTPUT_DIR CANDIDATE'
umask 077
for required in docker git realpath stat grep sed tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
[[ "$PWD" == "$repository" ]] || fail 'REPOSITORY must be the Git root'
base_commit=68dbdd4a8dd798886302b235c1f8f208452e0c6e
candidate="$4"
[[ "$candidate" =~ ^[0-9a-f]{40}$ ]] || fail 'full candidate commit required'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
checkout_before="$(git rev-parse HEAD)"
git merge-base --is-ancestor "$base_commit" "$checkout_before" || fail 'checkout precedes the expected live baseline'
git merge-base --is-ancestor "$checkout_before" "$candidate" || fail 'checkout is outside the candidate history'
[[ "$(git hash-object -- "$helper_source")" == "$(git rev-parse "$candidate:deploy/prepare-hls-startup-window.sh")" ]] || fail 'helper differs from the selected candidate'
git diff --quiet "$base_commit" "$candidate" -- client runtime apps build .dockerignore 'docker-compose*.yaml' deploy/collect-hls-media.sh deploy/deploy-hls-media.sh || fail 'client/runtime/build/deployment sources differ from the validated baseline'
[[ "$(git diff --name-only "$base_commit" "$candidate" -- server)" == server/internal/mediahls/packager.go ]] || fail 'candidate must change only the HLS startup window in server sources'
restored_blob="$(git show "$candidate:server/internal/mediahls/packager.go" |
  sed 's/ConventionalReadyWindow = 28 \* time.Second/ConventionalReadyWindow = 24 * time.Second/' |
  git hash-object --stdin)"
[[ "$restored_blob" == "$(git rev-parse "$base_commit:server/internal/mediahls/packager.go")" ]] || fail 'server delta is outside the reviewed 24-to-28-second change'

base_output="$(realpath -- "$2")"
output="$(realpath -m -- "$3")"
[[ -d "$base_output" && ! -L "$2" && "$(stat -c %a -- "$base_output")" == 700 ]] || fail 'private baseline evidence required'
case "$base_output/" in "$repository/"*) fail 'baseline evidence must be outside Git';; esac
case "$repository/" in "$base_output/"*) fail 'baseline evidence must not contain Git';; esac
[[ "$(cat "$base_output/validation-commit.txt")" == "$base_commit" ]] || fail 'exact-68 image preparation record required'
case "$output/" in "$repository/"*|"$base_output/"*) fail 'new OUTPUT_DIR must be outside Git and prior evidence';; esac
for retained_dir in "$repository" "$base_output"; do
  case "$retained_dir/" in "$output/"*) fail 'OUTPUT_DIR must not contain Git or prior evidence';; esac
done
[[ ! -e "$3" && ! -L "$3" ]] || fail 'new OUTPUT_DIR required; preserve any previous attempt'
container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$container" =~ ^[0-9a-f]{12,64}$ ]] || fail 'exactly one running Neko container required'
live_image="$(docker inspect -f '{{.Image}}' "$container")"
image_record="$(grep -E '^image=(docker.io/)?my-neko/brave:hls-68dbdd4a8dd7 id=sha256:[0-9a-f]{64} created=' "$base_output/images.txt")"
[[ "$image_record" != *$'\n'* ]] || fail 'ambiguous baseline image record'
base_image="${image_record#* id=}"
base_image="${base_image%% *}"
[[ "$live_image" == "$base_image" && "$(docker image inspect --format '{{.Id}}' my-neko/brave:hls-68dbdd4a8dd7)" == "$base_image" ]] || fail 'live/tagged image differs from the prepared exact-68 record'

# Guarded fast-forward and separate preparation. No stop/up/recreate/reload.
git merge --ff-only "$candidate"
[[ "$(git rev-parse HEAD)" == "$candidate" ]] || fail 'candidate checkout mismatch'
bash deploy/collect-hls-media.sh init "$output"
printf 'PENDING\n' >"$output/validation-commit.txt"
printf '%s\n' "$base_commit" >"$output/inherited-client-commit.txt"
printf '%s\n' "$base_output" >"$output/inherited-client-evidence.txt"
git hash-object -- "$helper_source" >"$output/operator-preparer-blob.txt"
printf '%s\n' 'server-only conventional startup allowance: 24 -> 28 seconds; unchanged exact-68 client/runtime/config evidence inherited; fresh HLS package/native startup and image checks required; live cold-start and wider acceptance pending' >"$output/preparation-scope.txt"
exec > >(tee "$output/validation.log") 2>&1
export NEKO_VALIDATION_COMMIT="$candidate"
image_tag="hls-${candidate:0:12}"
printf 'application_commit=%s\ninherited_client_commit=%s\nimage_tag=%s\n' "$candidate" "$base_commit" "$image_tag"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
# Rebuild the source-containing parent first; the native image must not use an
# earlier server-validation tag. Both suites bypass cached test results.
docker compose -f docker-compose.validation.yaml build server-checks
docker compose -f docker-compose.validation.yaml run --rm -T \
  server-checks go test ./internal/mediahls -count=1 </dev/null
docker compose -f docker-compose.validation.yaml build hls-packager-checks
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-packager-checks </dev/null
CLIENT_DIST='' bash -o pipefail ./build "my-neko/base:$image_tag" -y </dev/null
CLIENT_DIST='' bash -o pipefail ./build "my-neko/brave:$image_tag" -y </dev/null
docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
  --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
[[ "$(git rev-parse HEAD)" == "$candidate" && -z "$(git status --porcelain=v1)" ]] || fail 'candidate checkout changed during preparation'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during preparation'
bash deploy/collect-hls-media.sh snapshot "$output" startup-window-prepared-live-68
printf '%s\n' "$candidate" >"$output/validation-commit.txt"
printf 'STARTUP WINDOW PREPARATION PASSED; running service unchanged; live cold-start acceptance pending. Evidence: %s\n' "$output"
