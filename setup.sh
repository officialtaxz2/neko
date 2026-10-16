#!/usr/bin/env bash
#
# Install or update the local deployment of this fork. The browser image is
# chosen with NEKO_BROWSER in .env: brave (default) or helium.
#
# First installation:   git clone ... && cd neko && bash setup.sh
# Every update:         git pull && bash setup.sh
#
# The script is idempotent: it never overwrites existing values in .env,
# only adds keys that are new in .env.example and generates passwords that
# are still empty. One-time migrations of old defaults are logged and marked
# in .env, so a later deliberate change is kept. Images are rebuilt with the Docker cache, so an unchanged
# checkout rebuilds quickly and Compose only recreates what changed.
#
set -euo pipefail
cd "$(dirname "$0")"

NEKO_UID=1000 # uid/gid of the "neko" user inside the image (runtime/Dockerfile)
HEALTH_TIMEOUT=180

DO_BUILD=1
DO_CHECK=0
NO_CACHE=""

usage() {
  cat <<EOF
Usage: bash setup.sh [options]

Options:
  --no-build    Skip the image build, only prepare .env/directories and (re)start.
  --no-cache    Build without Docker cache (also pulls the latest browser release).
  --check       Run the server and client tests in containers first
                (docker-compose.validation.yaml, a few minutes); stop if they fail.
                Useful after a git pull that changed server/ or client/.
  -h, --help    Show this help.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build) DO_BUILD=0 ;;
    --no-cache) NO_CACHE="--no-cache" ;;
    --check) DO_CHECK=1 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
  shift
done

log()  { printf '\033[1;34m[setup]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[setup] WARN:\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31m[setup] ERROR:\033[0m %s\n' "$*" >&2; exit 1; }

# --------------------------------------------------------------------
# Prerequisites

command -v docker >/dev/null 2>&1 || die "docker is not installed (https://docs.docker.com/engine/install/)."
docker info >/dev/null 2>&1 || die "Cannot talk to the Docker daemon. Start it, run with sudo, or add your user to the 'docker' group."
docker compose version >/dev/null 2>&1 || die "Docker Compose v2 ('docker compose') is required (package docker-compose-plugin)."

[[ -f docker-compose.yaml ]] || die "docker-compose.yaml not found; run this script from the repository checkout."
[[ -f .env.example ]] || die ".env.example not found."

if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
  log "Checkout: $(git rev-parse --abbrev-ref HEAD) @ $(git rev-parse --short HEAD)"
  if [[ -n "$(git status --porcelain --untracked-files=no)" ]]; then
    warn "The checkout has uncommitted changes to tracked files; they are built as they are."
  fi
fi

# --------------------------------------------------------------------
# .env

# Value of KEY in FILE (last assignment wins, surrounding quotes stripped).
env_get() {
  local key="$1" file="$2" line
  line="$(grep -E "^${key}=" "$file" | tail -n 1 || true)"
  line="${line#*=}"
  line="${line%\"}"; line="${line#\"}"
  line="${line%\'}"; line="${line#\'}"
  printf '%s' "$line"
}

env_has() { grep -qE "^$1=" "$2"; }

# Replace the (last) assignment of KEY in .env, keeping file permissions.
env_set() {
  local key="$1" value="$2" tmp
  tmp="$(mktemp)"
  awk -v k="$key" -v v="$value" '
    { lines[NR] = $0; if (index($0, k "=") == 1) last = NR }
    END { for (i = 1; i <= NR; i++) print (i == last ? k "=" v : lines[i]) }
  ' .env > "$tmp"
  cat "$tmp" > .env
  rm -f "$tmp"
}

gen_password() {
  local pw=""
  while [[ ${#pw} -lt 24 ]]; do
    pw+="$(head -c 48 /dev/urandom | base64 | tr -dc 'A-Za-z0-9')"
  done
  printf '%s' "${pw:0:24}"
}

if [[ ! -f .env ]]; then
  log "Creating .env from .env.example"
  (umask 077; cp .env.example .env)
else
  # Add keys that were introduced in .env.example after the first install.
  added=()
  while IFS= read -r line; do
    key="${line%%=*}"
    if ! env_has "$key" .env; then
      [[ -n "$(tail -c 1 .env)" ]] && echo >> .env
      echo "$line" >> .env
      added+=("$key")
    fi
  done < <(grep -E '^[A-Za-z_][A-Za-z0-9_]*=' .env.example)
  if [[ ${#added[@]} -gt 0 ]]; then
    log "Added new settings to .env: ${added[*]}"
  fi
fi
chmod 600 .env

# WebCodecs was removed. Its origin/proxy settings used to feed HLS as well,
# so carry them over to the HLS keys and drop the deleted overlay from
# COMPOSE_FILE; Compose would otherwise fail on the missing file.
for pair in ALLOWED_ORIGINS TRUSTED_PROXIES; do
  legacy="$(env_get "NEKO_MEDIA_WEBCODECS_WS_$pair" .env)"
  if [[ -n "$legacy" && -z "$(env_get "NEKO_MEDIA_HLS_$pair" .env)" ]]; then
    env_set "NEKO_MEDIA_HLS_$pair" "$legacy"
    log "Moved NEKO_MEDIA_WEBCODECS_WS_$pair to NEKO_MEDIA_HLS_$pair in .env"
  fi
done
compose_value="$(env_get COMPOSE_FILE .env)"
if [[ ":$compose_value:" == *":docker-compose.webcodecs-ws.yaml:"* ]]; then
  compose_value="$(tr ':' '\n' <<< "$compose_value" | grep -vx 'docker-compose.webcodecs-ws.yaml' | paste -sd: -)"
  env_set COMPOSE_FILE "${compose_value:-docker-compose.yaml}"
  log "Removed the deleted docker-compose.webcodecs-ws.yaml from COMPOSE_FILE in .env"
fi
# Delete the now unused NEKO_MEDIA_WEBCODECS_WS_* lines (after the move above),
# plus the comment block directly above them when it mentions WebCodecs.
if grep -qE '^NEKO_MEDIA_WEBCODECS_WS_' .env; then
  removed="$(grep -E '^NEKO_MEDIA_WEBCODECS_WS_' .env | cut -d= -f1 | paste -sd' ' -)"
  tmp="$(mktemp)"
  awk '
    { line[NR] = $0 }
    END {
      for (i = 1; i <= NR; i++) {
        if (line[i] !~ /^NEKO_MEDIA_WEBCODECS_WS_/) continue
        drop[i] = 1
        # Comment block right above this run of keys.
        if (i > 1 && line[i - 1] ~ /^NEKO_MEDIA_WEBCODECS_WS_/) continue
        j = i - 1; mentions = 0
        while (j >= 1 && line[j] ~ /^#/) { if (tolower(line[j]) ~ /webcodecs/) mentions = 1; j-- }
        if (mentions) for (k = j + 1; k < i; k++) drop[k] = 1
      }
      # Avoid leaving two blank lines where a whole block was removed.
      for (i = 1; i <= NR; i++) {
        if (drop[i] || line[i] != "") continue
        p = i - 1; while (p >= 1 && drop[p]) p--
        if (p < i - 1 && (p < 1 || line[p] == "")) drop[i] = 1
      }
      for (i = 1; i <= NR; i++) if (!drop[i]) print line[i]
    }
  ' .env > "$tmp"
  cat "$tmp" > .env
  rm -f "$tmp"
  log "Removed the unused WebCodecs settings from .env: $removed"
fi

# LL-HLS became a default HLS mode next to conventional HLS. Upgrade the old
# default 'hls' once; the marker keeps a later deliberate 'hls' untouched.
ll_marker='# setup.sh: LL-HLS default applied'
if ! grep -qxF "$ll_marker" .env; then
  modes="$(env_get NEKO_MEDIA_HLS_MODES .env)"
  if [[ -z "$modes" || "$modes" == hls ]]; then
    env_set NEKO_MEDIA_HLS_MODES "hls ll-hls"
    log "Set NEKO_MEDIA_HLS_MODES to 'hls ll-hls' in .env (LL-HLS is now offered by default)"
  fi
  [[ -n "$(tail -c 1 .env)" ]] && echo >> .env
  echo "$ll_marker" >> .env
fi

# NEKO_IMAGE and NEKO_POLICY_FILE used to name the Brave files explicitly.
# Clear those old defaults once, so both follow NEKO_BROWSER; Brave resolves
# to the same image and policy file as before. The marker keeps a later
# deliberate value untouched.
browser_marker='# setup.sh: browser defaults follow NEKO_BROWSER'
if ! grep -qxF "$browser_marker" .env; then
  if [[ "$(env_get NEKO_IMAGE .env)" == my-neko/brave:latest ]]; then
    env_set NEKO_IMAGE ""
    log "Cleared NEKO_IMAGE in .env; it now defaults to my-neko/<NEKO_BROWSER>:latest"
  fi
  if [[ "$(env_get NEKO_POLICY_FILE .env)" == ./apps/brave/policies.json ]]; then
    env_set NEKO_POLICY_FILE ""
    log "Cleared NEKO_POLICY_FILE in .env; it now defaults to ./apps/<NEKO_BROWSER>/policies.json"
  fi
  [[ -n "$(tail -c 1 .env)" ]] && echo >> .env
  echo "$browser_marker" >> .env
fi

generated=()
for key in NEKO_MEMBER_MULTIUSER_USER_PASSWORD NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD; do
  if [[ -z "$(env_get "$key" .env)" ]]; then
    env_set "$key" "$(gen_password)"
    generated+=("$key")
  fi
done

if [[ "$(env_get NEKO_MEMBER_MULTIUSER_USER_PASSWORD .env)" == "$(env_get NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD .env)" ]]; then
  warn "User and admin password in .env are identical; everyone who logs in becomes admin."
fi

# Compose prefers NEKO_BROWSER from the environment over .env; so does this.
BROWSER="${NEKO_BROWSER:-$(env_get NEKO_BROWSER .env)}"; BROWSER="${BROWSER:-brave}"
[[ "$BROWSER" =~ ^[a-z][a-z0-9-]*$ && -f "apps/$BROWSER/Dockerfile" ]] \
  || die "NEKO_BROWSER '$BROWSER' has no apps/$BROWSER/Dockerfile; use brave or helium."
NEKO_IMAGE="$(env_get NEKO_IMAGE .env)";               NEKO_IMAGE="${NEKO_IMAGE:-my-neko/$BROWSER:latest}"
PROFILE_DIR="$(env_get NEKO_PROFILE_DIR .env)";        PROFILE_DIR="${PROFILE_DIR:-./files}"
DOWNLOADS_DIR="$(env_get NEKO_DOWNLOADS_DIR .env)";    DOWNLOADS_DIR="${DOWNLOADS_DIR:-./downloads}"
POLICY_FILE="$(env_get NEKO_POLICY_FILE .env)";        POLICY_FILE="${POLICY_FILE:-./apps/$BROWSER/policies.json}"

# The container mounts the profile and policy paths of NEKO_BROWSER, so an
# image or policy file of another browser would start with the wrong setup.
image_app="${NEKO_IMAGE##*/}"; image_app="${image_app%%:*}"
[[ "$image_app" == "$BROWSER" ]] \
  || die "NEKO_IMAGE '$NEKO_IMAGE' does not match NEKO_BROWSER '$BROWSER'; clear NEKO_IMAGE in .env or name a my-neko/$BROWSER:<tag> image."
if [[ "$POLICY_FILE" =~ (^|/)apps/([^/]+)/policies\.json$ && "${BASH_REMATCH[2]}" != "$BROWSER" ]]; then
  die "NEKO_POLICY_FILE '$POLICY_FILE' belongs to ${BASH_REMATCH[2]}, not NEKO_BROWSER '$BROWSER'; clear it in .env."
fi
log "Browser: $BROWSER ($NEKO_IMAGE)"
HTTP_BIND="$(env_get NEKO_HTTP_BIND .env)";            HTTP_BIND="${HTTP_BIND:-127.0.0.1}"
HTTP_PORT="$(env_get NEKO_HTTP_PORT .env)";            HTTP_PORT="${HTTP_PORT:-8082}"
WEBRTC_EPR="$(env_get NEKO_WEBRTC_EPR .env)";          WEBRTC_EPR="${WEBRTC_EPR:-51000-51100}"

# --------------------------------------------------------------------
# Compose files

# Compose reads COMPOSE_FILE from .env itself; it selects the overlays
# (docker-compose.adaptive.yaml, .hls.yaml). .env.example enables both;
# without COMPOSE_FILE only docker-compose.yaml is used.
COMPOSE_FILES="$(env_get COMPOSE_FILE .env)"
COMPOSE_FILES="${COMPOSE_FILE:-$COMPOSE_FILES}"
selected=" "
IFS=':' read -r -a selected_files <<< "${COMPOSE_FILES:-docker-compose.yaml}"
for f in "${selected_files[@]}"; do selected+="$(basename "$f") "; done
log "Compose files:${selected% }"

has_file() { [[ "$selected" == *" $1 "* ]]; }

if has_file docker-compose.webcodecs-ws.yaml; then
  die "COMPOSE_FILE still names docker-compose.webcodecs-ws.yaml, but WebCodecs was removed.
    Remove it from the COMPOSE_FILE environment variable."
fi

# Value from the environment first (Compose does the same), then from .env.
setting() { local key="$1"; printf '%s' "${!key:-$(env_get "$key" .env)}"; }

# Gateway of this project's Docker network: what a reverse proxy on the same
# host looks like from inside the container.
network_gateway() {
  local name
  name="$(docker compose -f docker-compose.yaml config 2>/dev/null | sed -n 's/^name: *//p' | head -n 1)"
  [[ -n "$name" ]] || return 0
  docker network inspect -f '{{range .IPAM.Config}}{{.Gateway}} {{end}}' "${name}_default" 2>/dev/null \
    | awk '{print $1}' || true
}

if has_file docker-compose.hls.yaml && ! has_file docker-compose.adaptive.yaml; then
  die "docker-compose.hls.yaml needs docker-compose.adaptive.yaml in COMPOSE_FILE as well."
fi

if has_file docker-compose.hls.yaml; then
  hls_origin="$(setting NEKO_MEDIA_HLS_ALLOWED_ORIGINS)"
  hls_proxies="$(setting NEKO_MEDIA_HLS_TRUSTED_PROXIES)"
  missing=()
  [[ -n "$hls_origin" ]] || missing+=(NEKO_MEDIA_HLS_ALLOWED_ORIGINS)
  [[ -n "$hls_proxies" ]] || missing+=(NEKO_MEDIA_HLS_TRUSTED_PROXIES)
  if [[ ${#missing[@]} -gt 0 ]]; then
    gateway="$(network_gateway)"
    hint=""
    [[ -n "$gateway" ]] && hint="($gateway is the gateway of the Docker network of this setup, right for a proxy on the same host.)"
    die "HLS is enabled through COMPOSE_FILE, but .env lacks ${missing[*]}.
    Set the public https address of this Neko, e.g.
      NEKO_MEDIA_HLS_ALLOWED_ORIGINS=https://neko.example
    and the reverse proxy address as Neko sees it, e.g.
      NEKO_MEDIA_HLS_TRUSTED_PROXIES=${gateway:-172.18.0.1}/32
    $hint
    To run without HLS instead, set
      COMPOSE_FILE=docker-compose.yaml:docker-compose.adaptive.yaml"
  fi
  if [[ "$hls_origin" != https://* ]]; then
    die "HLS requires an https origin, but the configured one is '$hls_origin'."
  fi
fi

# A running neko that was started with overlays must not silently lose them.
running_cid="$(docker compose ps -q neko 2>/dev/null | head -n 1 || true)"
if [[ -n "$running_cid" ]]; then
  running_files="$(docker inspect -f '{{index .Config.Labels "com.docker.compose.project.config_files"}}' "$running_cid")"
  missing=()
  IFS=',' read -r -a running_list <<< "$running_files"
  for f in "${running_list[@]}"; do
    name="$(basename "$f")"
    # The removed WebCodecs overlay is dropped on purpose.
    [[ "$name" == docker-compose.webcodecs-ws.yaml ]] && continue
    [[ "$selected" == *" $name "* ]] || missing+=("$name")
  done
  if [[ ${#missing[@]} -gt 0 ]]; then
    running_value=""
    for f in "${running_list[@]}"; do
      [[ "$(basename "$f")" == docker-compose.webcodecs-ws.yaml ]] && continue
      running_value+="${running_value:+:}$(basename "$f")"
    done
    die "neko is running with ${missing[*]}, which this run would drop. Keep them by adding
    COMPOSE_FILE=$running_value
to .env (or remove them from the running setup on purpose with 'docker compose down' first)."
  fi
fi

# --------------------------------------------------------------------
# Host directories and files

# Without this check Docker would silently create a directory at that path.
[[ -f "$POLICY_FILE" ]] || die "Policy file '$POLICY_FILE' (NEKO_POLICY_FILE) does not exist."

for dir in "$PROFILE_DIR" "$DOWNLOADS_DIR"; do
  [[ -d "$dir" ]] || { log "Creating $dir"; mkdir -p "$dir"; }
  owner="$(stat -c '%u:%g' "$dir")"
  if [[ "$owner" != "$NEKO_UID:$NEKO_UID" ]]; then
    # Done through a container so it works without sudo for docker-group users.
    log "Setting owner of $dir to $NEKO_UID:$NEKO_UID (was $owner)"
    docker run --rm -v "$(cd "$dir" && pwd):/target" alpine:latest \
      chown -R "$NEKO_UID:$NEKO_UID" /target
  fi
done

# --------------------------------------------------------------------
# Optional checks (--check)

if [[ "$DO_CHECK" == 1 ]]; then
  [[ -f docker-compose.validation.yaml ]] || die "docker-compose.validation.yaml not found."
  validation=(docker compose -f docker-compose.validation.yaml)
  commit="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"

  log "Checks: building the server test image"
  NEKO_VALIDATION_COMMIT="$commit" "${validation[@]}" build server-checks \
    || die "Building the server test image failed; the running containers were not touched."
  log "Checks: Go tests, fuzzing and server build"
  "${validation[@]}" run --rm server-checks \
    || die "Server checks failed; the running containers were not touched. See the output above."
  log "Checks: client tests, lint and build"
  "${validation[@]}" run --rm client-checks \
    || die "Client checks failed; the running containers were not touched. See the output above."
  log "Checks passed"
fi

# --------------------------------------------------------------------
# Images

if [[ "$DO_BUILD" == 1 ]]; then
  [[ "$NEKO_IMAGE" == */*:* ]] || die "NEKO_IMAGE '$NEKO_IMAGE' must look like <repository>/$BROWSER:<tag> to be built here."
  repository="${NEKO_IMAGE%/*}"
  tag="${NEKO_IMAGE##*:}"
  build_args=(-y)
  [[ -n "$NO_CACHE" ]] && build_args+=("$NO_CACHE")

  log "Building base image $repository/base:$tag"
  bash ./build "$repository/base:$tag" "${build_args[@]}"
  log "Building browser image $NEKO_IMAGE"
  bash ./build "$NEKO_IMAGE" "${build_args[@]}"
else
  docker image inspect "$NEKO_IMAGE" >/dev/null 2>&1 \
    || die "Image $NEKO_IMAGE does not exist yet; run without --no-build."
fi

# --------------------------------------------------------------------
# Start

log "Validating Compose configuration"
docker compose config --quiet

log "Starting containers"
docker compose up -d --remove-orphans

cid="$(docker compose ps -q neko)"
[[ -n "$cid" ]] || die "The neko container was not created."

log "Waiting for neko to become healthy (up to ${HEALTH_TIMEOUT}s)"
status=""
for ((i = 0; i < HEALTH_TIMEOUT; i += 5)); do
  status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$cid")"
  [[ "$status" == healthy || "$status" == unhealthy || "$status" == exited ]] && break
  sleep 5
done

if [[ "$status" != healthy ]]; then
  docker compose logs --tail 80 neko >&2 || true
  die "neko is '$status' after ${HEALTH_TIMEOUT}s; see the log above."
fi

# --------------------------------------------------------------------
# Summary

echo
log "neko is running: http://$HTTP_BIND:$HTTP_PORT"
if [[ "$HTTP_BIND" == 127.0.0.1 || "$HTTP_BIND" == localhost ]]; then
  log "HTTP is bound to localhost only; reach it through your reverse proxy or 'ssh -L $HTTP_PORT:127.0.0.1:$HTTP_PORT'."
fi
enabled=()
has_file docker-compose.adaptive.yaml && enabled+=("adaptive quality")
has_file docker-compose.hls.yaml && enabled+=("HLS")
if [[ ${#enabled[@]} -gt 0 ]]; then
  log "Enabled features: ${enabled[*]} (WebRTC stays the default media path)"
fi
if [[ -n "${hls_proxies:-}" ]]; then
  gateway="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.Gateway}} {{end}}' "$cid" | awk '{print $1}')"
  configured="$hls_proxies"
  if [[ -n "$gateway" && "$configured" != "$gateway" && "$configured" != "$gateway/32" ]]; then
    warn "Trusted proxy is '$configured', but the Docker network gateway is $gateway. If your"
    warn "reverse proxy runs on this host, HLS will reject it; set $gateway/32 in .env."
  fi
fi
log "WebRTC needs UDP $WEBRTC_EPR open in the firewall. If video stays black for remote"
log "clients, set NEKO_WEBRTC_NAT1TO1 in .env and enable it in docker-compose.yaml."
if [[ ${#generated[@]} -gt 0 ]]; then
  echo
  log "Generated passwords (stored in .env):"
  for key in "${generated[@]}"; do
    printf '    %s=%s\n' "$key" "$(env_get "$key" .env)"
  done
fi
