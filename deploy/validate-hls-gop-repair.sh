#!/usr/bin/env bash

# Target-only synthetic codec comparison; never replace the running service.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash validate-hls-gop-repair.sh REPOSITORY OUTPUT_DIR REPAIR_COMMIT'
umask 077
for required in docker git realpath stat grep tr; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 80020d99477a58318f210b7e14d19cdd92991a6d ]] || fail 'keep the application at the restored 80020d99 baseline'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly repair_commit="$3"
[[ "$repair_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full repair commit required'
git cat-file -e "$repair_commit^{commit}"
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-gop-repair.sh")" ]] || fail 'helper differs from the selected repair commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ "$(cat "$output/validation-commit.txt")" == "$application_commit" ]] || fail 'baseline preparation marker differs'
readonly report="$output/gop-repair-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
readonly image="$(docker image inspect --format '{{.Id}}' my-neko/hls-validation:testing)"
[[ "$image" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'existing codec-validation image required'
printf 'application_commit=%s\nrepair_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\n' \
  "$application_commit" "$repair_commit" "$image" "$(git hash-object -- "$helper")" \
  | tee "$report/provenance.txt"

# Verify that the comparison really uses the prior implementation, not a
# subsequently rebuilt mutable tag. All test runs use the pinned image ID.
for source in internal/mediahls/transcoder.go internal/mediahls/packager.go pkg/gst/gst.go pkg/gst/gst.c; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image sources differ from the 80020d99 baseline'
done
git show "$repair_commit:server/internal/mediahls/codec_integration_test.go" >"$report/codec_integration_test.go"
git show "$repair_commit:server/internal/mediahls/transcoder.go" >"$report/transcoder.go"

negative_exit=0
docker run --rm --network none --workdir /src \
  --mount "type=bind,src=$report/codec_integration_test.go,dst=/src/internal/mediahls/codec_integration_test.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./internal/mediahls \
  -run '^TestRealSceneCutsPreservePackagerGeneration$' -count=1 -v -timeout=60s \
  >"$report/negative-control.log" 2>&1 || negative_exit=$?
printf 'negative_control_exit_code=%s\n' "$negative_exit" | tee "$report/negative-control-exit.txt"
# These are uncredentialed local fixture logs, with no room/session requests.
cat "$report/negative-control.log"
[[ "$negative_exit" == 1 ]] || fail 'baseline did not reproduce the expected test failure; review before further repair/deployment'
grep -Fq 'HLS transcode timeline gap' "$report/negative-control.log" || fail 'baseline failed for another reason; review the fixture report'
printf 'PASS negative control: prior code reproduced the timeline-gap failure\n'

positive_exit=0
docker run --rm --network none --workdir /src \
  --mount "type=bind,src=$report/codec_integration_test.go,dst=/src/internal/mediahls/codec_integration_test.go,readonly" \
  --mount "type=bind,src=$report/transcoder.go,dst=/src/internal/mediahls/transcoder.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./pkg/gst ./internal/mediahls \
  -run 'TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture|TestRealCodecsReachConventionalPackagerReadiness|TestRealSceneCutsPreservePackagerGeneration' \
  -count=1 -v -timeout=90s >"$report/positive-control.log" 2>&1 || positive_exit=$?
printf 'positive_control_exit_code=%s\n' "$positive_exit" | tee "$report/positive-control-exit.txt"
cat "$report/positive-control.log"
[[ "$positive_exit" == 0 ]] || fail 'repaired fixture gate failed; keep live HLS disabled'
printf 'GOP A/B GATE PASSED; checkout and running service unchanged. Application build/deployment/live acceptance remain pending. Evidence: %s\n' "$report"
