#!/usr/bin/env bash

# Target-only observations of the failed 414639d2 cold-start gate. No live
# service access, build, checkout change or preparation-marker replacement.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash diagnose-hls-worker-startup.sh REPOSITORY OUTPUT_DIR HELPER_COMMIT'
umask 077
for required in docker git realpath stat grep tr date mkdir cat tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 414639d2493ad1d2106399337c647d1b007f1e9d ]] || fail 'keep the checkout at the failed exact 414639d2 preparation'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly helper_commit="$3"
[[ "$helper_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full helper commit required'
git cat-file -e "$helper_commit^{commit}"
git merge-base --is-ancestor "$application_commit" "$helper_commit" || fail 'helper must retain the reviewed application history'
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$helper_commit:deploy/diagnose-hls-worker-startup.sh")" ]] || fail 'helper differs from the selected commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ -f "$output/validation-commit.txt" && "$(cat "$output/validation-commit.txt")" == PENDING ]] || fail 'expected failed preparation marker PENDING'
readonly report="$output/worker-startup-diagnostic-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
readonly image="$(docker image inspect --format '{{.Id}}' my-neko/hls-validation:testing)"
[[ "$image" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'existing codec-validation image required'

# Full preparation rebuilt this mutable tag. Pin its image ID and verify eight
# source blobs against 414639d2 instead of assuming the old 97ba4ad9 image.
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/codec_integration_test.go internal/mediahls/packager_startup_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image sources differ from the exact failed preparation'
done
git show "$helper_commit:server/internal/mediahls/packager.go" >"$report/packager.go"
git show "$helper_commit:server/internal/mediahls/codec_integration_test.go" >"$report/codec_integration_test.go"
printf 'application_commit=%s\nhelper_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\npackager_observation_blob=%s\nfixture_observation_blob=%s\n' \
  "$application_commit" "$helper_commit" "$image" "$(git hash-object -- "$helper")" \
  "$(git hash-object -- "$report/packager.go")" "$(git hash-object -- "$report/codec_integration_test.go")" | tee "$report/provenance.txt"

required_tests=(
  TestNativeParseFailureDoesNotWaitForSampleRegistry
  TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture
  TestPackagerRetainsInitialSamplesUntilAnchor
  TestPackagerAnchorWaitCancels
  TestPackagerAnchorWaitHonorsOverflow
  TestRealCodecsReachConventionalPackagerReadiness
)
readonly selection='TestNativeParseFailureDoesNotWaitForSampleRegistry|TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture|TestPackagerRetainsInitialSamplesUntilAnchor|TestPackagerAnchorWait|TestRealCodecsReachConventionalPackagerReadiness'

# Fixed three fresh containers/processes, not three warmed starts in one Go
# process and not retry-until-pass. Mount observations only; production queues,
# admission, clocks, codecs, restart decisions and 24 s deadline are unchanged.
for attempt in 1 2 3; do
  status=0
  docker run --rm --network none --workdir /src \
    --mount "type=bind,src=$report/packager.go,dst=/src/internal/mediahls/packager.go,readonly" \
    --mount "type=bind,src=$report/codec_integration_test.go,dst=/src/internal/mediahls/codec_integration_test.go,readonly" \
    --entrypoint go "$image" test -tags hlsintegration ./pkg/gst ./internal/mediahls \
    -run "$selection" -count=1 -v -timeout=60s >"$report/cold-$attempt.log" 2>&1 || status=$?
  printf 'cold_attempt=%s test_exit_code=%s\n' "$attempt" "$status" | tee "$report/cold-$attempt-exit.txt"
  # Only synthetic fixtures run here; no credentials, live URLs or sessions.
  cat "$report/cold-$attempt.log"
  [[ "$status" == 0 || "$status" == 1 ]] || fail 'diagnostic container did not finish normally'
  failures=0
  for test_name in "${required_tests[@]}"; do
    passed="$(grep -c "^--- PASS: $test_name " "$report/cold-$attempt.log" || true)"
    failed="$(grep -c "^--- FAIL: $test_name " "$report/cold-$attempt.log" || true)"
    [[ $((passed + failed)) -eq 1 ]] || fail 'incomplete diagnostic test outcomes; inspect private log'
    failures=$((failures + failed))
  done
  [[ "$(grep -c 'CODEC_DIAGNOSTIC packager ' "$report/cold-$attempt.log" || true)" == 1 ]] || fail 'missing startup snapshot'
  [[ "$status" == 0 && "$failures" == 0 || "$status" == 1 && "$failures" -gt 0 ]] || fail 'exit status does not match the diagnostic test outcomes'
  printf 'cold_attempt=%s failed_checks=%s\n' "$attempt" "$failures" | tee "$report/cold-$attempt-counts.txt"
done
printf 'WORKER STARTUP DIAGNOSTIC COMPLETE; exit 0 means observations completed, not acceptance. Full preparation remains FAILED/PENDING. Keep HLS disabled; checkout, preparation markers and running service unchanged. Evidence: %s\n' "$report"
