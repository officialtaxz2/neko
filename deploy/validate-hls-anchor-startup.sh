#!/usr/bin/env bash

# Target-only ordered startup regression and real-codec comparison. No deploy.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash validate-hls-anchor-startup.sh REPOSITORY OUTPUT_DIR REPAIR_COMMIT'
umask 077
for required in docker git realpath stat grep tr; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 97ba4ad9ab3e635da936a58c8a7ec795da05ba46 ]] || fail 'keep the application at the restored 97ba4ad9 baseline'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly repair_commit="$3"
readonly registry_repair_commit=48f4acf2685b2247e8606deaedb162b9380afeb9
[[ "$repair_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full repair commit required'
git cat-file -e "$repair_commit^{commit}"
git merge-base --is-ancestor "$registry_repair_commit" "$repair_commit" || fail 'repair must retain the reviewed registry repair history'
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-anchor-startup.sh")" ]] || fail 'helper differs from the selected repair commit'
[[ "$(git rev-parse "$repair_commit:server/pkg/gst/gst.go")" == "$(git rev-parse "$registry_repair_commit:server/pkg/gst/gst.go")" ]] || fail 'registry correction changed; reassess the comparison'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ "$(cat "$output/validation-commit.txt")" == "$application_commit" ]] || fail 'baseline preparation marker differs'
readonly report="$output/anchor-startup-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
readonly image="$(docker image inspect --format '{{.Id}}' my-neko/hls-validation:testing)"
[[ "$image" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'existing codec-validation image required'
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/codec_integration_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image sources differ from the 97ba4ad9 baseline'
done
printf 'application_commit=%s\nrepair_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\n' \
  "$application_commit" "$repair_commit" "$image" "$(git hash-object -- "$helper")" | tee "$report/provenance.txt"
mounts=()
for source in pkg/gst/gst.go pkg/gst/registry_integration_test.go internal/mediahls/codec_integration_test.go internal/mediahls/packager_startup_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  git show "$repair_commit:server/$source" >"$record"
  printf '%s_blob=%s\n' "$source" "$(git hash-object -- "$record")" | tee -a "$report/provenance.txt"
  mounts+=(--mount "type=bind,src=$record,dst=/src/$source,readonly")
done
git show "$repair_commit:server/internal/mediahls/packager.go" >"$report/packager.go"
printf 'packager_repair_blob=%s\n' "$(git hash-object -- "$report/packager.go")" | tee -a "$report/provenance.txt"

# Both sides retain the same registry fix, test observations, encoders and
# deadlines. Only packager.go changes in the positive comparison.
negative_exit=0
docker run --rm --network none --workdir /src "${mounts[@]}" \
  --entrypoint go "$image" test -tags hlsintegration ./internal/mediahls \
  -run '^TestPackagerRetainsInitialSamplesUntilAnchor$' -count=1 -v -timeout=30s \
  >"$report/negative-control.log" 2>&1 || negative_exit=$?
printf 'negative_control_exit_code=%s\n' "$negative_exit" | tee "$report/negative-control-exit.txt"
cat "$report/negative-control.log"
[[ "$negative_exit" == 1 ]] || fail 'prior packager did not reproduce the ordered startup failure'
grep -Fq 'initial pre-anchor IDR was lost; three parents were not ready at 18 seconds' "$report/negative-control.log" || fail 'negative control failed for a different reason'
printf 'PASS negative control: prior packager lost the initial pre-anchor IDR\n'

positive_exit=0
docker run --rm --network none --workdir /src "${mounts[@]}" \
  --mount "type=bind,src=$report/packager.go,dst=/src/internal/mediahls/packager.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./pkg/gst ./internal/mediahls \
  -run 'TestNativeParseFailureDoesNotWaitForSampleRegistry|TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture|TestPackagerRetainsInitialSamplesUntilAnchor|TestPackagerAnchorWait|TestRealCodecsReachConventionalPackagerReadiness|TestRealSceneCutsPreservePackagerGeneration' \
  -count=3 -parallel=1 -v -timeout=180s >"$report/positive-control.log" 2>&1 || positive_exit=$?
printf 'positive_control_exit_code=%s\n' "$positive_exit" | tee "$report/positive-control-exit.txt"
cat "$report/positive-control.log"
[[ "$positive_exit" == 0 ]] || fail 'corrected anchor/registry/codec gate failed; keep live HLS disabled'
for test in TestNativeParseFailureDoesNotWaitForSampleRegistry TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture TestPackagerRetainsInitialSamplesUntilAnchor TestPackagerAnchorWaitCancels TestPackagerAnchorWaitHonorsOverflow TestRealCodecsReachConventionalPackagerReadiness TestRealSceneCutsPreservePackagerGeneration; do
  passed="$(grep -c "^--- PASS: $test " "$report/positive-control.log" || true)"
  [[ "$passed" == 3 ]] || fail 'required positive test did not pass all three repetitions'
done
printf 'ANCHOR STARTUP A/B GATE PASSED; checkout, live service and prepared-image markers unchanged. Full application checks/image, native capture and enabled HLS acceptance remain pending. Evidence: %s\n' "$report"
