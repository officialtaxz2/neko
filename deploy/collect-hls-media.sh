#!/usr/bin/env bash

set -Eeuo pipefail

readonly RESULT_TEMPLATE="docs/HLS_LL_HLS_RESULTS_TEMPLATE.md"
readonly METRICS_PATTERN='^(go_goroutines|go_memstats_heap_alloc_bytes|process_(cpu_seconds_total|resident_memory_bytes)|neko_(media_hls_|media_(deliveries|delivery_(opens|closes)_total|source_generation|subscriptions_active|subscription_(delivered_(bytes|units)_total|discontinuities_total|dropped_units_total|queue_(capacity|depth)))|websocket_(connections|handshakes_total|records_total|bytes_total|drops_total|resyncs_total|write_duration_seconds|queue_depth|queue_bytes|client_lag_milliseconds)|capture_(pipelines_active|streamsink_(bitrate|listeners|bytes))|webrtc_(connection_state|track_dropped_samples_total|receiver_estimated_target_bitrate|receiver_congestion_evidence|recovery_probe_)))'

usage() {
  cat <<'EOF'
Target-server HLS evidence (run from the repository root):
  bash deploy/collect-hls-media.sh init OUTPUT_DIR
  bash deploy/collect-hls-media.sh snapshot OUTPUT_DIR PHASE
  bash deploy/collect-hls-media.sh sample OUTPUT_DIR PHASE COUNT

OUTPUT_DIR must be outside the repository. Set NEKO_METRICS_URL for a changed
loopback port/prefix (default http://127.0.0.1:8082/metrics).
Only fixed metric families and selected container fields are archived.
No .env, full docker inspect, response headers, cookies or raw logs are copied.
The same collector covers WebRTC, WebCodecs and HLS without enabling overlays.
sample takes 1..60 snapshots, ten seconds apart (up to roughly ten minutes).
COUNT is a count, not a duration; slow collection extends elapsed wall time.
EOF
}

fail() { printf 'error: %s\n' "$*" >&2; exit 1; }

prepare_output() {
  local repository_root output_root
  repository_root="$(realpath -- "$(git rev-parse --show-toplevel)")"
  output_root="$(realpath -m -- "$1")"
  case "$output_root/" in "$repository_root/"*) fail 'OUTPUT_DIR must be outside the repository';; esac
  case "$repository_root/" in "$output_root/"*) fail 'OUTPUT_DIR must not be a parent of the repository';; esac
  if [[ -e "$output_root" ]]; then
    [[ -d "$output_root" && "$(stat -c %a -- "$output_root")" == 700 ]] || fail 'use a new directory or an existing private 0700 directory'
  fi
  mkdir -p -m 700 -- "$output_root"
  printf '%s' "$output_root"
}

container_id() {
  # Service discovery needs only the base file; it does not enable any backend.
  docker compose -f docker-compose.yaml ps -q neko
}

environment_record() {
  local container="$1"
  printf 'collected_at_utc=%s\ncommit=%s\nbranch=%s\n' \
    "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$(git rev-parse HEAD)" "$(git branch --show-current)"
  if [[ -z "$(git status --porcelain=v1)" ]]; then
    printf 'worktree_clean=true\n'
  else
    printf 'worktree_clean=false\n'
  fi
  docker compose version
  docker version --format 'docker_client={{.Client.Version}} docker_server={{.Server.Version}}'
  printf 'host_logical_cpus=%s\n' "$(getconf _NPROCESSORS_ONLN)"
  grep -E '^MemTotal:' /proc/meminfo
  grep -E '^cpu ' /proc/stat
  if [[ -z "$container" ]]; then
    printf 'container=not-running\n'
    return
  fi
  docker inspect "$container" --format \
    'configured_image={{.Config.Image}} image_id={{.Image}} status={{.State.Status}} restarts={{.RestartCount}} started_at={{.State.StartedAt}}{{if .State.Health}} health={{.State.Health.Status}}{{end}}'
  docker inspect "$container" --format 'nano_cpus={{.HostConfig.NanoCpus}} cpu_quota={{.HostConfig.CpuQuota}} cpu_period={{.HostConfig.CpuPeriod}} cpuset={{.HostConfig.CpusetCpus}} memory_limit={{.HostConfig.Memory}}'
  docker inspect "$container" --format \
    '{{range .Config.Env}}{{if or (eq . "NEKO_MEDIA_HLS_ENABLED=true") (eq . "NEKO_MEDIA_HLS_ENABLED=false") (eq . "NEKO_MEDIA_HLS_MODES=hls") (eq . "NEKO_MEDIA_HLS_MODES=hls ll-hls") (eq . "NEKO_MEDIA_WEBCODECS_WS_ENABLED=true")}}{{println .}}{{end}}{{end}}'
}

main() {
  local action="${1:-}" output stamp container metrics_url label count iteration
  case "$action" in -h|--help|help) usage; return;; esac
  [[ "$action" == init && $# -eq 2 || "$action" == snapshot && $# -eq 3 || "$action" == sample && $# -eq 4 ]] || { usage >&2; exit 2; }
  if [[ "$action" == sample ]]; then
    count="$4"
    [[ "$count" =~ ^([1-9]|[1-5][0-9]|60)$ ]] || fail 'COUNT must be 1..60'
    for ((iteration=1; iteration<=count; iteration++)); do
      main snapshot "$2" "$3"
      if ((iteration < count)); then sleep 10; fi
    done
    return
  fi
  umask 077
  for required in docker git curl grep realpath stat getconf; do
    command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
  done
  [[ -f docker-compose.yaml && -f "$RESULT_TEMPLATE" ]] || fail 'run from the repository root'
  output="$(prepare_output "$2")"
  stamp="$(date -u +%Y%m%dT%H%M%S%NZ)"
  container="$(container_id)"
  [[ -n "$container" || "$action" == init ]] || fail 'neko container is not running'
  environment_record "$container" >"$output/$stamp-environment.txt"

  if [[ "$action" == init ]]; then
    [[ -e "$output/RESULTS.md" ]] || cp -- "$RESULT_TEMPLATE" "$output/RESULTS.md"
  else
    label="$3"
    [[ "$label" =~ ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$ ]] || fail 'PHASE must be a short plain label'
    metrics_url="${NEKO_METRICS_URL:-http://127.0.0.1:8082/metrics}"
    [[ "$metrics_url" =~ ^http://(127\.0\.0\.1|localhost|\[::1\])(:[0-9]+)?(/[a-zA-Z0-9._/-]+)?/metrics$ ]] || \
      fail 'NEKO_METRICS_URL must be a credential-free loopback HTTP metrics URL'
    curl --disable --noproxy '*' --fail --silent --max-time 15 "$metrics_url" \
      | grep -E "$METRICS_PATTERN" >"$output/$stamp-$label.metrics.prom"
    docker stats --no-stream --format \
      'cpu={{.CPUPerc}} memory={{.MemUsage}} memory_percent={{.MemPerc}} network={{.NetIO}} block={{.BlockIO}} pids={{.PIDs}}' \
      "$container" >"$output/$stamp-$label.resources.txt"
  fi
  printf 'Recorded %s evidence: %s\n' "$action" "$output"
}

main "$@"
