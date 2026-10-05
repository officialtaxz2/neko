#!/usr/bin/env bash

# Target server only: assemble exact-68 client fixes after the isolated client
# gate. Inherit the identical validated backend; leave the live service running.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 5 ]] || fail 'Usage: bash prepare-hls-client-stability.sh REPOSITORY BASE_OUTPUT CLIENT_REPORT OUTPUT_DIR HELPER_COMMIT'
umask 077
for required in docker git realpath stat grep tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
[[ "$PWD" == "$repository" ]] || fail 'REPOSITORY must be the Git root'
base_commit=73d5ff6d29110e3dd06999726a7e88718d09ea23
candidate=68dbdd4a8dd798886302b235c1f8f208452e0c6e
helper_commit="$5"
[[ "$helper_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full helper commit required'
[[ "$(git branch --show-current)" == testing && -z "$(git status --porcelain=v1)" ]] || fail 'clean testing checkout required'
checkout_before="$(git rev-parse HEAD)"
git merge-base --is-ancestor "$base_commit" "$checkout_before" || fail 'checkout precedes the expected live baseline'
git merge-base --is-ancestor "$checkout_before" "$candidate" || fail 'checkout is outside the reviewed candidate history'
git merge-base --is-ancestor "$candidate" "$helper_commit" || fail 'helper does not descend from the reviewed candidate'
[[ "$(git hash-object -- "$helper_source")" == "$(git rev-parse "$helper_commit:deploy/prepare-hls-client-stability.sh")" ]] || fail 'helper differs from selected tooling commit'
git diff --quiet "$base_commit" "$candidate" -- server runtime apps build .dockerignore 'docker-compose*.yaml' || fail 'backend/build/deployment sources differ from the validated baseline'
git diff --quiet "$base_commit" "$candidate" -- client/package.json client/package-lock.json || fail 'client dependencies differ from the reviewed repair scope'

base_output="$(realpath -- "$2")"
client_report="$(realpath -- "$3")"
output="$(realpath -m -- "$4")"
for private_dir in "$base_output" "$client_report"; do
  [[ -d "$private_dir" && ! -L "$private_dir" && "$(stat -c %a -- "$private_dir")" == 700 ]] || fail 'private baseline/client evidence required'
  case "$private_dir/" in "$repository/"*) fail 'evidence must be outside Git';; esac
  case "$repository/" in "$private_dir/"*) fail 'evidence must not contain Git';; esac
done
[[ "$(cat "$base_output/validation-commit.txt")" == "$base_commit" ]] || fail 'exact baseline image preparation record required'
[[ "$(cat "$client_report/candidate-commit.txt")" == "$candidate" && "$(cat "$client_report/client-check-commit.txt")" == "$candidate" ]] || fail 'passed exact candidate client gate required'
[[ -f "$client_report/client-check.log" && -d "$client_report/source" && ! -L "$client_report/source" ]] || fail 'isolated client log/source missing'
# Match the source consumed by client-checks, including the manifests, every
# client asset/test and the shared server fixtures. Read paths with NUL framing.
git ls-tree -rz --name-only "$candidate" -- client docker-compose.validation.yaml server/internal/mediahls/testdata server/internal/mediaws/testdata |
  while IFS= read -r -d '' source; do
    [[ -f "$client_report/source/$source" && ! -L "$client_report/source/$source" ]] || fail 'isolated client gate source is missing or a symlink'
    [[ "$(git hash-object --no-filters -- "$client_report/source/$source")" == "$(git rev-parse "$candidate:$source")" ]] || fail 'isolated client gate source differs from the candidate'
  done
case "$output/" in "$repository/"*|"$base_output/"*|"$client_report/"*) fail 'new OUTPUT_DIR must be outside Git and prior evidence';; esac
for retained_dir in "$repository" "$base_output" "$client_report"; do
  case "$retained_dir/" in "$output/"*) fail 'OUTPUT_DIR must not contain Git or prior evidence';; esac
done
[[ ! -e "$4" && ! -L "$4" ]] || fail 'new OUTPUT_DIR required; preserve any previous attempt'
container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$container" =~ ^[0-9a-f]{12,64}$ ]] || fail 'exactly one running Neko container required'
live_image="$(docker inspect -f '{{.Image}}' "$container")"
image_record="$(grep -E '^image=(docker.io/)?my-neko/brave:hls-73d5ff6d2911 id=sha256:[0-9a-f]{64} created=' "$base_output/images.txt")"
[[ "$image_record" != *$'\n'* ]] || fail 'ambiguous baseline image record'
base_image="${image_record#* id=}"
base_image="${base_image%% *}"
[[ "$live_image" == "$base_image" && "$(docker image inspect --format '{{.Id}}' my-neko/brave:hls-73d5ff6d2911)" == "$base_image" ]] || fail 'live/tagged image differs from the prepared baseline record'

# All guards precede the checkout change. Building creates candidate image tags
# only; this helper has no stop/up/recreate/reload action.
git merge --ff-only "$candidate"
[[ "$(git rev-parse HEAD)" == "$candidate" ]] || fail 'candidate checkout mismatch'
bash deploy/collect-hls-media.sh init "$output"
printf 'PENDING\n' >"$output/validation-commit.txt"
printf '%s\n' "$base_commit" >"$output/inherited-backend-commit.txt"
printf '%s\n' "$base_output" >"$output/inherited-backend-evidence.txt"
printf '%s\n' "$client_report" >"$output/client-check-evidence.txt"
printf '%s\n' "$helper_commit" >"$output/operator-helper-commit.txt"
git hash-object -- "$helper_source" >"$output/operator-preparer-blob.txt"
printf '%s\n' 'client-only candidate: supplied exact-68 isolated 60-test/type/build gate passed; backend/build/deployment sources match exact-73, whose backend evidence is inherited from a7ff; no fresh server/codec/fuzz or live-playback acceptance claim' >"$output/preparation-scope.txt"
exec > >(tee "$output/validation.log") 2>&1
image_tag="hls-${candidate:0:12}"
printf 'application_commit=%s\ninherited_backend_commit=%s\nimage_tag=%s\n' "$candidate" "$base_commit" "$image_tag"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml \
  -f docker-compose.webcodecs-ws.yaml -f docker-compose.hls.yaml config --quiet
CLIENT_DIST='' bash -o pipefail ./build "my-neko/base:$image_tag" -y </dev/null
CLIENT_DIST='' bash -o pipefail ./build "my-neko/brave:$image_tag" -y </dev/null
docker image inspect "my-neko/base:$image_tag" "my-neko/brave:$image_tag" \
  --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}' >"$output/images.txt"
[[ "$(git rev-parse HEAD)" == "$candidate" && -z "$(git status --porcelain=v1)" ]] || fail 'candidate checkout changed during image preparation'
[[ "$(docker compose -f docker-compose.yaml ps -q neko)" == "$container" && "$(docker inspect -f '{{.Image}}' "$container")" == "$live_image" ]] || fail 'live container changed during image preparation'
bash deploy/collect-hls-media.sh snapshot "$output" client-stability-prepared-live-73
printf '%s\n' "$candidate" >"$output/validation-commit.txt"
printf 'CLIENT STABILITY IMAGE GATE PASSED; backend evidence inherited; running service unchanged. Evidence: %s\n' "$output"
