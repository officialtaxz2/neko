# HLS behind the host Caddy service

The operator confirmed on 2026-10-04 that Caddy runs as a **system service on the
host**, outside the Neko Compose project. The supplied Phase 4 output confirms
**Caddy 2.6.2** and an active service. The operator then supplied the Neko site:
`neko.taxzvps.de { reverse_proxy 127.0.0.1:8082 }`, with only commented upload/old
port examples. No access logger, imports, buffering, rewriting, cache or global
options were present in the supplied text. The operator subsequently confirmed
the active configuration path as `/etc/caddy/Caddyfile`; actual runtime/error
output still needs confirmation. At exact application commit `93f1fa63`, the
target validated both the active file and prepared proposal successfully;
the only proposal warning concerns formatting. This is a review guide; no Caddy
configuration has been changed or reloaded by Codex.

The concrete reviewed proposal is [deploy/caddy-hls.example](../deploy/caddy-hls.example).
Set its example hostname to the supplied public hostname and merge, rather than
discard, any other active sites/global options. It retains the existing port,
keeps site access logging absent, adds the runtime encoder and removes ambiguous
forwarding metadata/automatic upstream compression. Validate the actual final
file before an operator-applied reload; reload can reconnect existing streams.

The [activation helper](../deploy/activate-hls-phase4.sh) was prepared under
the assumption that the supplied site was the complete file. The actual target
attempt at helper commit `58ca75e9` rejected that sole-site condition with exit
1, **before backup/reload or any Neko stop/deployment**.
The [read-only structural inspection](../deploy/inspect-hls-caddy.sh) then
passed on the target at helper `68a7f904` (exit 0). It reported one HTTP server,
11 distinct explicit hosts, one exact Neko route and one proxy with only
`127.0.0.1:8082`. Forwarded deletion, disabled upstream compression and explicit
request/response buffering were absent. The default runtime logger has no
reported standard encoder/filter, one additional logger exists, Neko has no
exact host-to-logger association and credential logging is disabled. The actual
default encoder and additional logger routing still need the merge guards;
`unset-or-other` alone does not prove absence of a custom module.

The updated [activation helper](../deploy/activate-hls-phase4.sh) uses a matching
[offline source merger](../deploy/merge-hls-caddy.py), extracted beside it. The
merger supports the supplied explicit Neko site/bare loopback proxy and a default
logger with no existing encoder. It adds only the reviewed proxy settings and
runtime filter, preserving source text elsewhere. Existing global/default writer,
level, include/exclude settings and other loggers remain intact. It rejects an
unreviewed encoder, Neko access/fallback logger, additional runtime/error logger,
debug logging, existing proxy options, or imported/aliased Neko source block.

Caddy adapts the candidate from a private temporary file in `/etc/caddy` so
relative imports retain their directory. The entire adapted JSON must equal the
original with **only** the two Neko proxy settings and default encoder added.
Every other host, route, TLS/global option and logger must remain identical.
Caddy validation and a fresh source/adaptation check precede backup/reload.
Raw configurations/validation output remain private; only fixed verdicts/counts
are printed. The temporary file is removed on success or failure.

The later synthetic 502, private journal check, tested-image enablement and
failure restoration remain unchanged. The new merge/activation variant is
statically reviewed, **NOT EXECUTED IN CODEX**, and awaits target execution.
No reload, live-log pass or enabled playback is claimed yet. The default filter
also deletes headers from other runtime error records; existing separate site
access loggers are preserved, and no Neko access logger is added.

## Existing routing and trust

Use the existing public HTTPS Neko origin and preserve its current routing,
event-WebSocket support and site certificate. The HLS overlay reuses the
reviewed WebCodecs origin/proxy settings when the HLS-specific values are empty.
The public origin must be exactly `https://host[:port]`, without path or slash.
If Neko uses `server.path_prefix`, include that path in the HTTP check's
`NEKO_PUBLIC_BASE_URL`, not in the allowed origin.

Trust only the immediate peer Neko actually receives. A host connection to the
published Docker port may arrive inside Neko as the Docker bridge gateway;
do not assume this is `127.0.0.1`. Preserve the already reviewed peer value and
verify it on the target. Keep the published backend port bound to loopback.
Do not expand trust to all private networks to resolve a failed probe.

Inspect the actual service unit/configuration locally on the server. The
operator-confirmed Caddyfile is `/etc/caddy/Caddyfile`. Record
`caddy version` and `systemctl is-active caddy`. Validate the actual active
Caddyfile with `caddy validate --config <actual-path> --adapter caddyfile`
before any operator-applied change. Share only the relevant sanitized Neko
routing/logging fragment, not the whole multi-site configuration or environment.

## Streaming and headers

The existing Neko proxy should preserve the original path, cookies, Origin,
Range, status and response security headers. Do not add caching or wildcard
CORS to `/api/media/hls/*`. A second `handle_path` that strips a configured Neko
prefix can break its cookie scope and media URLs.

Within the existing `reverse_proxy` block, these are the relevant settings to
review and, where required, incorporate:

```caddyfile
header_up -Forwarded
transport http {
    compression off
}
```

Removing a client-supplied `Forwarded` avoids ambiguity with Caddy's generated
`X-Forwarded-Proto`; the HLS policy rejects both forms being present together.
Keep the backend's normal HTTPS detection through the reviewed proxy peer.
Disable whole-response buffering on this route if explicitly configured and
verify that an aborted downstream request cancels its upstream handler.

Caddy documents that `flush_interval -1` also leaves the backend request running
after a client disconnect. Therefore do not add it as an unconditional streaming
fix. Start with the existing streaming behavior and measure LL blocking reload,
publication and cancellation before changing it. `compression off` disables
Caddy's automatic upstream gzip request, while client-requested playlist gzip
can still reach Neko. MP4 responses must stay uncompressed. See the official
[reverse_proxy reference](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy).

## Safe logging before valid credentials

The supplied minimal site needs no extra HLS route. It preserves path and
cookies through the existing upstream. No site `log` directive means no access
logger is enabled by that snippet; runtime HTTP errors still capture request
data in [Caddy 2.6.2 server code](https://github.com/caddyserver/caddy/blob/v2.6.2/modules/caddyhttp/server.go).
Therefore review/apply the runtime encoder separately before valid leases.

Neko normalizes its own HLS access paths. Caddy access **and runtime/error** logs
need their own review. A query-only filter does not remove the public lease ID
embedded in the path. Redacting only the cookie value also leaves its scoped
`Path` in `Set-Cookie`.

This encoder fragment can be integrated into the existing Neko access logger
and separately into the existing global runtime logger; it is not a replacement
for the site's routing or the complete global options block:

```caddyfile
format filter {
    fields {
        request>uri regexp "^(/[^?]*)?/api/media/hls.*$" "/api/media/hls/:redacted"
        uri regexp "^(/[^?]*)?/api/media/hls.*$" "/api/media/hls/:redacted"
        request>headers delete
        resp_headers delete
    }
    wrap json
}
```

The installed version already contains the regexp/delete encoder modules in
[Caddy v2.6.2 source](https://github.com/caddyserver/caddy/blob/v2.6.2/modules/logging/filters.go).
This is source-level capability evidence; the actual Caddyfile must still pass
validation and the target access/error-log checks.

This removes header fields from that logger and normalizes the HLS path and
query, including deployments under a prefix. Keep `log_credentials` disabled;
review imported log encoders, additional URI fields, log appenders and debug
messages as well. The encoder cannot sanitize an arbitrary URI inserted into a
log message. Access `log_skip` alone does not cover runtime errors. The official
[log reference](https://caddyserver.com/docs/caddyfile/directives/log) describes
encoder filters; the [global log option](https://caddyserver.com/docs/caddyfile/options#log)
applies to runtime logging.

Before any real HLS lease, use only synthetic unknown paths/tickets to check
normal responses, malformed requests and forced downstream aborts. Inspect the
actual access output and Caddy journal locally: no raw `/api/media/hls/<id>/`,
query, credential/header or credential-bearing Referer may remain. Record only
the sanitized conclusion. Test errors as well as successful responses after
enablement. Until both log paths pass, keep valid-credential media acceptance
pending. Do not enable global debug/body/header logging for diagnosis.

For LL-HLS additionally establish real client-facing HTTP/2 or HTTP/3 and path
p95 RTT at or below 333 ms. Backend HTTP/1.1 is a separate hop and does not prove
or disprove the public protocol gate. Conventional HLS can be tested first.
