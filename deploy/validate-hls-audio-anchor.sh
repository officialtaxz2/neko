#!/usr/bin/env bash

# Target-only AAC anchor A/B. Keep the failed 414 checkout and live HLS off.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash validate-hls-audio-anchor.sh REPOSITORY OUTPUT_DIR REPAIR_COMMIT'
umask 077
for required in docker git realpath stat grep tr date mkdir cat tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 414639d2493ad1d2106399337c647d1b007f1e9d ]] || fail 'keep checkout at the failed exact 414639d2 preparation'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly repair_commit="$3"
readonly observation_commit=5303449509b05aaeaabb6d6d5d2f7b6eb94fbc19
[[ "$repair_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full repair commit required'
git cat-file -e "$repair_commit^{commit}"
git merge-base --is-ancestor "$observation_commit" "$repair_commit" || fail 'repair must retain the reviewed observation history'
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-audio-anchor.sh")" ]] || fail 'helper differs from selected repair commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ -f "$output/validation-commit.txt" && "$(cat "$output/validation-commit.txt")" == PENDING ]] || fail 'expected failed preparation marker PENDING'
readonly report="$output/audio-anchor-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
# The supplied worker diagnostic verified this exact image's eight 414 blobs.
# Pin by ID; future mutable validation-tag changes must not move this A/B.
readonly image=sha256:d6e7e49a1423ec92994cf3c82559936ae6ff10bd0080be776cc391696355bc3a
[[ "$(docker image inspect --format '{{.Id}}' "$image")" == "$image" ]] || fail 'recorded 414 codec image missing'
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/codec_integration_test.go internal/mediahls/packager_startup_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image source differs from the failed exact preparation'
done
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/transcoder.go; do
  [[ "$(git rev-parse "$repair_commit:server/$source")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'A/B requires unchanged native/capture/encoder sources'
done
printf 'application_commit=%s\nrepair_commit=%s\nobservation_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\n' \
  "$application_commit" "$repair_commit" "$observation_commit" "$image" "$(git hash-object -- "$helper")" | tee "$report/provenance.txt"
mounts=()
for source in internal/mediahls/codec_integration_test.go internal/mediahls/packager_startup_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  git show "$repair_commit:server/$source" >"$record"
  printf '%s_blob=%s\n' "$source" "$(git hash-object -- "$record")" | tee -a "$report/provenance.txt"
  mounts+=(--mount "type=bind,src=$record,dst=/src/$source,readonly")
done
git show "$observation_commit:server/internal/mediahls/packager.go" >"$report/prior-packager.go"
git show "$repair_commit:server/internal/mediahls/packager.go" >"$report/repaired-packager.go"
printf 'prior_packager_blob=%s\nrepaired_packager_blob=%s\n' \
  "$(git hash-object -- "$report/prior-packager.go")" "$(git hash-object -- "$report/repaired-packager.go")" | tee -a "$report/provenance.txt"

# Both negative controls use the same new fixtures and old behavior with fixed
# logging. An unexpected pass or failure stops the gate; it is not a retry.
negative_unit=0
docker run --rm --network none --workdir /src "${mounts[@]}" \
  --mount "type=bind,src=$report/prior-packager.go,dst=/src/internal/mediahls/packager.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./internal/mediahls \
  -run '^TestPackagerAudioDrainsBeforeAnchor$' -count=1 -v -timeout=15s \
  >"$report/negative-unit.log" 2>&1 || negative_unit=$?
printf 'negative_unit_exit_code=%s\n' "$negative_unit" | tee "$report/negative-unit-exit.txt"
cat "$report/negative-unit.log"
[[ "$negative_unit" == 1 ]] || fail 'old audio hold did not reproduce the unit failure'
grep -Fq 'audio output waited for the high anchor' "$report/negative-unit.log" || fail 'negative unit failed for a different reason'

negative_codec=0
docker run --rm --network none --workdir /src "${mounts[@]}" \
  --mount "type=bind,src=$report/prior-packager.go,dst=/src/internal/mediahls/packager.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./internal/mediahls \
  -run '^TestRealDelayedHighAnchorPreservesPackagerGeneration$' -count=1 -v -timeout=40s \
  >"$report/negative-codec.log" 2>&1 || negative_codec=$?
printf 'negative_codec_exit_code=%s\n' "$negative_codec" | tee "$report/negative-codec-exit.txt"
cat "$report/negative-codec.log"
[[ "$negative_codec" == 1 ]] || fail 'delayed-high old codec did not reproduce the failure'
grep -Fq 'cold fixture caused 2 packager generations' "$report/negative-codec.log" || fail 'negative codec failed for a different reason'
grep -Fq '"variant":"audio","stage":"anchor","reason":"queue_full"' "$report/negative-codec.log" || fail 'negative codec did not identify the expected audio anchor overflow'
printf 'PASS negative controls: old AAC hold blocked drainage and restarted the delayed-high fixture\n'

required_tests=(
  TestNativeParseFailureDoesNotWaitForSampleRegistry
  TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture
  TestPackagerRetainsInitialSamplesUntilAnchor
  TestPackagerAnchorWaitCancels
  TestPackagerAnchorWaitHonorsOverflow
  TestPackagerAudioDrainsBeforeAnchor
  TestRealCodecsReachConventionalPackagerReadiness
  TestRealDelayedHighAnchorPreservesPackagerGeneration
)
readonly selection='TestNativeParseFailureDoesNotWaitForSampleRegistry|TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture|TestPackagerRetainsInitialSamplesUntilAnchor|TestPackagerAnchorWait|TestPackagerAudioDrainsBeforeAnchor|TestRealCodecsReachConventionalPackagerReadiness|TestRealDelayedHighAnchorPreservesPackagerGeneration'
# Fixed three cold processes. Scene cuts are checked once in the first process;
# the two cold-start cases and lifecycle/timestamp checks run in every process.
for attempt in 1 2 3; do
  selected="$selection"
  if [[ "$attempt" == 1 ]]; then selected+='|TestRealSceneCutsPreservePackagerGeneration'; fi
  positive_exit=0
  docker run --rm --network none --workdir /src "${mounts[@]}" \
    --mount "type=bind,src=$report/repaired-packager.go,dst=/src/internal/mediahls/packager.go,readonly" \
    --entrypoint go "$image" test -tags hlsintegration ./pkg/gst ./internal/mediahls \
    -run "$selected" -count=1 -v -timeout=90s >"$report/positive-$attempt.log" 2>&1 || positive_exit=$?
  printf 'positive_cold_attempt=%s exit_code=%s\n' "$attempt" "$positive_exit" | tee "$report/positive-$attempt-exit.txt"
  cat "$report/positive-$attempt.log"
  [[ "$positive_exit" == 0 ]] || fail 'corrected AAC/anchor/codec gate failed; keep live HLS disabled'
  for test_name in "${required_tests[@]}"; do
    [[ "$(grep -c "^--- PASS: $test_name " "$report/positive-$attempt.log" || true)" == 1 ]] || fail 'required positive check did not pass once'
  done
  if [[ "$attempt" == 1 ]]; then
    [[ "$(grep -c '^--- PASS: TestRealSceneCutsPreservePackagerGeneration ' "$report/positive-$attempt.log" || true)" == 1 ]] || fail 'scene-cut check missing'
  fi
done
printf 'AUDIO ANCHOR A/B GATE PASSED; controlled mechanism only, not proof of the earlier unobserved worker failure or live outage. Checkout, preparation markers and live service unchanged. Full preparation and enabled acceptance pending. Evidence: %s\n' "$report"
