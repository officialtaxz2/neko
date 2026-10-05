#!/usr/bin/env bash

# Target-only synthetic startup comparison. This is evidence collection, not
# an acceptance gate, and never enables HLS or replaces the running service.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash diagnose-hls-codec-startup.sh REPOSITORY OUTPUT_DIR HELPER_COMMIT'
umask 077
for required in docker git realpath stat grep tr; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
readonly helper="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
readonly application_commit="$(git rev-parse HEAD)"
[[ "$application_commit" == 97ba4ad9ab3e635da936a58c8a7ec795da05ba46 ]] || fail 'keep the application at the restored 97ba4ad9 baseline'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
readonly helper_commit="$3"
readonly registry_repair_commit=48f4acf2685b2247e8606deaedb162b9380afeb9
[[ "$helper_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full helper commit required'
git cat-file -e "$helper_commit^{commit}"
git merge-base --is-ancestor "$registry_repair_commit" "$helper_commit" || fail 'helper must retain the reviewed registry repair history'
[[ "$(git hash-object -- "$helper")" == "$(git rev-parse "$helper_commit:deploy/diagnose-hls-codec-startup.sh")" ]] || fail 'helper differs from the selected commit'
readonly repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
readonly output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ "$(cat "$output/validation-commit.txt")" == "$application_commit" ]] || fail 'baseline preparation marker differs'
readonly report="$output/codec-startup-diagnostic-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
readonly image="$(docker image inspect --format '{{.Id}}' my-neko/hls-validation:testing)"
[[ "$image" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'existing codec-validation image required'

# Pin the existing image and confirm the exact unchanged source before mounts.
for source in pkg/gst/gst.go pkg/gst/gst.c pkg/gst/gst.h internal/capture/streamsink.go internal/mediahls/packager.go internal/mediahls/transcoder.go internal/mediahls/codec_integration_test.go; do
  record="$report/$(printf '%s' "$source" | tr / _)"
  docker run --rm --network none --entrypoint cat "$image" "/src/$source" >"$record"
  [[ "$(git hash-object -- "$record")" == "$(git rev-parse "$application_commit:server/$source")" ]] || fail 'codec image sources differ from the 97ba4ad9 baseline'
done
git show "$registry_repair_commit:server/pkg/gst/gst.go" >"$report/gst.go"
git show "$helper_commit:server/internal/mediahls/codec_integration_test.go" >"$report/codec_integration_test.go"
printf 'application_commit=%s\nhelper_commit=%s\nregistry_repair_commit=%s\ncodec_image_id=%s\nhelper_blob=%s\nfixture_blob=%s\nregistry_repair_blob=%s\n' \
  "$application_commit" "$helper_commit" "$registry_repair_commit" "$image" \
  "$(git hash-object -- "$helper")" "$(git hash-object -- "$report/codec_integration_test.go")" \
  "$(git hash-object -- "$report/gst.go")" | tee "$report/provenance.txt"

# Three sequential fresh packagers per constructor variant expose startup
# ordering without a retry-until-pass rule or changes to the 24 s deadline.
# Both variants use identical observations and the production codec/packager.
for variant in original isolated; do
  mounts=(--mount "type=bind,src=$report/codec_integration_test.go,dst=/src/internal/mediahls/codec_integration_test.go,readonly")
  if [[ "$variant" == isolated ]]; then
    mounts+=(--mount "type=bind,src=$report/gst.go,dst=/src/pkg/gst/gst.go,readonly")
  fi
  status=0
  docker run --rm --network none --workdir /src "${mounts[@]}" \
    --entrypoint go "$image" test -tags hlsintegration ./internal/mediahls \
    -run '^TestRealCodecsReachConventionalPackagerReadiness$' \
    -count=3 -parallel=1 -v -timeout=120s >"$report/$variant.log" 2>&1 || status=$?
  printf '%s_exit_code=%s\n' "$variant" "$status" | tee "$report/$variant-exit.txt"
  # Synthetic fixtures only: no live credentials, URLs or room sessions.
  cat "$report/$variant.log"
  [[ "$status" == 0 || "$status" == 1 ]] || fail 'diagnostic container did not finish normally'
  passed="$(grep -c '^--- PASS: TestRealCodecsReachConventionalPackagerReadiness ' "$report/$variant.log" || true)"
  failed="$(grep -c '^--- FAIL: TestRealCodecsReachConventionalPackagerReadiness ' "$report/$variant.log" || true)"
  snapshots="$(grep -c 'CODEC_DIAGNOSTIC packager ' "$report/$variant.log" || true)"
  [[ $((passed + failed)) -eq 3 && "$snapshots" == 3 ]] || fail 'incomplete codec diagnostic; review logs before any further action'
  [[ "$status" == 0 && "$failed" == 0 || "$status" == 1 && "$failed" -gt 0 ]] || fail 'exit status does not match the three codec outcomes'
  printf '%s_passed=%s\n%s_failed=%s\n' "$variant" "$passed" "$variant" "$failed" | tee "$report/$variant-counts.txt"
done
printf 'CODEC STARTUP DIAGNOSTIC COMPLETE; the earlier acceptance gate remains FAILED. Keep live HLS disabled; checkout, prepared-image markers and running service unchanged. Evidence: %s\n' "$report"
