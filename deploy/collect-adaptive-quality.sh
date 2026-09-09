#!/usr/bin/env bash

set -Eeuo pipefail

readonly COMPOSE_BASE="docker-compose.yaml"
readonly COMPOSE_ADAPTIVE="docker-compose.adaptive.yaml"
readonly RESULT_TEMPLATE="docs/ADAPTIVE_QUALITY_RESULTS_TEMPLATE.md"
readonly DEFAULT_METRICS_URL="http://127.0.0.1:8082/metrics"

usage() {
  cat <<'EOF'
Collect reproducible evidence for the opt-in adaptive-quality profile.

Run from the repository root on the real target server:

  ./deploy/collect-adaptive-quality.sh init OUTPUT_DIR
  ./deploy/collect-adaptive-quality.sh snapshot OUTPUT_DIR PHASE
  ./deploy/collect-adaptive-quality.sh logs OUTPUT_DIR [SINCE]

Examples:

  ./deploy/collect-adaptive-quality.sh init ../neko-adaptive-results
  ./deploy/collect-adaptive-quality.sh snapshot ../neko-adaptive-results baseline-start
  ./deploy/collect-adaptive-quality.sh logs ../neko-adaptive-results 60m

Set NEKO_METRICS_URL when metrics are not reachable at
http://127.0.0.1:8082/metrics. The script never reads or prints .env.
EOF
}

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

require_repository_root() {
  [[ -f "$COMPOSE_BASE" ]] || fail "run this script from the repository root"
  [[ -f "$COMPOSE_ADAPTIVE" ]] || fail "missing $COMPOSE_ADAPTIVE"
  [[ -f "$RESULT_TEMPLATE" ]] || fail "missing $RESULT_TEMPLATE"
}

prepare_output_dir() {
  local output_dir="$1"

  [[ -n "$output_dir" ]] || fail "OUTPUT_DIR must not be empty"
  mkdir -p -- "$output_dir"
  chmod 700 -- "$output_dir"
}

safe_label() {
  local label="$1"

  label="${label//[^[:alnum:]._-]/_}"
  [[ -n "$label" ]] || fail "PHASE must contain at least one usable character"
  printf '%s' "$label"
}

compose() {
  docker compose -f "$COMPOSE_BASE" -f "$COMPOSE_ADAPTIVE" "$@"
}

write_environment_record() {
  local output_file="$1"
  local container_id

  container_id="$(compose ps -q neko)"

  {
    printf 'collected_at_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    printf 'commit=%s\n' "$(git rev-parse HEAD)"
    printf 'branch=%s\n' "$(git branch --show-current)"
    if [[ -n "$(git status --porcelain=v1)" ]]; then
      printf 'worktree_clean=false\n'
      printf '%s\n' 'worktree_status_begin'
      git status --short
      printf '%s\n' 'worktree_status_end'
    else
      printf 'worktree_clean=true\n'
    fi
    printf 'host_kernel=%s\n' "$(uname -srmo)"
    printf 'host_cpu_count=%s\n' "$(getconf _NPROCESSORS_ONLN 2>/dev/null || printf 'unknown')"
    if [[ -n "$container_id" ]]; then
      docker inspect "$container_id" --format 'container_id={{.Id}}'
      docker inspect "$container_id" --format 'configured_image={{.Config.Image}}'
      docker inspect "$container_id" --format 'image_id={{.Image}}'
      docker inspect "$container_id" --format 'container_status={{.State.Status}}'
      docker inspect "$container_id" --format 'container_started_at={{.State.StartedAt}}'
    else
      printf 'container_id=not-running\n'
    fi
    printf '%s\n' 'compose_ps_begin'
    compose ps
    printf '%s\n' 'compose_ps_end'
  } >"$output_file"
}

write_resource_snapshot() {
  local output_file="$1"
  local container_id

  container_id="$(compose ps -q neko)"

  {
    printf 'collected_at_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    printf '%s\n' 'uptime_begin'
    uptime
    printf '%s\n' 'uptime_end'
    if command -v free >/dev/null 2>&1; then
      printf '%s\n' 'memory_begin'
      free -h
      printf '%s\n' 'memory_end'
    fi
    if [[ -n "$container_id" ]]; then
      printf '%s\n' 'docker_stats_begin'
      docker stats --no-stream \
        --format 'container={{.Container}} name={{.Name}} cpu={{.CPUPerc}} memory={{.MemUsage}} memory_percent={{.MemPerc}} network={{.NetIO}} block={{.BlockIO}} pids={{.PIDs}}' \
        "$container_id"
      printf '%s\n' 'docker_stats_end'
    else
      printf 'docker_stats=neko-container-not-running\n'
    fi
  } >"$output_file"
}

write_metrics_snapshot() {
  local output_file="$1"
  local metrics_url="${NEKO_METRICS_URL:-$DEFAULT_METRICS_URL}"

  curl --fail --silent --show-error --max-time 15 "$metrics_url" \
    | grep -E '^neko_(capture_(streamsink_(bitrate|listeners|bytes_total)|pipelines_active)|webrtc_(connection_state|receiver_estimated_target_bitrate|track_dropped_samples_total|video_listeners))' \
    >"$output_file"
}

init_record() {
  local output_dir="$1"
  local stamp

  prepare_output_dir "$output_dir"
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  if [[ ! -e "$output_dir/RESULTS.md" ]]; then
    cp -- "$RESULT_TEMPLATE" "$output_dir/RESULTS.md"
  fi
  write_environment_record "$output_dir/environment-$stamp.txt"
  printf 'Initialized evidence directory: %s\n' "$output_dir"
}

snapshot_record() {
  local output_dir="$1"
  local phase="$2"
  local label
  local stamp

  prepare_output_dir "$output_dir"
  label="$(safe_label "$phase")"
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  write_metrics_snapshot "$output_dir/$stamp-$label.metrics.prom"
  write_resource_snapshot "$output_dir/$stamp-$label.resources.txt"
  printf 'Recorded snapshot: %s (%s)\n' "$phase" "$stamp"
}

logs_record() {
  local output_dir="$1"
  local since="${2:-30m}"
  local stamp
  local output_file

  prepare_output_dir "$output_dir"
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  output_file="$output_dir/$stamp-estimator.log"
  compose logs --no-color --since "$since" neko \
    | grep -E 'got bitrate from estimator|downgraded video stream|upgraded video stream|set video|dropping sample' \
    >"$output_file" || true
  printf 'Recorded filtered logs: %s\n' "$output_file"
}

main() {
  local action="${1:-}"

  umask 077
  require_repository_root
  require_command docker
  require_command git
  require_command curl
  require_command grep

  case "$action" in
    init)
      [[ $# -eq 2 ]] || { usage >&2; exit 2; }
      init_record "$2"
      ;;
    snapshot)
      [[ $# -eq 3 ]] || { usage >&2; exit 2; }
      snapshot_record "$2" "$3"
      ;;
    logs)
      [[ $# -ge 2 && $# -le 3 ]] || { usage >&2; exit 2; }
      logs_record "$2" "${3:-30m}"
      ;;
    -h|--help|help)
      usage
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
}

main "$@"
