#!/usr/bin/env bash

# Real target server only. This prepares images; it never replaces the service.
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
[[ $# -eq 1 ]] || { printf 'Usage: bash deploy/validate-hls-phase4.sh OUTPUT_DIR\n' >&2; exit 2; }
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
readonly image_tag="hls-${NEKO_VALIDATION_COMMIT:0:12}"
bash deploy/collect-hls-media.sh init "$1"
readonly output="$(realpath -m -- "$1")"
umask 077
printf 'PENDING\n' >"$output/validation-commit.txt"
exec > >(tee "$output/validation.log") 2>&1
printf 'validation_commit=%s\nimage_tag=%s\n' "$NEKO_VALIDATION_COMMIT" "$image_tag"

docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
bash -n deploy/collect-hls-media.sh deploy/validate-hls-phase4.sh deploy/deploy-hls-media.sh
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks --help </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T client-checks </dev/null
docker compose -f docker-compose.validation.yaml build server-checks </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T server-checks </dev/null
docker compose -f docker-compose.validation.yaml build hls-packager-checks </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T --entrypoint gst-inspect-1.0 hls-packager-checks --version </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T hls-packager-checks </dev/null

# Findings normally return 1. Keep the private report for classification; do not
# confuse that code with a passing audit or apply npm audit fix automatically.
if docker compose -f docker-compose.validation.yaml run --rm -T dependency-audit </dev/null >"$output/dependency-audit.json"; then
  audit_status=0
else
  audit_status=$?
fi
printf 'dependency_audit_exit_code=%s (report requires review)\n' "$audit_status"
printf '%s\n' "$audit_status" >"$output/dependency-audit-exit-code.txt"

./build "my-neko/base:$image_tag" -y </dev/null
./build "my-neko/brave:$image_tag" -y </dev/null
docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
  --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
bash deploy/collect-hls-media.sh snapshot "$output" default-off-baseline
printf '%s\n' "$NEKO_VALIDATION_COMMIT" >"$output/validation-commit.txt"
printf 'AUTOMATED/IMAGE GATE PASSED; running service unchanged. Evidence: %s\n' "$output"
