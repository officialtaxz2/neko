# Sourced by the deploy helpers after they cd into the repository.
# Resolves the browser application image the same way setup.sh and
# docker-compose.yaml do: NEKO_BROWSER from the environment, then from .env,
# defaulting to brave. Prints the application name (apps/<name>).
neko_browser() {
  local value="${NEKO_BROWSER:-}"
  if [[ -z "$value" && -f .env ]]; then
    value="$(grep -E '^NEKO_BROWSER=' .env | tail -n 1 || true)"
    value="${value#*=}"; value="${value//\"/}"; value="${value//\'/}"
  fi
  value="${value:-brave}"
  [[ "$value" =~ ^[a-z][a-z0-9-]*$ && -f "apps/$value/Dockerfile" ]] || {
    printf 'error: NEKO_BROWSER=%s has no apps/<name>/Dockerfile\n' "$value" >&2
    return 1
  }
  printf '%s' "$value"
}
