#!/usr/bin/env python3
"""Target-only, file-only summary of an already saved HLS failure diagnostic."""

import collections
import json
import math
import pathlib
import re
import stat
import sys


def fail(message):
    print("error: " + message, file=sys.stderr)
    raise SystemExit(1)


def read_private(path):
    if path.is_symlink():
        fail("symlink evidence file rejected")
    info = path.stat()
    if not stat.S_ISREG(info.st_mode) or info.st_mode & 0o077:
        fail("private regular evidence file required")
    if info.st_size > 8 * 1024 * 1024:
        fail("evidence file exceeds summary bound")
    return path.read_text(encoding="utf-8", errors="replace")


def field(line, record, name):
    if record is not None:
        value = record.get(name)
        return value if isinstance(value, str) else ""
    match = re.search(r'(?:^|\s)' + re.escape(name) + r'=("[^"\r\n]*"|[^\s]+)(?=\s|$)', line)
    return match[1].strip('"') if match else ""


def bootstrap_durations(text):
    """Return only the fixed bootstrap histogram, not arbitrary metric labels."""
    # Match the registered buckets in server/internal/mediahls/metrics.go.
    bounds = ("0.001", "0.005", "0.01", "0.025", "0.05", "0.1", "0.25",
              "0.5", "1", "2", "5", "7", "15", "30", "+Inf")
    modes = {}
    for line in text.splitlines():
        match = re.fullmatch(
            r'neko_media_hls_request_duration_seconds_(sum|count|bucket)\{([^{}]*)\} ([0-9.eE+\-]+)', line)
        if not match:
            continue
        labels = {}
        for pair in match[2].split(","):
            label = re.fullmatch(r'([a-z_]+)="([^"\\]*)"', pair)
            if not label or label[1] in labels:
                break
            labels[label[1]] = label[2]
        else:
            expected = {"mode", "resource", "le"} if match[1] == "bucket" else {"mode", "resource"}
            if set(labels) != expected or labels["mode"] not in {"hls", "ll-hls"} or labels["resource"] != "bootstrap":
                continue
            if match[1] == "bucket" and labels["le"] not in bounds:
                continue
            value = float(match[3])
            if not math.isfinite(value) or value < 0 or (match[1] != "sum" and not value.is_integer()):
                continue
            series = modes.setdefault(labels["mode"], {})
            key = (match[1], labels.get("le", ""))
            if key in series:
                fail("duplicate bootstrap histogram series in saved evidence")
            series[key] = value
    return [{
        "mode": mode,
        "observations": int(series[("count", "")]) if ("count", "") in series else None,
        "total_seconds": series.get(("sum", "")),
        "cumulative_buckets": [{"upper_bound_seconds": bound,
                                "observations": int(series[("bucket", bound)])}
                               for bound in bounds if ("bucket", bound) in series],
    } for mode, series in sorted(modes.items())]


def main():
    if len(sys.argv) != 3 or not re.fullmatch(r"[0-9a-f]{40}", sys.argv[2]):
        fail("Usage: python3 summarize-hls-startup.py OUTPUT_DIR APPLICATION_COMMIT")
    supplied = pathlib.Path(sys.argv[1])
    if supplied.is_symlink():
        fail("symlink output directory rejected")
    root = supplied.resolve(strict=True)
    if not root.is_dir() or stat.S_IMODE(root.stat().st_mode) != 0o700:
        fail("existing private 0700 output directory required")
    pattern = r"playback-diagnostic-([0-9]{8}T[0-9]{15}Z)"
    reports = sorted(path for path in root.iterdir() if re.fullmatch(pattern, path.name))
    if not reports:
        fail("saved playback diagnostic missing; do not enable HLS for this summary")
    report = reports[-1]
    if report.is_symlink() or not report.is_dir() or stat.S_IMODE(report.stat().st_mode) != 0o700:
        fail("private regular diagnostic directory required")
    if read_private(report / "application-commit.txt").strip() != sys.argv[2]:
        fail("latest diagnostic differs from expected application commit")

    stages = (
        ("creating pipeline", "capture_pipeline_create"),
        ("capture pipeline parsed", "capture_pipeline_parsed"),
        ("capture appsink attached", "capture_appsink_attached"),
        ("capture pipeline play started", "capture_play_started"),
        ("capture pipeline play completed", "capture_play_completed"),
        ("first listener, starting", "capture_first_listener"),
        ("adding listener", "capture_listener_adding"),
        ("started emitting samples", "capture_sample_loop_start"),
        ("removing listener", "capture_listener_removing"),
        ("last listener, stopping", "capture_last_listener"),
        ("destroying pipeline", "capture_pipeline_destroy"),
        ("stopped emitting samples", "capture_sample_loop_stop"),
        ("HLS negotiation rejected", "hls_negotiation_rejected"),
        ("HLS generation start failed", "hls_generation_start_failed"),
        ("HLS packager generation started", "hls_generation_started"),
        ("HLS packager ready", "hls_packager_ready"),
        ("HLS lease opened", "hls_lease_opened"),
        ("HLS sample rejected", "hls_sample_rejected"),
        ("HLS worker restart requested", "hls_worker_restart_requested"),
        ("HLS packager idle stop scheduled", "hls_idle_stop_scheduled"),
        ("HLS packager stopped after idle grace", "hls_idle_stopped"),
    )
    source_ids = {"audio", "high", "medium", "low", "default"}
    rejects = {"invalid_payload", "not_allowed", "rate_limited"}
    event_names = {"media/hls/capabilities/request", "media/hls/create"}
    sequence = []
    counts = collections.Counter()
    lines = read_private(report / "application.log").splitlines()
    for index, raw in enumerate(lines, 1):
        line = re.sub(r"\x1b\[[0-9;]*m", "", raw)
        record = None
        try:
            decoded = json.loads(line)
            if isinstance(decoded, dict):
                record = decoded
        except json.JSONDecodeError:
            pass
        module = field(line, record, "module")
        if module not in {"capture", "mediahls"}:
            continue
        message = record.get("message", "") if record is not None else line
        if not isinstance(message, str):
            continue
        for marker, stage in stages:
            if marker not in message:
                continue
            item = {"line": index, "stage": stage}
            source = field(line, record, "id")
            if source in source_ids:
                item["source"] = source
            if stage == "hls_negotiation_rejected":
                reason = field(line, record, "reason")
                event = field(line, record, "event")
                item["reason"] = reason if reason in rejects else "unknown"
                item["event"] = event if event in event_names else "unknown"
            if stage == "hls_worker_restart_requested":
                variant = field(line, record, "variant")
                worker_stage = field(line, record, "stage")
                reason = field(line, record, "reason")
                item["variant"] = variant if variant in {"audio", "high", "medium", "low"} else "unknown"
                item["worker_stage"] = worker_stage if worker_stage in {"input", "output", "anchor", "monitor"} else "unknown"
                item["reason"] = reason if reason in {"push_failed", "queue_full", "drops_closed", "samples_closed", "output_stall"} else "unknown"
            counts[stage] += 1
            sequence.append(item)

    families = {
        "neko_media_delivery_opens_total", "neko_media_deliveries",
        "neko_media_subscriptions_active", "neko_media_source_generation",
        "neko_capture_pipelines_active", "neko_capture_streamsink_listeners",
        "neko_capture_streamsink_bytes", "neko_capture_streamsink_bitrate",
        "neko_media_hls_packager_starts_total", "neko_media_hls_generations_total",
    }
    allowed = {
        "backend": {"hls", "webrtc", "webcodecs-ws"},
        "source_id": source_ids, "video_id": source_ids,
        "kind": {"audio", "video"}, "codec_name": {"opus", "vp8"},
        "codec_type": {"audio", "video"},
        "variant": {"audio", "high", "medium", "low", "unknown"},
        "result": {"attempt", "success", "error"},
        "state": {"opening", "active", "paused", "failed", "closed"},
        "reason": {"initial", "startup", "resume", "timestamp_reset", "source_restart", "source_end", "format_change", "provider_overflow", "worker_failure", "rendition_rejoin", "unknown"},
    }
    metrics = []
    metric_text = read_private(report / "metrics.prom")
    for line in metric_text.splitlines():
        match = re.fullmatch(r'([a-z_]+)\{([^{}]*)\} ([0-9.eE+\-]+)', line)
        if not match or match[1] not in families:
            continue
        labels = {}
        for pair in match[2].split(","):
            label = re.fullmatch(r'([a-z_]+)="([a-z_0-9\-]+)"', pair)
            if not label or label[1] not in allowed or label[2] not in allowed[label[1]] or label[1] in labels:
                break
            labels[label[1]] = label[2]
        else:
            value = float(match[3])
            if math.isfinite(value):
                metrics.append({"metric": match[1], "labels": labels, "value": value})

    # Summarize the closest earlier environment record only if its exact image
    # and application match the failure diagnostic. Never print raw env lines.
    stamp = re.fullmatch(pattern, report.name)[1]
    environments = sorted(path for path in root.iterdir()
                          if re.fullmatch(r"[0-9]{8}T[0-9]{15}Z-environment\.txt", path.name)
                          and path.name.split("-", 1)[0] <= stamp)
    environment = {"matching_saved_record": False, "hls_enabled": None}
    if environments:
        text = read_private(environments[-1])
        saved_container = json.loads(read_private(report / "container.json"))
        image = saved_container.get("image_id") if isinstance(saved_container, dict) else None
        if isinstance(image, str) and re.fullmatch(r"sha256:[0-9a-f]{64}", image):
            image_match = re.search(r"(?:^|\s)image_id=(sha256:[0-9a-f]{64})(?=\s|$)", text)
            commit_match = re.search(r"^commit=([0-9a-f]{40})$", text, re.MULTILINE)
            if image_match and commit_match and image_match[1] == image and commit_match[1] == sys.argv[2]:
                environment["matching_saved_record"] = True
                flags = {line for line in text.splitlines()
                         if line in {"NEKO_MEDIA_HLS_ENABLED=true", "NEKO_MEDIA_HLS_ENABLED=false"}}
                if len(flags) == 1:
                    environment["hls_enabled"] = "NEKO_MEDIA_HLS_ENABLED=true" in flags

    print(json.dumps({
        "application_commit": sys.argv[2], "diagnostic_stamp": stamp,
        "application_log_lines_captured": len(lines),
        "stage_counts": dict(counts), "stage_sequence": sequence[-80:],
        "capture_delivery_metrics": metrics, "saved_environment": environment,
        "bootstrap_request_durations": bootstrap_durations(metric_text),
        "limits": "Saved bounded logs and cumulative metrics only; sequence spans all participants, not one correlated HLS attempt. Bootstrap durations aggregate all results; histogram buckets give ranges, not exact per-attempt times or the time media first became ready. An earlier matching environment record is not proof of live enablement. No service access/change, credentialed request or new playback attempt.",
    }, indent=2, sort_keys=True))


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, OverflowError):
        # Exception details can contain private file/log data; keep them local.
        fail("saved evidence unavailable or invalid; no service change applied")
