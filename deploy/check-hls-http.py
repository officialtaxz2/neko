#!/usr/bin/env python3
"""Credential-free target-server checks; no valid ticket or cookie is accepted."""

import argparse
import json
import os
import ssl
import sys
import urllib.error
import urllib.parse
import urllib.request

BOOTSTRAP = "/api/media/hls/session"
UNKNOWN_LEASE = "/api/media/hls/" + "A" * 22
UNKNOWN_TICKET = json.dumps({"ticket": "A" * 32}).encode()


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def configuration(mode):
    base = os.environ.get("NEKO_PUBLIC_BASE_URL", "").rstrip("/")
    origin = os.environ.get("NEKO_PUBLIC_ORIGIN", "")
    target = urllib.parse.urlsplit(base)
    allowed = urllib.parse.urlsplit(origin)
    for value in (target, allowed):
        if not value.hostname or value.username is not None or value.password is not None or value.query or value.fragment:
            raise ValueError("use credential-free base/origin URLs")
    if allowed.scheme != "https" or allowed.path or origin != allowed.geturl():
        raise ValueError("NEKO_PUBLIC_ORIGIN must be one exact HTTPS origin")
    if target.scheme not in ("http", "https") or base != target.geturl() or "%" in target.path or ".." in target.path or "//" in target.path:
        raise ValueError("invalid public base URL or path prefix")
    try:
        target.port, allowed.port
    except ValueError:
        raise ValueError("invalid URL port") from None
    if mode == "enabled" and (target.scheme != "https" or target.netloc != allowed.netloc):
        raise ValueError("enabled checks require same-origin public HTTPS")
    if mode == "insecure-denied" and (target.scheme != "http" or target.hostname not in ("localhost", "127.0.0.1", "::1")):
        raise ValueError("insecure-denied checks require a direct loopback HTTP base")
    return base, origin


def check(opener, base, name, expected, path, method="POST", body=UNKNOWN_TICKET, headers=None, secure_headers=True):
    try:
        request = urllib.request.Request(base + path, data=body, method=method, headers=headers or {})
        try:
            response = opener.open(request, timeout=15)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            actual = response.code
            fields = response.headers
            cache = {item.strip().lower() for item in fields.get("Cache-Control", "").split(",")}
            vary = {item.strip().lower() for item in fields.get("Vary", "").split(",")}
            safe = not fields.get("Access-Control-Allow-Origin") and not fields.get("Set-Cookie")
            if secure_headers:
                safe = safe and {"private", "no-store", "max-age=0"} <= cache and {"cookie", "accept-encoding"} <= vary
                safe = safe and fields.get("Pragma") == "no-cache"
                safe = safe and fields.get("Referrer-Policy") == "no-referrer" and fields.get("X-Content-Type-Options") == "nosniff"
        passed = actual == expected and (safe or not secure_headers)
        header_status = ("ok" if safe else "rejected") if secure_headers else "not-required"
        print(f"{'PASS' if passed else 'FAIL'} {name}: expected={expected} actual={actual} headers={header_status}")
        return passed
    except (OSError, ValueError, urllib.error.URLError):
        print(f"FAIL {name}: connection failure (details suppressed)")
        return False


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("disabled", "enabled", "insecure-denied"))
    mode = parser.parse_args().mode
    try:
        base, origin = configuration(mode)
    except ValueError as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    # Never inherit proxy settings, cookie jars, login handlers or redirects.
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect(), urllib.request.HTTPSHandler(context=ssl.create_default_context()))
    headers = {"Origin": origin, "Content-Type": "application/json", "Accept-Encoding": "gzip"}
    cases = []
    if mode == "disabled":
        cases = [("disabled bootstrap", 404, BOOTSTRAP, "POST", UNKNOWN_TICKET, headers),
                 ("disabled media", 404, UNKNOWN_LEASE + "/master.m3u8", "GET", None, headers)]
    elif mode == "insecure-denied":
        cases = [("cleartext bootstrap", 403, BOOTSTRAP, "POST", UNKNOWN_TICKET, headers),
                 ("cleartext media", 403, UNKNOWN_LEASE + "/master.m3u8", "GET", None, headers)]
    else:
        media = UNKNOWN_LEASE + "/master.m3u8"
        cases = [
            ("bootstrap query", 400, BOOTSTRAP + "?probe=1", "POST", UNKNOWN_TICKET, headers),
            ("missing origin", 403, BOOTSTRAP, "POST", UNKNOWN_TICKET, {"Content-Type": "application/json"}),
            ("wrong origin", 403, BOOTSTRAP, "POST", UNKNOWN_TICKET, {**headers, "Origin": "https://invalid.example"}),
            ("wrong content type", 400, BOOTSTRAP, "POST", UNKNOWN_TICKET, {**headers, "Content-Type": "text/plain"}),
            ("malformed JSON", 400, BOOTSTRAP, "POST", b"{", headers),
            ("unknown field", 400, BOOTSTRAP, "POST", json.dumps({"ticket": "A" * 32, "extra": True}).encode(), headers),
            ("malformed ticket", 400, BOOTSTRAP, "POST", b'{"ticket":"short"}', headers),
            ("unknown ticket", 401, BOOTSTRAP, "POST", UNKNOWN_TICKET, headers),
            ("oversized body", 413, BOOTSTRAP, "POST", b" " * 2049, headers),
            ("no cookie", 404, media, "GET", None, headers),
            ("native GET without origin", 404, media, "GET", None, {}),
            ("unknown lease/cookie", 404, media, "GET", None, {**headers, "Cookie": "__Secure-neko-hls=" + "A" * 32}),
            ("HEAD without cookie", 404, media, "HEAD", None, headers),
            ("keepalive without cookie", 404, UNKNOWN_LEASE + "/keepalive", "POST", b"", headers),
            ("cross-origin media", 403, media, "GET", None, {**headers, "Origin": "https://invalid.example"}),
            ("cross-site media", 403, media, "GET", None, {**headers, "Sec-Fetch-Site": "cross-site"}),
            ("malformed lease path", 400, "/api/media/hls/short/master.m3u8", "GET", None, headers),
        ]
    results = [check(opener, base, *case, secure_headers=mode != "disabled") for case in cases]
    print(f"Checks: {sum(results)}/{len(results)} passed; valid-credential/packager cases remain separate.")
    return 0 if all(results) else 1


if __name__ == "__main__":
    sys.exit(main())
