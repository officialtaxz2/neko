#!/usr/bin/env bash

# Read-only target-server diagnosis. Never print raw logs, credentials or URLs.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 2 ]] || fail 'Usage: bash diagnose-hls-playback.sh REPOSITORY OUTPUT_DIR'
umask 077
for required in docker git curl python3 realpath stat; do
  command -v "$required" >/dev/null 2>&1 || fail "required command not found: $required"
done
helper_source="$(realpath -- "${BASH_SOURCE[0]}")"
cd -- "$1"
[[ -f docker-compose.yaml ]] || fail 'repository Compose file missing'
repository="$(realpath -- "$(git rev-parse --show-toplevel)")"
output="$(realpath -m -- "$2")"
case "$output/" in "$repository/"*) fail 'OUTPUT_DIR must be outside Git';; esac
case "$repository/" in "$output/"*) fail 'OUTPUT_DIR must not contain the repository';; esac
[[ -d "$output" && ! -L "$output" && "$(stat -c %a -- "$output")" == 700 ]] || fail 'existing private 0700 OUTPUT_DIR required'
[[ -f "$output/validation-commit.txt" ]] || fail 'exact preparation marker missing'
[[ "$(cat "$output/validation-commit.txt")" == "$(git rev-parse HEAD)" ]] || fail 'application checkout differs from preparation marker'
container="$(docker compose -f docker-compose.yaml ps -q neko)"
[[ "$container" =~ ^[0-9a-f]{12,64}$ ]] || fail 'exactly one running Neko container required'
report="$output/playback-diagnostic-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
git rev-parse HEAD >"$report/application-commit.txt"
git hash-object -- "$helper_source" >"$report/diagnostic-helper-blob.txt"
docker inspect "$container" --format \
  '{"status":{{json .State.Status}},"image_id":{{json .Image}},"restarts":{{json .RestartCount}},"oom_killed":{{json .State.OOMKilled}},"health":{{if .State.Health}}{{json .State.Health.Status}}{{else}}"absent"{{end}}}' \
  >"$report/container.json"
docker exec "$container" tail -n 3000 /var/log/neko/neko.log \
  >"$report/application.log" 2>"$report/application-capture-error.txt" || true
docker logs --since 10m --tail 600 "$container" \
  >"$report/supervisor.log" 2>&1 || true
docker exec "$container" gst-inspect-1.0 --version \
  >"$report/gstreamer-version.txt" 2>"$report/gstreamer-capture-error.txt" || true
if curl --disable --noproxy '*' --fail --silent --max-time 10 \
  http://127.0.0.1:8082/metrics >"$report/metrics.prom" 2>"$report/metrics-capture-error.txt"; then
  printf 'true\n' >"$report/metrics-available.txt"
else
  printf 'false\n' >"$report/metrics-available.txt"
fi

# Reconstruct output from fixed markers/enumerations and numeric values only.
# No original log line, error string, header, session ID or lease path is output.
python3 - "$report" <<'PY'
import collections
import json
import math
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
def read(name):
    return (root / name).read_text(encoding="utf-8", errors="replace")

messages = (
    "HLS negotiation rejected", "HLS bootstrap ticket creation failed",
    "HLS generation start failed", "HLS packager generation started",
    "HLS packager ready", "HLS sample rejected",
    "HLS rendition removed after output stall", "HLS rendition rejoined on fresh keyframe",
    "HLS lease opened", "HLS lease state changed", "HLS lease paused",
    "HLS lease resume warming", "HLS lease resume failed", "HLS lease resumed",
    "HLS lease closed", "HLS lease limit rejected", "HLS request limit rejected",
    "HLS bootstrap response write failed", "HLS packager stopped after idle grace",
)
errors = (
    "not-negotiated", "Internal data stream error", "could not link",
    "no element", "no property", "channel-mapping-family",
    "HLS source codec unsupported", "HLS transcoder output format does not match advertised rendition",
    "HLS transcoder codec configuration changed", "HLS transcode timeline regressed",
    "HLS transcode timeline gap", "HLS two-second boundary did not begin with IDR",
    "HLS decode timestamp regressed", "SIGSEGV", "SIGABRT", "stack smashing",
)
counts = collections.Counter()
error_counts = collections.Counter()
recent = []
lines = read("application.log").splitlines()
for raw in lines:
    line = re.sub(r"\x1b\[[0-9;]*m", "", raw)
    for marker in messages:
        if marker in line:
            counts[marker] += 1
            recent.append(marker)
    # Require the GStreamer/media module, or a process crash marker. This still
    # is a bounded log sample, not an assertion that any marker caused playback.
    if "gstreamer" in line or "mediahls" in line or any(x in line for x in ("SIGSEGV", "SIGABRT", "stack smashing")):
        for marker in errors:
            if marker in line:
                error_counts[marker] += 1

allowed = {
    "mode": {"hls", "ll-hls"},
    "variant": {"audio", "high", "medium", "low", "unknown"},
    "kind": {"audio", "video", "init", "part", "segment", "unknown"},
    "stage": {"provider", "worker", "packager"},
    "resource": {"bootstrap", "master", "playlist", "init", "segment", "part", "keepalive"},
    "state": {"opening", "active", "paused", "closed", "running", "failed", "blocked", "waiting"},
    "reason": {"initial", "startup", "resume", "timestamp_reset", "source_restart", "source_end", "format_change", "provider_overflow", "worker_failure", "rendition_rejoin", "queue_full", "keyframe_admission", "unknown"},
    "result": {"success", "error", "bad_request", "read_error", "too_large", "gone", "unauthorized", "denied", "not_ready", "rate_limited", "backend_error", "write_error", "not_found", "paused", "range_invalid", "timeout", "canceled"},
}
families = {"bootstrap_total", "requests_total", "leases", "packagers", "packager_starts_total", "generations_total", "drops_total", "objects", "published_bytes_total"}
metrics = []
for line in read("metrics.prom").splitlines():
    match = re.fullmatch(r"neko_media_hls_([a-z_]+)\{([^{}]*)\} ([0-9.eE+\-]+)", line)
    if not match or match[1] not in families:
        continue
    pairs = match[2].split(",")
    labels = {}
    for pair in pairs:
        label = re.fullmatch(r'([a-z_]+)="([a-z_\-]+)"', pair)
        if not label or label[1] not in allowed or label[2] not in allowed[label[1]] or label[1] in labels:
            break
        labels[label[1]] = label[2]
    else:
        value = float(match[3])
        if math.isfinite(value):
            metrics.append({"metric": match[1], "labels": labels, "value": value})

container = json.loads(read("container.json"))
safe_container = {
    "status": container.get("status") if container.get("status") in {"running", "exited", "restarting", "paused", "dead", "created", "removing"} else "unknown",
    "health": container.get("health") if container.get("health") in {"healthy", "unhealthy", "starting", "absent"} else "unknown",
    "restarts": container.get("restarts") if type(container.get("restarts")) is int else None,
    "oom_killed": container.get("oom_killed") is True,
    "image_id": container.get("image_id") if re.fullmatch(r"sha256:[0-9a-f]{64}", str(container.get("image_id"))) else "unknown",
}
version = re.search(r"GStreamer (\d+\.\d+\.\d+)", read("gstreamer-version.txt"))
supervisor = read("supervisor.log")
summary = {
    "application_commit": read("application-commit.txt").strip(),
    "diagnostic_helper_blob": read("diagnostic-helper-blob.txt").strip(),
    "container": safe_container,
    "gstreamer_version": version[1] if version else "unavailable",
    "application_log_lines_captured": len(lines),
    "neko_process_exits_in_supervisor_sample": len(re.findall(r"exited: neko\b", supervisor)),
    "metrics_available": read("metrics-available.txt").strip() == "true",
    "hls_log_marker_counts": dict(counts),
    "recent_hls_markers": recent[-16:],
    "fixed_error_marker_counts": dict(error_counts),
    "hls_metrics": metrics,
    "limits": "Log counts are bounded samples; metrics are cumulative since process start. Missing markers are not proof of success. No credentialed playback request was made.",
}
encoded = json.dumps(summary, indent=2, sort_keys=True)
(root / "safe-summary.json").write_text(encoded + "\n", encoding="utf-8")
print(encoded)
PY
printf 'Private diagnostic saved; share only the safe summary above.\n'
