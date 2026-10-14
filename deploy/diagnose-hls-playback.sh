#!/usr/bin/env bash

# Read-only target-server diagnosis. Never print raw logs, credentials or URLs.
set -Eeuo pipefail
fail() { printf 'error: %s\n' "$1" >&2; exit 1; }
[[ $# -eq 2 || ( $# -eq 3 && "${3:-}" == --progress ) ]] || fail 'Usage: bash diagnose-hls-playback.sh REPOSITORY OUTPUT_DIR [--progress]'
progress=false
[[ $# -ne 3 ]] || progress=true
umask 077
for required in docker git curl python3 realpath stat date sleep grep; do
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
if [[ "$progress" == true ]]; then
  application_commit="$(git rev-parse HEAD)"
  image_tag="hls-${application_commit:0:12}"
  [[ -f "$output/images.txt" ]] || fail 'prepared image record missing'
  image_record="$(grep -E "^image=(docker.io/)?my-neko/brave:$image_tag id=sha256:[0-9a-f]{64} created=" "$output/images.txt")" || fail 'prepared application image not recorded'
  [[ "$image_record" != *$'\n'* ]] || fail 'ambiguous prepared image record'
  expected_image="${image_record#* id=}"
  expected_image="${expected_image%% *}"
  [[ "$(docker inspect -f '{{.Image}}' "$container")" == "$expected_image" ]] || fail 'live image differs from the prepared application'
fi
report="$output/playback-diagnostic-$(date -u +%Y%m%dT%H%M%S%NZ)"
mkdir -m 700 -- "$report"
printf '%s\n' "$progress" >"$report/progress-mode.txt"
git rev-parse HEAD >"$report/application-commit.txt"
git hash-object -- "$helper_source" >"$report/diagnostic-helper-blob.txt"
docker inspect "$container" --format \
  '{"status":{{json .State.Status}},"image_id":{{json .Image}},"started_at":{{json .State.StartedAt}},"restarts":{{json .RestartCount}},"oom_killed":{{json .State.OOMKilled}},"health":{{if .State.Health}}{{json .State.Health.Status}}{{else}}"absent"{{end}}}' \
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
if [[ "$progress" == true ]]; then
  date +%s.%N >"$report/progress-started-at.txt"
  sleep 12
  if curl --disable --noproxy '*' --fail --silent --max-time 10 \
    http://127.0.0.1:8082/metrics >"$report/metrics-after.prom" 2>"$report/metrics-after-capture-error.txt"; then
    printf 'true\n' >"$report/metrics-after-available.txt"
  else
    printf 'false\n' >"$report/metrics-after-available.txt"
  fi
  date +%s.%N >"$report/progress-finished-at.txt"
  docker inspect "$container" --format \
    '{"status":{{json .State.Status}},"image_id":{{json .Image}},"started_at":{{json .State.StartedAt}},"restarts":{{json .RestartCount}},"oom_killed":{{json .State.OOMKilled}},"health":{{if .State.Health}}{{json .State.Health.Status}}{{else}}"absent"{{end}}}' \
    >"$report/container-after.json"
  # Retain the newest bounded log sample, including events during the interval.
  docker exec "$container" tail -n 3000 /var/log/neko/neko.log \
    >"$report/application.log" 2>"$report/application-capture-error.txt" || true
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

progress = read("progress-mode.txt").strip() == "true"
messages = (
    "HLS negotiation rejected", "HLS bootstrap ticket creation failed",
    "HLS generation start failed", "HLS packager generation started",
    "HLS packager ready", "HLS sample rejected", "HLS worker restart requested",
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
metric_text = read("metrics-after.prom" if progress else "metrics.prom")
for line in metric_text.splitlines():
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

container = json.loads(read("container-after.json" if progress else "container.json"))
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
    "metrics_available": read("metrics-after-available.txt" if progress else "metrics-available.txt").strip() == "true",
    "hls_log_marker_counts": dict(counts),
    "recent_hls_markers": recent[-16:],
    "fixed_error_marker_counts": dict(error_counts),
    "hls_metrics": metrics,
    "limits": "Log counts are bounded samples; metrics are cumulative since process start. Missing markers are not proof of success. No credentialed playback request was made.",
}
encoded = json.dumps(summary, indent=2, sort_keys=True)
(root / "safe-summary.json").write_text(encoded + "\n", encoding="utf-8")
if not progress:
    print(encoded)
else:
    # Fixed family/label allowlists exclude peer IDs, credentials and URLs.
    # Gauges are shown as before/after; counters have deltas only when both
    # samples exist in the same process and no observed counter regresses.
    source_ids = {"audio", "high", "medium", "low", "default"}
    progress_allowed = dict(allowed)
    progress_allowed.update({
        "backend": {"hls", "webrtc", "webcodecs-ws"},
        "source_id": source_ids, "video_id": source_ids,
        "codec_name": {"vp8", "opus"}, "codec_type": {"video", "audio"},
        "submodule": {"streamsink"},
        "result": allowed["result"] | {"attempt"},
        "reason": allowed["reason"] | {"source_switch", "paused", "resumed"},
    })
    progress_families = {
        "neko_media_hls_published_bytes_total", "neko_media_hls_requests_total",
        "neko_media_hls_generations_total", "neko_media_hls_drops_total",
        "neko_media_hls_leases", "neko_media_hls_packagers",
        "neko_media_source_generation", "neko_media_subscriptions_active",
        "neko_media_subscription_delivered_units_total",
        "neko_media_subscription_discontinuities_total",
        "neko_capture_streamsink_bytes", "neko_capture_pipelines_total",
        "neko_media_delivery_opens_total",
    }

    def parse_metrics(text):
        values = {}
        for line in text.splitlines():
            match = re.fullmatch(r'([a-z_]+)\{([^{}]*)\} ([0-9.eE+\-]+)', line)
            if not match or match[1] not in progress_families:
                continue
            labels = {}
            for pair in match[2].split(","):
                label = re.fullmatch(r'([a-z_]+)="([a-z_0-9\-]+)"', pair)
                if not label or label[1] not in progress_allowed or label[2] not in progress_allowed[label[1]] or label[1] in labels:
                    break
                labels[label[1]] = label[2]
            else:
                value = float(match[3])
                if math.isfinite(value) and value >= 0:
                    values[(match[1], tuple(sorted(labels.items())))] = value
        return values

    before_text = read("metrics.prom")
    before, after = parse_metrics(before_text), parse_metrics(metric_text)

    def process_start(text):
        match = re.search(r'^process_start_time_seconds ([0-9.eE+\-]+)$', text, re.MULTILINE)
        if match:
            value = float(match[1])
            if math.isfinite(value) and value > 0:
                return value
        return None

    initial_container = json.loads(read("container.json"))
    started_before, started_after = process_start(before_text), process_start(metric_text)
    same_process = (
        started_before is not None and started_before == started_after and
        initial_container.get("image_id") == container.get("image_id") and
        initial_container.get("started_at") == container.get("started_at") and
        initial_container.get("restarts") == container.get("restarts")
    )
    available = read("metrics-available.txt").strip() == "true" and read("metrics-after-available.txt").strip() == "true"
    # streamsink_bytes is a cumulative counter despite its legacy name.
    counter_families = {name for name in progress_families if name.endswith("_total")} | {"neko_capture_streamsink_bytes"}
    reset_seen = any(before[key] > value for key, value in after.items() if key in before and key[0] in counter_families)
    delta_valid = available and same_process and not reset_seen

    def rows(family, *, delta=False, positive=False):
        result = []
        for key, value in sorted(after.items()):
            if key[0] != family:
                continue
            labels = dict(key[1])
            if delta:
                # Prometheus CounterVec emits no series before its first event.
                # An absent baseline is zero only for an otherwise valid scrape.
                value = value - before.get(key, 0) if delta_valid else None
            if positive and value == 0:
                continue
            item = {"labels": labels, "delta" if delta else "value": value}
            if not delta and key[0] not in counter_families:
                item["before"] = before.get(key)
            result.append(item)
        return result

    seconds = float(read("progress-finished-at.txt")) - float(read("progress-started-at.txt"))
    elapsed = round(seconds, 3) if math.isfinite(seconds) and seconds >= 0 else None
    compact = {
        "application_commit": summary["application_commit"],
        "diagnostic_helper_blob": summary["diagnostic_helper_blob"],
        "container": safe_container,
        "sample_seconds": elapsed, "delta_valid": delta_valid,
        "same_process": same_process, "counter_reset_seen": reset_seen,
        "metrics_available": {
            "before": read("metrics-available.txt").strip() == "true",
            "after": read("metrics-after-available.txt").strip() == "true",
        },
        "application_log_lines_captured": summary["application_log_lines_captured"],
        "neko_process_exits_in_supervisor_sample": summary["neko_process_exits_in_supervisor_sample"],
        "source_generation": rows("neko_media_source_generation"),
        "capture_bytes_delta": rows("neko_capture_streamsink_bytes", delta=True),
        "hls_publication_delta": rows("neko_media_hls_published_bytes_total", delta=True),
        "hls_subscriptions": [r for r in rows("neko_media_subscriptions_active")
                              if r["labels"].get("backend") == "hls"],
        "hls_source_units_delta": [r for r in rows("neko_media_subscription_delivered_units_total", delta=True)
                                   if r["labels"].get("backend") == "hls"],
        "capture_pipeline_creations_delta": rows("neko_capture_pipelines_total", delta=True, positive=True),
        "hls_generations": rows("neko_media_hls_generations_total"),
        "hls_generations_delta": rows("neko_media_hls_generations_total", delta=True, positive=True),
        "hls_drops_delta": rows("neko_media_hls_drops_total", delta=True, positive=True),
        "hls_discontinuities": [r for r in rows("neko_media_subscription_discontinuities_total")
                                if r["labels"].get("backend") == "hls"],
        "hls_discontinuities_delta": [r for r in rows("neko_media_subscription_discontinuities_total", delta=True, positive=True)
                                      if r["labels"].get("backend") == "hls"],
        "http_results_delta": rows("neko_media_hls_requests_total", delta=True, positive=True),
        "webrtc_opens": [r for r in rows("neko_media_delivery_opens_total")
                         if r["labels"].get("backend") == "webrtc" and r["labels"].get("result") == "success"],
        "webrtc_opens_delta": [r for r in rows("neko_media_delivery_opens_total", delta=True)
                               if r["labels"].get("backend") == "webrtc" and r["labels"].get("result") == "success"],
        "leases": rows("neko_media_hls_leases"), "packagers": rows("neko_media_hls_packagers"),
        "hls_log_marker_counts": summary["hls_log_marker_counts"],
        "recent_hls_markers": summary["recent_hls_markers"],
        "fixed_error_marker_counts": summary["fixed_error_marker_counts"],
        "limits": "Two server scrapes and bounded cumulative logs only. No browser state, per-viewer request attribution or join/freeze causality. Reload erased the earlier browser state. Missing/changed process invalidates deltas; absent series alone is not zero production. No credentialed media request or service change.",
    }
    result = json.dumps(compact, indent=2, sort_keys=True)
    (root / "safe-progress-summary.json").write_text(result + "\n", encoding="utf-8")
    print(result)
PY
printf 'Private diagnostic saved; share only the safe summary above.\n'
