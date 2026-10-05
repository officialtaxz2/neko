#!/usr/bin/env bash

# Target server only: assemble the client-only repair image using its passed
# client gate and unchanged, already validated backend. Never replace Neko.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 5 ]] || fail 'Usage: bash prepare-hls-client-repair.sh REPOSITORY BASE_OUTPUT CLIENT_REPORT OUTPUT_DIR HELPER_COMMIT'
umask 077
for required in docker git realpath stat grep tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
[[ "$PWD" == "$repository" ]] || fail 'REPOSITORY must be the Git root'
base_commit=a7ffb8b13448a8329c6df24fdcb182ac32ca398c
repair_commit=73d5ff6d29110e3dd06999726a7e88718d09ea23
helper_commit="$5"
[[ "$helper_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full helper commit required'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
case "$(git rev-parse HEAD)" in
  "$base_commit"|"$repair_commit") ;;
  *) fail 'unexpected application checkout';;
esac
git cat-file -e "$repair_commit^{commit}"
[[ "$(git hash-object -- "$helper_source")" == "$(git rev-parse "$helper_commit:deploy/prepare-hls-client-repair.sh")" ]] || fail 'helper differs from selected tooling commit'
git merge-base --is-ancestor "$base_commit" "$repair_commit"
git merge-base --is-ancestor "$repair_commit" "$helper_commit"
changed_files="$(git diff --name-only "$base_commit" "$repair_commit")"
while IFS= read -r file; do
  case "$file" in
    client/src/neko/hls/controller.ts|client/tests/hls-controller.test.mjs|deploy/validate-hls-client-readiness.sh|AGENTS.md|README.md|docs/*) ;;
    *) fail 'application repair changed files outside the bounded client/readiness block';;
  esac
done <<< "$changed_files"
base_output="$(realpath -- "$2")"
client_report="$(realpath -- "$3")"
output="$(realpath -m -- "$4")"
for private_dir in "$base_output" "$client_report"; do
  [[ -d "$private_dir" && ! -L "$private_dir" && "$(stat -c %a -- "$private_dir")" == 700 ]] || fail 'private baseline/client evidence required'
  case "$private_dir/" in "$repository/"*) fail 'evidence must be outside Git';; esac
  case "$repository/" in "$private_dir/"*) fail 'evidence must not contain Git';; esac
done
case "$client_report/" in "$base_output/"client-readiness-*/) ;; *) fail 'client report must belong to the baseline output';; esac
[[ "$(cat "$base_output/validation-commit.txt")" == "$base_commit" ]] || fail 'exact baseline preparation required'
[[ "$(cat "$client_report/base-commit.txt")" == "$base_commit" && "$(cat "$client_report/repair-commit.txt")" == "$repair_commit" && "$(cat "$client_report/client-check-commit.txt")" == "$repair_commit" ]] || fail 'passed exact client repair gate required'
[[ "$(cat "$client_report/helper-blob.txt")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-client-readiness.sh")" ]] || fail 'client gate helper provenance differs'
for source in controller.ts hls-controller.test.mjs; do
  case "$source" in
    controller.ts) target=client/src/neko/hls/controller.ts;;
    hls-controller.test.mjs) target=client/tests/hls-controller.test.mjs;;
  esac
  [[ "$(git hash-object -- "$client_report/$source")" == "$(git rev-parse "$repair_commit:$target")" ]] || fail 'client gate source differs from application repair'
done
case "$output/" in "$repository/"*|"$base_output/"*) fail 'new OUTPUT_DIR must be outside Git and prior evidence';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain Git';; esac
case "$base_output/" in "$output/"*) fail 'OUTPUT_DIR must not contain prior evidence';; esac
[[ ! -e "$4" && ! -L "$4" ]] || fail 'new OUTPUT_DIR required; preserve any prior attempt'
container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$container" =~ ^[0-9a-f]{12,64}$ ]] || fail 'exactly one running Neko container required'
live_image="$(docker inspect -f '{{.Image}}' "$container")"
image_record="$(grep -E '^image=(docker.io/)?my-neko/brave:hls-a7ffb8b13448 id=sha256:[0-9a-f]{64} created=' "$base_output/images.txt")"
[[ "$image_record" != *$'\n'* ]] || fail 'ambiguous baseline image record'
base_image="${image_record#* id=}"
base_image="${base_image%% *}"
[[ "$live_image" == "$base_image" && "$(docker image inspect --format '{{.Id}}' my-neko/brave:hls-a7ffb8b13448)" == "$base_image" ]] || fail 'live/tagged image differs from the prepared baseline record'

# All evidence/source guards run before the checkout moves or image work begins.
git merge --ff-only "$repair_commit"
[[ "$(git rev-parse HEAD)" == "$repair_commit" ]] || fail 'repair checkout mismatch'
bash deploy/collect-hls-media.sh init "$output"
printf 'PENDING\n' >"$output/validation-commit.txt"
printf '%s\n' "$base_commit" >"$output/inherited-backend-commit.txt"
printf '%s\n' "$client_report" >"$output/client-check-evidence.txt"
printf '%s\n' "$helper_commit" >"$output/operator-helper-commit.txt"
git hash-object -- "$helper_source" >"$output/operator-preparer-blob.txt"
printf '%s\n' 'client-only repair: passed isolated A/B, full client tests/type/build; backend tests/codec/fuzz evidence inherited from identical a7ff sources; no fresh backend test or enabled-playback acceptance claim' >"$output/preparation-scope.txt"
exec > >(tee "$output/validation.log") 2>&1
image_tag="hls-${repair_commit:0:12}"
printf 'application_commit=%s\ninherited_backend_commit=%s\nimage_tag=%s\n' "$repair_commit" "$base_commit" "$image_tag"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
# Assemble fresh client artifacts; do not accept an operator's old CLIENT_DIST.
CLIENT_DIST='' bash -o pipefail ./build "my-neko/base:$image_tag" -y </dev/null
CLIENT_DIST='' bash -o pipefail ./build "my-neko/brave:$image_tag" -y </dev/null
docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
  --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
[[ "$(git rev-parse HEAD)" == "$repair_commit" && -z "$(git status --porcelain=v1)" ]] || fail 'application checkout changed during preparation'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during preparation'
bash deploy/collect-hls-media.sh snapshot "$output" client-repair-prepared-live-a7ff
printf '%s\n' "$repair_commit" >"$output/validation-commit.txt"
printf 'CLIENT REPAIR IMAGE GATE PASSED; backend evidence inherited; running service unchanged. Evidence: %s\n' "$output"
