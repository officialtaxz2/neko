#!/usr/bin/env bash

# Target-only common-video-source A/B. Do not touch the working default-off
# service, application checkout, Caddy configuration or preparation markers.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash validate-hls-shared-clock.sh REPOSITORY OUTPUT_DIR REPAIR_COMMIT'
umask 077
for required in docker git realpath stat grep tr date mkdir cat tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 71a14d2174dafbc12b1880adde6dc68176bfe9af ]] || fail 'keep checkout at the working default-off exact-71 application'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly repair_commit="$3"
readonly diagnostic_commit=409482b47e4ed962a2f9a3fd129114a53563f15a
[[ "$repair_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full repair commit required'
git cat-file -e "$repair_commit^{commit}"
git merge-base --is-ancestor "$diagnostic_commit" "$repair_commit" || fail 'repair must retain the reviewed diagnostic history'
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-shared-clock.sh")" ]] || fail 'helper differs from selected repair commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ -f "$output/validation-commit.txt" && "$(cat "$output/validation-commit.txt")" == "$application_commit" ]] || fail 'successful exact-71 preparation marker required'
readonly report="$output/shared-clock-ab-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
# The supplied clock-skew diagnostic recorded this immutable exact-71 image.
readonly image=sha256:2c885aa463ee9514b120d541e9852f4e9cd5cc334a72e9373cacf6896c963c48
[[ "$(docker image inspect --format '{{.Id}}' "$image")" == "$image" ]] || fail 'recorded exact-71 codec image missing'
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/model.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/packager_test.go internal/mediahls/codec_integration_test.go internal/mediahls/packager_startup_test.go internal/mediahls/cold_start_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image source differs from exact-71 preparation'
done
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/transcoder.go; do
  [[ "$(git rev-parse "$repair_commit:server/$source")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'A/B requires unchanged native/capture/encoder sources'
done
printf 'application_commit=%s\nrepair_commit=%s\ndiagnostic_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\n' \
  "$application_commit" "$repair_commit" "$diagnostic_commit" "$image" "$(git hash-object -- "$helper")" | tee "$report/provenance.txt"

negative_mounts=()
for source in internal/mediahls/codec_integration_test.go internal/mediahls/codec_clock_skew_diagnostic_test.go; do
  record="$report/negative_$(printf '%s' "$source" | tr / _)"
  git show "$diagnostic_commit:server/$source" >"$record"
  printf 'negative_%s_blob=%s\n' "$source" "$(git hash-object -- "$record")" | tee -a "$report/provenance.txt"
  negative_mounts+=(--mount "type=bind,src=$record,dst=/src/$source,readonly")
done
negative_status=0
docker run --rm --network none --workdir /src "${negative_mounts[@]}" \
  --entrypoint go "$image" test -tags 'hlsintegration hlsdiagnostic' ./internal/mediahls \
  -run '^TestDiagnosticSkewedVideoKeyframesBlockReadiness$' -count=1 -v -timeout=40s \
  >"$report/negative.log" 2>&1 || negative_status=$?
printf 'negative_signature_exit_code=%s\n' "$negative_status" | tee "$report/negative-exit.txt"
cat "$report/negative.log"
[[ "$negative_status" == 0 ]] || fail 'old source-phase failure did not reproduce with its required signature'
[[ "$(grep -c '^--- PASS: TestDiagnosticSkewedVideoKeyframesBlockReadiness ' "$report/negative.log" || true)" == 1 ]] || fail 'negative signature check did not complete once'
grep -Fq 'PHASE_DIAGNOSTIC reproduced not-ready: audio/high ready; medium/low initial admission blocked despite flowing IDRs; generation=1' "$report/negative.log" || fail 'negative signature missing'

mounts=()
for source in internal/mediahls/model.go internal/mediahls/packager.go internal/mediahls/packager_test.go internal/mediahls/shared_video_input_test.go internal/mediahls/codec_integration_test.go; do
  record="$report/repaired_$(printf '%s' "$source" | tr / _)"
  git show "$repair_commit:server/$source" >"$record"
  printf 'repaired_%s_blob=%s\n' "$source" "$(git hash-object -- "$record")" | tee -a "$report/provenance.txt"
  mounts+=(--mount "type=bind,src=$record,dst=/src/$source,readonly")
done
required_tests=(
  TestNativeParseFailureDoesNotWaitForSampleRegistry
  TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture
  TestPackagerSharesOneWorkerSetAndPublishesBothModes
  TestColdCaptureStartBindsOpenedGenerationAndInitialCaps
  TestColdCapsWithDifferentGenerationOrDimensionsStillRestart
  TestPackagerRetainsInitialSamplesUntilAnchor
  TestPackagerAnchorWaitCancels
  TestPackagerAnchorWaitHonorsOverflow
  TestPackagerAudioDrainsBeforeAnchor
  TestSharedVideoInputPreservesUnitsAndReportsPeerFailure
  TestSharedVideoConstructionFailureClosesOwnedResources
  TestSharedVideoConstructionStopsBothInputsAndAllEncoders
  TestRealCodecsReachConventionalPackagerReadiness
  TestRealDelayedHighAnchorPreservesPackagerGeneration
  TestRealSkewedSourcesShareVideoClock
)
readonly selection='TestNativeParseFailureDoesNotWaitForSampleRegistry|TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture|TestPackagerSharesOneWorkerSet|TestColdCaptureStart|TestColdCapsWithDifferent|TestPackagerRetainsInitialSamplesUntilAnchor|TestPackagerAnchorWait|TestPackagerAudioDrainsBeforeAnchor|TestSharedVideoInput|TestSharedVideoConstruction|TestRealCodecsReachConventionalPackagerReadiness|TestRealDelayedHighAnchorPreservesPackagerGeneration|TestRealSkewedSourcesShareVideoClock'
for attempt in 1 2 3; do
  selected="$selection"
  if [[ "$attempt" == 1 ]]; then selected+='|TestRealSceneCutsPreservePackagerGeneration'; fi
  positive_status=0
  docker run --rm --network none --workdir /src "${mounts[@]}" \
    --entrypoint go "$image" test -tags hlsintegration ./pkg/gst ./internal/mediahls \
    -run "$selected" -count=1 -v -timeout=120s \
    >"$report/positive-$attempt.log" 2>&1 || positive_status=$?
  printf 'positive_cold_attempt=%s exit_code=%s\n' "$attempt" "$positive_status" | tee "$report/positive-$attempt-exit.txt"
  cat "$report/positive-$attempt.log"
  [[ "$positive_status" == 0 ]] || fail 'shared-clock repair gate failed; keep live HLS disabled'
  for test_name in "${required_tests[@]}"; do
    [[ "$(grep -c "^--- PASS: $test_name " "$report/positive-$attempt.log" || true)" == 1 ]] || fail 'required positive check did not pass once'
  done
  if [[ "$attempt" == 1 ]]; then
    [[ "$(grep -c '^--- PASS: TestRealSceneCutsPreservePackagerGeneration ' "$report/positive-$attempt.log" || true)" == 1 ]] || fail 'scene-cut check missing'
  fi
done
printf 'SHARED VIDEO CLOCK A/B GATE PASSED; controlled mechanism only, not measured live capture skew or HLS playback acceptance. Checkout, preparation markers and running service unchanged. Keep HLS disabled; full exact-repair preparation and live acceptance remain pending. Evidence: %s\n' "$report"
