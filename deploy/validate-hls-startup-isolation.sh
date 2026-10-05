#!/usr/bin/env bash

# Target-only registry/codec comparison. Never replace the running service.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash validate-hls-startup-isolation.sh REPOSITORY OUTPUT_DIR REPAIR_COMMIT'
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
[[ "$repair_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full repair commit required'
git cat-file -e "$repair_commit^{commit}"
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-startup-isolation.sh")" ]] || fail 'helper differs from the selected repair commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ "$(cat "$output/validation-commit.txt")" == "$application_commit" ]] || fail 'baseline preparation marker differs'
readonly report="$output/startup-isolation-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
readonly image="$(docker image inspect --format '{{.Id}}' my-neko/hls-validation:testing)"
[[ "$image" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'existing codec-validation image required'
printf 'application_commit=%s\nrepair_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\n' \
  "$application_commit" "$repair_commit" "$image" "$(git hash-object -- "$helper")" \
  | tee "$report/provenance.txt"

# Verify the mutable tag's source before comparing. Every run below then uses
# the pinned image ID; there is no image rebuild or production service access.
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/codec_integration_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image sources differ from the 97ba4ad9 baseline'
done
git show "$repair_commit:server/pkg/gst/registry_integration_test.go" >"$report/registry_integration_test.go"
git show "$repair_commit:server/pkg/gst/gst.go" >"$report/gst.go"

negative_exit=0
docker run --rm --network none --workdir /src \
  --mount "type=bind,src=$report/registry_integration_test.go,dst=/src/pkg/gst/registry_integration_test.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./pkg/gst \
  -run '^TestNativeParseFailureDoesNotWaitForSampleRegistry$' -count=1 -v -timeout=30s \
  >"$report/negative-control.log" 2>&1 || negative_exit=$?
printf 'negative_control_exit_code=%s\n' "$negative_exit" | tee "$report/negative-control-exit.txt"
# Only uncredentialed local fixtures run; these logs contain no live sessions.
cat "$report/negative-control.log"
[[ "$negative_exit" == 1 ]] || fail 'baseline did not reproduce the expected registry-coupling failure'
grep -Fq 'native pipeline parse waited for sample registry' "$report/negative-control.log" || fail 'baseline failed for another reason; review before repair/deployment'
printf 'PASS negative control: prior constructor waited for the shared sample registry\n'

positive_exit=0
docker run --rm --network none --workdir /src \
  --mount "type=bind,src=$report/registry_integration_test.go,dst=/src/pkg/gst/registry_integration_test.go,readonly" \
  --mount "type=bind,src=$report/gst.go,dst=/src/pkg/gst/gst.go,readonly" \
  --entrypoint go "$image" test -tags hlsintegration ./pkg/gst ./internal/mediahls \
  -run 'TestNativeParseFailureDoesNotWaitForSampleRegistry|TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture|TestRealCodecsReachConventionalPackagerReadiness|TestRealSceneCutsPreservePackagerGeneration' \
  -count=1 -v -timeout=90s >"$report/positive-control.log" 2>&1 || positive_exit=$?
printf 'positive_control_exit_code=%s\n' "$positive_exit" | tee "$report/positive-control-exit.txt"
cat "$report/positive-control.log"
[[ "$positive_exit" == 0 ]] || fail 'repaired registry/codec fixture gate failed; keep live HLS disabled'
printf 'STARTUP ISOLATION A/B GATE PASSED; checkout and running service unchanged. Capture-stage tracing, application build/deployment and live HLS acceptance remain pending. Evidence: %s\n' "$report"
