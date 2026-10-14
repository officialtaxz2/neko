#!/usr/bin/env bash

# Target server only: reproduce the old rolling-playlist/watchdog defects in
# isolated copies, check the repair and prepare images without activating them.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 4 ]] || fail 'Usage: bash prepare-hls-playlist-window.sh REPOSITORY BASE_OUTPUT OUTPUT_DIR CANDIDATE'
umask 077
for required in docker git realpath stat grep tar tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
[[ "$PWD" == "$repository" ]] || fail 'REPOSITORY must be the Git root'
base_commit=7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b
candidate="$4"
[[ "$candidate" =~ ^[0-9a-f]{40}$ ]] || fail 'full candidate commit required'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
checkout_before="$(git rev-parse HEAD)"
git merge-base --is-ancestor "$base_commit" "$checkout_before" || fail 'checkout precedes the live baseline'
git merge-base --is-ancestor "$checkout_before" "$candidate" || fail 'checkout is outside the candidate history'
[[ "$(git hash-object -- "$helper_source")" == "$(git rev-parse "$candidate:deploy/prepare-hls-playlist-window.sh")" ]] || fail 'helper differs from the selected candidate'
changed_files="$(git diff --name-only "$base_commit" "$candidate")"
while IFS= read -r file; do
  case "$file" in
    server/internal/mediahls/packager.go|server/internal/mediahls/packager_playlist_window_test.go|client/src/neko/hls/controller.ts|client/tests/hls-controller.test.mjs|deploy/prepare-hls-playlist-window.sh|deploy/diagnose-hls-playback.sh|deploy/summarize-hls-startup.py|deploy/inspect-hls-browser-timing.js|deploy/trace-hls-browser.js|AGENTS.md|README.md|docs/*) ;;
    *) fail 'candidate changed files outside the reviewed playlist/watchdog and prior diagnostic block' ;;
  esac
done <<< "$changed_files"
git diff --quiet "$base_commit" "$candidate" -- runtime apps build .dockerignore 'docker-compose*.yaml' deploy/collect-hls-media.sh deploy/deploy-hls-media.sh || fail 'runtime/build/deployment sources differ from the baseline'

base_output="$(realpath -- "$2")"
output="$(realpath -m -- "$3")"
[[ -d "$base_output" && ! -L "$2" && "$(stat -c %a -- "$base_output")" == 700 ]] || fail 'private baseline evidence required'
case "$base_output/" in "$repository/"*) fail 'baseline evidence must be outside Git';; esac
case "$repository/" in "$base_output/"*) fail 'baseline evidence must not contain Git';; esac
[[ "$(cat "$base_output/validation-commit.txt")" == "$base_commit" ]] || fail 'exact-7dcc image preparation record required'
case "$output/" in "$repository/"*|"$base_output/"*) fail 'new OUTPUT_DIR must be outside Git and prior evidence';; esac
for retained_dir in "$repository" "$base_output"; do
  case "$retained_dir/" in "$output/"*) fail 'OUTPUT_DIR must not contain Git or prior evidence';; esac
done
[[ ! -e "$3" && ! -L "$3" ]] || fail 'new OUTPUT_DIR required; preserve any previous attempt'
container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$container" =~ ^[0-9a-f]{12,64}$ ]] || fail 'exactly one running Neko container required'
live_image="$(docker inspect -f '{{.Image}}' "$container")"
image_record="$(grep -E '^image=(docker.io/)?my-neko/brave:hls-7dcc3c5e4ba0 id=sha256:[0-9a-f]{64} created=' "$base_output/images.txt")"
[[ "$image_record" != *$'\n'* ]] || fail 'ambiguous baseline image record'
base_image="${image_record#* id=}"
base_image="${base_image%% *}"
[[ "$live_image" == "$base_image" && "$(docker image inspect --format '{{.Id}}' my-neko/brave:hls-7dcc3c5e4ba0)" == "$base_image" ]] || fail 'live/tagged image differs from the prepared exact-7dcc record'

mkdir -m 700 -- "$output"
mkdir -m 700 -- "$output/source"
printf '%s\n' "$base_commit" >"$output/base-commit.txt"
printf '%s\n' "$candidate" >"$output/candidate-commit.txt"
printf 'PENDING\n' >"$output/client-check-commit.txt"
printf 'PENDING\n' >"$output/server-check-commit.txt"
printf 'PENDING\n' >"$output/validation-commit.txt"
git hash-object -- "$helper_source" >"$output/operator-preparer-blob.txt"
git archive "$candidate" client server docker-compose.validation.yaml | tar -x -C "$output/source"
git show "$base_commit:client/src/neko/hls/controller.ts" >"$output/baseline-controller.ts"
git show "$base_commit:server/internal/mediahls/packager.go" >"$output/baseline-packager.go"
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
old_status=0
node --test --test-reporter=tap \
  --test-name-pattern='^live-edge seeks without current data cannot hide an HLS playback stall$' \
  tests/hls-controller.test.mjs >/work/old-watchdog.tap 2>&1 || old_status=$?
if [ "$old_status" -ne 1 ] ||
   ! grep -Fq "error: 'live-edge seeks hid the frozen HLS playback'" /work/old-watchdog.tap ||
   ! grep -Fqx '# fail 1' /work/old-watchdog.tap; then
  cat /work/old-watchdog.tap
  printf 'error: expected old seek-only progress assertion was not reproduced\n' >&2
  exit 1
fi
printf 'PASS old client: seek-only jumps hide the controlled playback stall\n'
cp /source/client/src/neko/hls/controller.ts src/neko/hls/controller.ts
npm test
npm run lint
npm run build
SH
cat >"$output/run-server-check.sh" <<'SH'
#!/bin/sh
set -eu
mkdir -p /work
cp -a /source/server /work/server
cd /work/server
cp /candidate/baseline-packager.go internal/mediahls/packager.go
old_status=0
go test ./internal/mediahls -run '^TestPackagerRollingPlaylistPreservesDiscontinuityNumbers$' -count=1 \
  >/work/old-playlist.txt 2>&1 || old_status=$?
if [ "$old_status" -ne 1 ] ||
   ! grep -Fq 'playlist discontinuity changed for retained segment:' /work/old-playlist.txt ||
   ! grep -Fq -- '--- FAIL: TestPackagerRollingPlaylistPreservesDiscontinuityNumbers' /work/old-playlist.txt; then
  cat /work/old-playlist.txt
  printf 'error: expected old rolling-playlist assertion was not reproduced\n' >&2
  exit 1
fi
printf 'PASS old server: retained segment discontinuity numbers change at the fourth parent\n'
cp /source/server/internal/mediahls/packager.go internal/mediahls/packager.go
go test ./internal/mediahls -run '^TestPackagerRollingPlaylistPreservesDiscontinuityNumbers$' -count=1 -v
go test ./internal/mediahls -count=1
SH
exec > >(tee "$output/validation.log") 2>&1
printf 'base_commit=%s\ncandidate_commit=%s\n' "$base_commit" "$candidate"
project="neko-hls-playlist-${candidate:0:12}"
docker compose -p "$project" -f "$output/source/docker-compose.validation.yaml" run --rm -T \
  -v "$output:/candidate:ro" client-checks sh /candidate/run-client-check.sh </dev/null
printf '%s\n' "$candidate" >"$output/client-check-commit.txt"
printf 'PASS old client reproduction; repaired full client tests/type/build\n' >"$output/check-results.txt"
export NEKO_VALIDATION_COMMIT="$candidate"
docker compose -p "$project" -f "$output/source/docker-compose.validation.yaml" build server-checks
docker compose -p "$project" -f "$output/source/docker-compose.validation.yaml" run --rm -T \
  -v "$output/source:/source:ro" -v "$output:/candidate:ro" \
  server-checks sh /candidate/run-server-check.sh </dev/null
printf '%s\n' "$candidate" >"$output/server-check-commit.txt"
printf 'PASS old playlist reproduction; repaired all-track/both-mode rolling wire check; full HLS package\n' >>"$output/check-results.txt"
[[ "$(git rev-parse HEAD)" == "$checkout_before" && -z "$(git status --porcelain=v1)" ]] || fail 'checkout changed during isolated checks'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during isolated checks'

# Fast-forward only after both isolated old/new checks and client/HLS suites.
# Image assembly creates candidate tags; no stop/up/recreate/reload is issued.
git merge --ff-only "$candidate"
[[ "$(git rev-parse HEAD)" == "$candidate" ]] || fail 'candidate checkout mismatch'
bash deploy/collect-hls-media.sh init "$output"
printf '%s\n' "$base_commit" >"$output/inherited-runtime-commit.txt"
printf '%s\n' "$base_output" >"$output/inherited-runtime-evidence.txt"
printf '%s\n' 'fresh old/new client and playlist-window checks, full client tests/type/build, HLS package and server/base/Brave builds; codec/source/transcoder/security configuration unchanged from exact-7dcc/exact-8f; prior native/fuzz evidence inherited, not rerun; live playback acceptance pending' >"$output/preparation-scope.txt"
image_tag="hls-${candidate:0:12}"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
CLIENT_DIST='' PUSH='' bash -o pipefail ./build "my-neko/base:$image_tag" -b "my-neko/base:$image_tag" -y </dev/null
CLIENT_DIST='' PUSH='' bash -o pipefail ./build "my-neko/brave:$image_tag" -b "my-neko/base:$image_tag" -y </dev/null
docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
  --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
[[ "$(git rev-parse HEAD)" == "$candidate" && -z "$(git status --porcelain=v1)" ]] || fail 'candidate checkout changed during image preparation'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during image preparation'
bash deploy/collect-hls-media.sh snapshot "$output" playlist-window-prepared-live-7dcc
printf '%s\n' "$candidate" >"$output/validation-commit.txt"
printf 'PASS candidate server/base/Brave builds; live exact-7dcc container/image retained\n' >>"$output/check-results.txt"
cat "$output/check-results.txt"
printf 'PLAYLIST WINDOW PREPARATION PASSED; running service unchanged; live playback acceptance pending. Evidence: %s\n' "$output"
