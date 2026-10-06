#!/usr/bin/env bash

# Target server only: check the bounded client start-order candidate, then
# prepare its images. Keep the live exact-8f HLS/WebRTC container running.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 4 ]] || fail 'Usage: bash prepare-hls-client-start.sh REPOSITORY BASE_OUTPUT OUTPUT_DIR CANDIDATE'
umask 077
for required in docker git realpath stat grep tar tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
[[ "$PWD" == "$repository" ]] || fail 'REPOSITORY must be the Git root'
base_commit=8f54970f025e3491a123540cc94870508b66f119
candidate="$4"
[[ "$candidate" =~ ^[0-9a-f]{40}$ ]] || fail 'full candidate commit required'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
[[ "$(git rev-parse HEAD)" == "$base_commit" ]] || fail 'expected exact running application checkout'
git merge-base --is-ancestor "$base_commit" "$candidate"
[[ "$(git hash-object -- "$helper_source")" == "$(git rev-parse "$candidate:deploy/prepare-hls-client-start.sh")" ]] || fail 'helper differs from the selected candidate'
changed_files="$(git diff --name-only "$base_commit" "$candidate")"
while IFS= read -r file; do
  case "$file" in
    client/src/neko/hls/controller.ts|client/tests/hls-controller.test.mjs|deploy/prepare-hls-client-start.sh|AGENTS.md|README.md|docs/*) ;;
    *) fail 'candidate changed files outside the bounded client/start-order block' ;;
  esac
done <<< "$changed_files"
git diff --quiet "$base_commit" "$candidate" -- server runtime apps build .dockerignore 'docker-compose*.yaml' deploy/collect-hls-media.sh deploy/deploy-hls-media.sh || fail 'backend/build/deployment sources differ from the validated baseline'

base_output="$(realpath -- "$2")"
output="$(realpath -m -- "$3")"
[[ -d "$base_output" && ! -L "$2" && "$(stat -c %a -- "$base_output")" == 700 ]] || fail 'private baseline evidence required'
case "$base_output/" in "$repository/"*) fail 'baseline evidence must be outside Git';; esac
case "$repository/" in "$base_output/"*) fail 'baseline evidence must not contain Git';; esac
[[ "$(cat "$base_output/validation-commit.txt")" == "$base_commit" ]] || fail 'exact-8f image preparation record required'
case "$output/" in "$repository/"*|"$base_output/"*) fail 'new OUTPUT_DIR must be outside Git and prior evidence';; esac
for retained_dir in "$repository" "$base_output"; do
  case "$retained_dir/" in "$output/"*) fail 'OUTPUT_DIR must not contain Git or prior evidence';; esac
done
[[ ! -e "$3" && ! -L "$3" ]] || fail 'new OUTPUT_DIR required; preserve any previous attempt'
container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$container" =~ ^[0-9a-f]{12,64}$ ]] || fail 'exactly one running Neko container required'
live_image="$(docker inspect -f '{{.Image}}' "$container")"
image_record="$(grep -E '^image=(docker.io/)?my-neko/brave:hls-8f54970f025e id=sha256:[0-9a-f]{64} created=' "$base_output/images.txt")"
[[ "$image_record" != *$'\n'* ]] || fail 'ambiguous baseline image record'
base_image="${image_record#* id=}"
base_image="${base_image%% *}"
[[ "$live_image" == "$base_image" && "$(docker image inspect --format '{{.Id}}' my-neko/brave:hls-8f54970f025e)" == "$base_image" ]] || fail 'live/tagged image differs from the prepared exact-8f record'

mkdir -m 700 -- "$output"
mkdir -m 700 -- "$output/source"
printf '%s\n' "$base_commit" >"$output/base-commit.txt"
printf '%s\n' "$candidate" >"$output/candidate-commit.txt"
printf 'PENDING\n' >"$output/client-check-commit.txt"
printf 'PENDING\n' >"$output/validation-commit.txt"
git hash-object -- "$helper_source" >"$output/operator-preparer-blob.txt"
git archive "$candidate" client docker-compose.validation.yaml server/internal/mediaws/testdata server/internal/mediahls/testdata |
  tar -x -C "$output/source"
git show "$base_commit:client/src/neko/hls/controller.ts" >"$output/baseline-controller.ts"
cat >"$output/run-client-check.sh" <<'SH'
#!/bin/sh
set -eu
cp -a /source/client /work/client
mkdir -p /work/server/internal/mediaws/testdata /work/server/internal/mediahls/testdata
cp /source/server/internal/mediaws/testdata/neko_media_v1_golden.json /work/server/internal/mediaws/testdata/
cp /source/server/internal/mediahls/testdata/*.m3u8 /work/server/internal/mediahls/testdata/
cd /work/client
cp /candidate/baseline-controller.ts src/neko/hls/controller.ts
npm ci
# Only the missing early-Play assertion and exactly one failed test count as
# old-order reproduction. This controlled case does not prove a browser cause.
old_status=0
node --test --test-reporter=tap \
  --test-name-pattern='^initial HLS autoplay is requested even without a canplay event$' \
  tests/hls-controller.test.mjs >/work/old-start.tap 2>&1 || old_status=$?
if [ "$old_status" -ne 1 ] ||
   ! grep -Fq "error: 'initial HLS autoplay waited for canplay'" /work/old-start.tap ||
   ! grep -Fqx '# fail 1' /work/old-start.tap; then
  cat /work/old-start.tap
  printf 'error: expected old start-order assertion was not reproduced\n' >&2
  exit 1
fi
printf 'PASS old client: autoplay waits for canplay in the controlled first-frame case\n'
cp /source/client/src/neko/hls/controller.ts src/neko/hls/controller.ts
npm test
npm run lint
npm run build
SH
exec > >(tee "$output/validation.log") 2>&1
printf 'base_commit=%s\ncandidate_commit=%s\n' "$base_commit" "$candidate"
docker compose -p "neko-hls-client-start-${candidate:0:12}" \
  -f "$output/source/docker-compose.validation.yaml" run --rm -T \
  -v "$output:/candidate:ro" client-checks sh /candidate/run-client-check.sh </dev/null
[[ "$(git rev-parse HEAD)" == "$base_commit" && -z "$(git status --porcelain=v1)" ]] || fail 'checkout changed during isolated client checks'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during isolated client checks'
printf '%s\n' "$candidate" >"$output/client-check-commit.txt"

# The checked candidate is now assembled from the same Git blobs. Image builds
# create candidate tags only; there is no stop/up/recreate/reload operation.
git merge --ff-only "$candidate"
[[ "$(git rev-parse HEAD)" == "$candidate" ]] || fail 'candidate checkout mismatch'
bash deploy/collect-hls-media.sh init "$output"
printf '%s\n' "$base_commit" >"$output/inherited-backend-commit.txt"
printf '%s\n' "$base_output" >"$output/inherited-backend-evidence.txt"
printf '%s\n' 'client-only start-order candidate: fresh isolated old/new, full client tests/type/build and candidate image assembly; exact-8f backend/codec/fuzz evidence inherited where supplied, no new backend test or live first-start acceptance claim' >"$output/preparation-scope.txt"
image_tag="hls-${candidate:0:12}"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
CLIENT_DIST='' PUSH='' bash -o pipefail ./build "my-neko/base:$image_tag" -b "my-neko/base:$image_tag" -y </dev/null
CLIENT_DIST='' PUSH='' bash -o pipefail ./build "my-neko/brave:$image_tag" -b "my-neko/base:$image_tag" -y </dev/null
docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
  --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
[[ "$(git rev-parse HEAD)" == "$candidate" && -z "$(git status --porcelain=v1)" ]] || fail 'candidate checkout changed during image preparation'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during image preparation'
bash deploy/collect-hls-media.sh snapshot "$output" client-start-prepared-live-8f
printf '%s\n' "$candidate" >"$output/validation-commit.txt"
printf 'CLIENT START PREPARATION PASSED; running service unchanged; live first-start acceptance pending. Evidence: %s\n' "$output"
