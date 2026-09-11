#!/usr/bin/env python3
"""Exercise the deployed view-only HTTP boundary without printing credentials."""

from __future__ import annotations

import json
import os
import secrets
import sys
import urllib.error
import urllib.request


BASE_URL = os.environ.get("NEKO_BASE_URL", "http://127.0.0.1:8082").rstrip("/")
VIEW_ONLY_TOKEN = os.environ.get("NEKO_VIEW_ONLY_TOKEN", "")


def request(
    method: str,
    path: str,
    *,
    session_token: str | None = None,
    payload: object | None = None,
) -> tuple[int, bytes]:
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    headers = {"Accept": "application/json"}
    if data is not None:
        headers["Content-Type"] = "application/json"
    if session_token:
        headers["Authorization"] = f"Bearer {session_token}"

    req = urllib.request.Request(BASE_URL + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            return response.status, response.read()
    except urllib.error.HTTPError as error:
        return error.code, error.read()


def expect(name: str, actual: int, expected: int) -> None:
    if actual != expected:
        raise RuntimeError(f"{name}: HTTP {actual}, expected HTTP {expected}")
    print(f"PASS {name}: HTTP {actual}")


def main() -> int:
    if len(VIEW_ONLY_TOKEN) != 64 or any(char not in "0123456789abcdefABCDEF" for char in VIEW_ONLY_TOKEN):
        print("ERROR NEKO_VIEW_ONLY_TOKEN must contain exactly 64 hexadecimal characters", file=sys.stderr)
        return 2

    username = f"http-check-{secrets.token_hex(4)}"
    status, body = request(
        "POST",
        "/api/login",
        payload={"username": username, "password": VIEW_ONLY_TOKEN},
    )
    expect("view-only login", status, 200)

    session_token = ""
    try:
        login = json.loads(body)
        session_token = login["token"]
        profile = login["profile"]
    except (KeyError, TypeError, json.JSONDecodeError) as error:
        raise RuntimeError("view-only login returned an unexpected response shape") from error

    try:
        if not session_token:
            raise RuntimeError("view-only login returned no bearer session token; cookie auth must be disabled for this check")
        if not profile.get("is_view_only") or not profile.get("can_watch"):
            raise RuntimeError("login did not return a watch-capable view-only profile")

        denied_flags = ("is_admin", "can_host", "can_share_media", "can_access_clipboard", "sends_inactive_cursor")
        if any(profile.get(flag) for flag in denied_flags):
            raise RuntimeError("view-only profile contains an interactive capability")
        print("PASS normalized profile: passive marker present and interactive capabilities absent")

        allowed = [
            ("whoami", "GET", "/api/whoami", None),
            ("stats", "GET", "/api/stats", None),
            ("screen read", "GET", "/api/room/screen/", None),
        ]
        for name, method, path, payload in allowed:
            status, _ = request(method, path, session_token=session_token, payload=payload)
            expect(name, status, 200)

        denied = [
            ("profile mutation", "POST", "/api/profile", {"name": "forbidden"}),
            ("session API", "GET", "/api/sessions/", None),
            ("member API", "GET", "/api/members/", None),
            ("control request", "POST", "/api/room/control/request", None),
            ("keyboard API", "GET", "/api/room/keyboard/map", None),
            ("clipboard API", "GET", "/api/room/clipboard/", None),
            ("upload API", "POST", "/api/room/upload/dialog", None),
            ("chat plugin API", "POST", "/api/chat/", {"text": "view-only-denial-check"}),
            ("file-transfer API", "GET", "/api/filetransfer/?filename=__view_only_probe__", None),
            ("open-in-app API", "POST", "/api/openinapp/openlink", {"text": "https://example.invalid"}),
        ]
        for name, method, path, payload in denied:
            status, _ = request(method, path, session_token=session_token, payload=payload)
            expect(name, status, 403)
    finally:
        if session_token:
            status, _ = request("POST", "/api/logout", session_token=session_token)
            expect("probe logout", status, 200)

    print("PASS all view-only HTTP boundary checks")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as error:  # keep output useful without exposing response bodies or tokens
        print(f"FAIL {error}", file=sys.stderr)
        raise SystemExit(1)
