#!/usr/bin/env bash

# Target server only. Adapt the active configuration in memory and report only
# fixed structural metadata. No reload, file change or raw configuration output.
set -Eeuo pipefail
[[ $# -le 1 ]] || { printf 'Usage: bash inspect-hls-caddy.sh [CADDYFILE]\n' >&2; exit 2; }
caddy adapt --config "${1:-/etc/caddy/Caddyfile}" --adapter caddyfile |
  docker run --rm -i --network none --read-only python:3.12-alpine python -c '
import json
import sys

config = json.load(sys.stdin)
target = "neko.taxzvps.de"
servers = config.get("apps", {}).get("http", {}).get("servers", {})
hosts = set()
target_routes = []

def walk(value):
    if isinstance(value, dict):
        match_hosts = {
            host for matcher in value.get("match", []) if isinstance(matcher, dict)
            for host in matcher.get("host", [])
        }
        hosts.update(match_hosts)
        if target in match_hosts:
            target_routes.append(value)
        for child in value.values():
            walk(child)
    elif isinstance(value, list):
        for child in value:
            walk(child)

for server in servers.values():
    walk(server.get("routes", []))

proxies = []
def proxy_walk(value):
    if isinstance(value, dict):
        if value.get("handler") == "reverse_proxy":
            proxies.append(value)
        for child in value.values():
            proxy_walk(child)
    elif isinstance(value, list):
        for child in value:
            proxy_walk(child)

for route in target_routes:
    proxy_walk(route.get("handle", []))

logs = config.get("logging", {}).get("logs", {})
default = logs.get("default", {})
encoder = default.get("encoder", {})
known_formats = {"json", "console", "filter"}
summary = {
    "http_server_count": len(servers),
    "distinct_explicit_site_host_count": len(hosts),
    "neko_exact_host_route_count": len(target_routes),
    "neko_reverse_proxy_count": len(proxies),
    "neko_proxy_structure": [{
        "upstream_count": len(proxy.get("upstreams", [])),
        "only_expected_loopback_upstream": proxy.get("upstreams") == [{"dial": "127.0.0.1:8082"}],
        "forwarded_header_deleted": "Forwarded" in proxy.get("headers", {}).get("request", {}).get("delete", []),
        "upstream_compression_disabled": proxy.get("transport", {}).get("compression") is False,
        "explicit_request_buffering": "request_buffers" in proxy,
        "explicit_response_buffering": "response_buffers" in proxy,
    } for proxy in proxies],
    "runtime_logging_config_present": "logging" in config,
    "runtime_default_logger_present": "default" in logs,
    "runtime_default_encoder_format": encoder.get("format") if encoder.get("format") in known_formats else "unset-or-other",
    "runtime_default_filter_fields_count": len(encoder.get("fields", {})),
    "other_runtime_logger_count": len(logs) - int("default" in logs),
    "neko_site_logger_configured": any(target in server.get("logs", {}).get("logger_names", {}) for server in servers.values()),
    "any_log_credentials_enabled": any(server.get("logs", {}).get("should_log_credentials", False) for server in servers.values()),
}
print(json.dumps(summary, indent=2))
'
