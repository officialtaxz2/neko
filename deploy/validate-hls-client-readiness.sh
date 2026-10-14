#!/usr/bin/env bash

# Target server only: reproduce the old client deadline defect, then check the
# repaired client. Checkout, images, Caddy and the running service are unchanged.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 3 ]] || fail 'Usage: bash validate-hls-client-readiness.sh REPOSITORY OUTPUT_DIR REPAIR_COMMIT'
umask 077
for required in docker git realpath stat date tee; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
[[ "$PWD" == "$repository" ]] || fail 'REPOSITORY must be the Git root'
base_commit=a7ffb8b13448a8329c6df24fdcb182ac32ca398c
repair_commit="$3"
[[ "$repair_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'full repair commit required'
[[ "$(git branch --show-current)" == testing ]] || fail 'testing checkout required'
[[ "$(git rev-parse HEAD)" == "$base_commit" ]] || fail 'expected exact running application checkout'
[[ -z "$(git status --porcelain=v1)" ]] || fail 'tracked checkout must be clean'
git cat-file -e "$repair_commit^{commit}"
git merge-base --is-ancestor "$base_commit" "$repair_commit"
[[ "$(git hash-object -- "$helper_source")" == "$(git rev-parse "$repair_commit:deploy/validate-hls-client-readiness.sh")" ]] || fail 'helper differs from repair commit'
changed_files="$(git diff --name-only "$base_commit" "$repair_commit")"
while IFS= read -r file; do
  case "$file" in
    client/src/neko/hls/controller.ts|client/tests/hls-controller.test.mjs|deploy/validate-hls-client-readiness.sh|AGENTS.md|README.md|docs/*) ;;
    *) fail 'repair changed files outside the bounded client/readiness block' ;;
  esac
done <<< "$changed_files"
output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ -f "$output/validation-commit.txt" && "$(cat "$output/validation-commit.txt")" == "$base_commit" ]] || fail 'exact baseline preparation marker required'
report="$output/client-readiness-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
printf '%s\n' "$base_commit" >"$report/base-commit.txt"
printf '%s\n' "$repair_commit" >"$report/repair-commit.txt"
git hash-object -- "$helper_source" >"$report/helper-blob.txt"
git show "$repair_commit:client/src/neko/hls/controller.ts" >"$report/controller.ts"
git show "$repair_commit:client/tests/hls-controller.test.mjs" >"$report/hls-controller.test.mjs"
printf 'PENDING\n' >"$report/client-check-commit.txt"
cat >"$report/run-client-check.sh" <<'SH'
#!/bin/sh
set -eu
cp -a /source/client /work/client
mkdir -p /work/server/internal/mediaws/testdata /work/server/internal/mediahls/testdata
cp /source/server/internal/mediaws/testdata/neko_media_v1_golden.json /work/server/internal/mediaws/testdata/
cp /source/server/internal/mediahls/testdata/*.m3u8 /work/server/internal/mediahls/testdata/
cd /work/client
cp /repair/hls-controller.test.mjs tests/hls-controller.test.mjs
npm ci
# New regression against the unchanged old controller. Only the specific
# assertion and exactly one failed test qualify as reproduction, not any error.
old_status=0
node --test --test-reporter=tap \
  --test-name-pattern='^initial HLS readiness cannot later turn buffering into a startup timeout$' \
  tests/hls-controller.test.mjs >/work/old-readiness.tap 2>&1 || old_status=$?
if [ "$old_status" -ne 1 ] ||
   ! grep -Fq "error: 'ready HLS playback was terminated by the startup deadline'" /work/old-readiness.tap ||
   ! grep -Fqx '# fail 1' /work/old-readiness.tap; then
  cat /work/old-readiness.tap
  printf 'error: expected old readiness-deadline assertion was not reproduced\n' >&2
  exit 1
fi
printf 'PASS old client: specific post-readiness startup-deadline defect reproduced\n'
cp /repair/controller.ts src/neko/hls/controller.ts
npm test
npm run lint
npm run build
SH
exec > >(tee "$report/client-check.log") 2>&1
printf 'base_commit=%s\nrepair_commit=%s\n' "$base_commit" "$repair_commit"
docker compose -f docker-compose.validation.yaml run --rm -T \
  -v "$report:/repair:ro" client-checks sh /repair/run-client-check.sh </dev/null
[[ "$(git rev-parse HEAD)" == "$base_commit" && -z "$(git status --porcelain=v1)" ]] || fail 'checkout changed during isolated client checks'
printf '%s\n' "$repair_commit" >"$report/client-check-commit.txt"
printf 'CLIENT READINESS A/B/TYPE/BUILD GATE PASSED; running service unchanged. Evidence: %s\n' "$report"
