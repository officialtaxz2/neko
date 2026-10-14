#!/usr/bin/env bash

# Target-only source-phase diagnostic. No live access, image build, checkout
# movement, marker replacement or changes to production clocks/admission.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash diagnose-hls-clock-skew.sh REPOSITORY OUTPUT_DIR HELPER_COMMIT'
umask 077
for required in docker git realpath stat grep tr date mkdir cat tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 71a14d2174dafbc12b1880adde6dc68176bfe9af ]] || fail 'keep checkout at the prepared exact 71a14d21 application'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly helper_commit="$3"
[[ "$helper_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full helper commit required'
git cat-file -e "$helper_commit^{commit}"
git merge-base --is-ancestor "$application_commit" "$helper_commit" || fail 'helper must retain the reviewed application history'
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$helper_commit:deploy/diagnose-hls-clock-skew.sh")" ]] || fail 'helper differs from selected commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ -f "$output/validation-commit.txt" && "$(cat "$output/validation-commit.txt")" == "$application_commit" ]] || fail 'successful exact-71 preparation marker required'
readonly report="$output/clock-skew-diagnostic-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
# Resolve the mutable validation tag once, then verify recorded sources and
# use only its immutable ID. Never substitute the Brave application image.
readonly image="$(docker image inspect --format '{{.Id}}' my-neko/hls-validation:testing)"
[[ "$image" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'existing codec-validation image missing'
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/codec_integration_test.go internal/mediahls/packager_startup_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image source differs from exact-71 preparation'
done
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go; do
  [[ "$(git rev-parse "$helper_commit:server/$source")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'diagnostic requires unchanged production/native sources'
done
printf 'application_commit=%s\nhelper_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\n' \
  "$application_commit" "$helper_commit" "$image" "$(git hash-object -- "$helper")" | tee "$report/provenance.txt"
mounts=()
for source in internal/mediahls/codec_integration_test.go internal/mediahls/codec_clock_skew_diagnostic_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  git show "$helper_commit:server/$source" >"$record"
  printf '%s_blob=%s\n' "$source" "$(git hash-object -- "$record")" | tee -a "$report/provenance.txt"
  mounts+=(--mount "type=bind,src=$record,dst=/src/$source,readonly")
done

control_status=0
docker run --rm --network none --workdir /src "${mounts[@]}" \
  --entrypoint go "$image" test -tags 'hlsintegration hlsdiagnostic' ./internal/mediahls \
  -run '^TestRealCodecsReachConventionalPackagerReadiness$' -count=1 -v -timeout=40s \
  >"$report/aligned-control.log" 2>&1 || control_status=$?
printf 'aligned_control_exit_code=%s\n' "$control_status" | tee "$report/aligned-control-exit.txt"
cat "$report/aligned-control.log"
[[ "$control_status" == 0 ]] || fail 'aligned control failed; inspect observations before another attempt'
[[ "$(grep -c '^--- PASS: TestRealCodecsReachConventionalPackagerReadiness ' "$report/aligned-control.log" || true)" == 1 ]] || fail 'aligned control did not pass once'

# Three predetermined cold processes, not retry-until-pass. Exit 0 here
# requires the expected defect signature, and never means HLS playback passed.
for attempt in 1 2 3; do
  diagnostic_status=0
  docker run --rm --network none --workdir /src "${mounts[@]}" \
    --entrypoint go "$image" test -tags 'hlsintegration hlsdiagnostic' ./internal/mediahls \
    -run '^TestDiagnosticSkewedVideoKeyframesBlockReadiness$' -count=1 -v -timeout=40s \
    >"$report/skew-$attempt.log" 2>&1 || diagnostic_status=$?
  printf 'skew_cold_attempt=%s exit_code=%s\n' "$attempt" "$diagnostic_status" | tee "$report/skew-$attempt-exit.txt"
  cat "$report/skew-$attempt.log"
  [[ "$diagnostic_status" == 0 ]] || fail 'source-phase hypothesis not reproduced with the required signature; inspect bounded observations'
  [[ "$(grep -c '^--- PASS: TestDiagnosticSkewedVideoKeyframesBlockReadiness ' "$report/skew-$attempt.log" || true)" == 1 ]] || fail 'skew diagnostic did not complete once'
  grep -Fq 'PHASE_DIAGNOSTIC reproduced not-ready: audio/high ready; medium/low initial admission blocked despite flowing IDRs; generation=1' "$report/skew-$attempt.log" || fail 'expected phase signature missing'
done
printf 'CLOCK SKEW DIAGNOSIS COMPLETE; exit 0 reproduces the controlled defect, not HLS acceptance or measured live capture skew. Production sources, checkout, preparation markers and running service unchanged. Keep HLS disabled. Evidence: %s\n' "$report"
