#!/usr/bin/env python3
"""Target-only, offline Caddyfile merge with an exact adapted-JSON guard.

Raw configurations stay in the operator's private evidence directory. This is
a bounded source editor, not a replacement Caddy parser. Caddy must adapt and
validate the candidate, and verify() rejects every unexpected semantic change.
"""

import copy
import json
import pathlib
import re
import sys
from dataclasses import dataclass


HOST = "neko.taxzvps.de"
REGEXP = "^(/[^?]*)?/api/media/hls.*$"
REDACTED = "/api/media/hls/:redacted"
FORMAT = '''format filter {
    fields {
        request>uri regexp "^(/[^?]*)?/api/media/hls.*$" "/api/media/hls/:redacted"
        uri regexp "^(/[^?]*)?/api/media/hls.*$" "/api/media/hls/:redacted"
        request>headers delete
        resp_headers delete
    }
    wrap json
}'''


def require(condition, message):
    if not condition:
        raise ValueError(message)


def objects(value):
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from objects(child)
    elif isinstance(value, list):
        for child in value:
            yield from objects(child)


def reviewed_proxy(config):
    servers = config.get("apps", {}).get("http", {}).get("servers", {})
    matches = []
    for server in servers.values():
        for route in objects(server.get("routes", [])):
            hosts = {
                host for matcher in route.get("match", []) if isinstance(matcher, dict)
                for host in matcher.get("host", [])
            }
            if HOST in hosts:
                require(hosts == {HOST}, "Neko shares a host route; review required")
                matches.append((server, route))
    require(len(matches) == 1, "expected one exact Neko host route")
    server, route = matches[0]
    logs = server.get("logs", {})
    names = logs.get("logger_names", {})
    require(not logs.get("default_logger_name"), "shared fallback logger requires review")
    logger_hosts = {HOST}
    labels = HOST.split(".")
    for index in range(len(labels)):
        labels[index] = "*"
        logger_hosts.add(".".join(labels))
    require(not logger_hosts.intersection(names), "Neko logger association requires review")
    require(not logs or HOST in logs.get("skip_hosts", []) or logs.get("skip_unmapped_hosts"),
            "Neko access logging requires separate redaction review")
    require(not logs.get("should_log_credentials"), "credential logging must stay disabled")
    proxies = [node for node in objects(route.get("handle", []))
               if node.get("handler") == "reverse_proxy"]
    require(len(proxies) == 1, "expected one Neko reverse proxy")
    proxy = proxies[0]
    require(proxy.get("upstreams") == [{"dial": "127.0.0.1:8082"}],
            "Neko upstream differs from the reviewed loopback endpoint")
    require("transport" not in proxy and "headers" not in proxy,
            "existing Neko transport/header options require review")
    require("request_buffers" not in proxy and "response_buffers" not in proxy,
            "explicit Neko buffering requires review")
    return proxy


def expected_config(baseline):
    expected = copy.deepcopy(baseline)
    proxy = reviewed_proxy(expected)
    proxy["headers"] = {"request": {"delete": ["Forwarded"]}}
    proxy["transport"] = {"protocol": "http", "compression": False}
    logs = expected.get("logging", {}).get("logs", {})
    require("default" in logs, "expected the inspected runtime default logger")
    default = logs["default"]
    require(not default.get("encoder"), "existing default encoder requires a reviewed merge")
    require(default.get("level", "INFO").upper() != "DEBUG", "debug logging requires review")
    for name, logger in logs.items():
        if name != "default":
            includes = logger.get("include", [])
            require(includes and all(namespace.startswith("http.log.access.") for namespace in includes),
                    "additional runtime/error logger requires separate redaction review")
    # Keep all default writer/level/include/exclude options, and all other logs.
    default["encoder"] = {
        "format": "filter",
        "wrap": {"format": "json"},
        "fields": {
            "request>uri": {"filter": "regexp", "regexp": REGEXP, "value": REDACTED},
            "uri": {"filter": "regexp", "regexp": REGEXP, "value": REDACTED},
            "request>headers": {"filter": "delete"},
            "resp_headers": {"filter": "delete"},
        },
    }
    return expected


@dataclass
class Token:
    value: str
    start: int
    end: int
    line: int
    quoted: bool = False


@dataclass
class Statement:
    head: list
    opening: object = None
    closing: object = None
    children: object = None


def statements(source):
    # Braces are structural only when they are standalone unquoted tokens.
    # Thus quoted braces and Caddy placeholders such as {$ENV} stay literal.
    pattern = re.compile(r'\s+|\#[^\n]*|"(?:\\.|[^"\\])*"|`[^`]*`|[^\s]+')
    tokens = []
    line = 1
    for match in pattern.finditer(source):
        raw = match.group()
        if not raw.isspace() and not raw.startswith("#"):
            quoted = raw[0] in ('"', '`')
            require(quoted or "\\" not in raw, "source line continuation/escape requires review")
            tokens.append(Token(raw[1:-1] if quoted else raw,
                                match.start(), match.end(), line, quoted))
        line += raw.count("\n")

    def structural(token, value):
        return not token.quoted and token.value == value

    def parse(index, nested=False):
        result = []
        while index < len(tokens):
            if structural(tokens[index], "}"):
                require(nested, "unfamiliar source block structure")
                return result, index + 1, tokens[index]
            head = []
            opening = closing = children = None
            while index < len(tokens):
                token = tokens[index]
                if structural(token, "{"):
                    opening = token
                    children, index, closing = parse(index + 1, True)
                    break
                if structural(token, "}") or (head and token.line != head[-1].line):
                    break
                head.append(token)
                index += 1
            require(head or opening, "unfamiliar source statement")
            result.append(Statement(head, opening, closing, children))
        require(not nested, "unclosed source block")
        return result, index, None

    return parse(0)[0]


def values(statement):
    return [token.value for token in statement.head]


def write_candidate(directory):
    baseline = json.loads((directory / "Caddyfile.before-merge.json").read_text())
    expected_config(baseline)  # Refuse unreviewed semantics before source edits.
    source = (directory / "Caddyfile.merge-source").read_bytes().decode("utf-8")
    require(not source.startswith("\ufeff"), "source byte-order mark requires review")
    roots = statements(source)
    sites = [node for node in roots if values(node) == [HOST] and node.opening]
    require(len(sites) == 1, "Neko site must be explicit in the active file; imports/aliases require review")
    proxies = [node for node in sites[0].children if values(node) and values(node)[0] == "reverse_proxy"]
    require(len(proxies) == 1 and values(proxies[0]) == ["reverse_proxy", "127.0.0.1:8082"]
            and proxies[0].opening is None, "Neko source proxy differs from the reviewed bare directive")
    proxy = proxies[0]
    patches = [(proxy.head[0].start, proxy.head[-1].end,
                "reverse_proxy 127.0.0.1:8082 {\n"
                "        header_up -Forwarded\n"
                "        transport http {\n"
                "            compression off\n"
                "        }\n"
                "    }")]
    globals_ = [node for node in roots if not node.head and node.opening]
    require(len(globals_) <= 1, "multiple global source blocks require review")
    default_logs = []
    if globals_:
        default_logs = [node for node in globals_[0].children
                        if values(node) in (["log"], ["log", "default"])]
    require(len(default_logs) <= 1, "multiple default source logger blocks require review")
    if default_logs:
        logger = default_logs[0]
        require(logger.opening is not None, "default source logger must have a block")
        require(not any(values(node) and values(node)[0] == "format" for node in logger.children),
                "existing source default format requires review")
        position = logger.closing.start
        addition = "\n" + FORMAT + "\n"
    elif globals_:
        position = globals_[0].closing.start
        addition = "\nlog default {\n" + FORMAT + "\n}\n"
    else:
        position = 0
        addition = "{\nlog default {\n" + FORMAT + "\n}\n}\n\n"
    patches.append((position, position, addition))
    for start, end, replacement in sorted(patches, reverse=True):
        source = source[:start] + replacement + source[end:]
    (directory / "Caddyfile.hls").write_bytes(source.encode("utf-8"))
    print("PASS candidate source: existing sites/options retained; bounded Neko/default-log additions prepared")


def verify(directory):
    baseline = json.loads((directory / "Caddyfile.before-merge.json").read_text())
    candidate = json.loads((directory / "Caddyfile.hls.json").read_text())
    require(candidate == expected_config(baseline),
            "adapted candidate differs outside the approved proxy/default-encoder changes; no service change")
    hosts = {
        host for server in baseline["apps"]["http"]["servers"].values()
        for route in objects(server.get("routes", []))
        for matcher in route.get("match", []) if isinstance(matcher, dict)
        for host in matcher.get("host", [])
    }
    result = {"merge": "passed", "explicit_host_count": len(hosts),
              "other_configuration_preserved": True, "neko_access_logger_added": False}
    (directory / "caddy-merge-check.json").write_text(json.dumps(result, indent=2) + "\n")
    print("PASS adapted configuration: " + str(len(hosts)) + " hosts retained; all other configuration equal")


def main():
    require(len(sys.argv) == 3 and sys.argv[1] in ("write", "verify"),
            "usage: merge-hls-caddy.py write|verify PRIVATE_OUTPUT_DIR")
    directory = pathlib.Path(sys.argv[2])
    if sys.argv[1] == "write":
        write_candidate(directory)
    else:
        verify(directory)


if __name__ == "__main__":
    try:
        main()
    except json.JSONDecodeError:
        sys.exit("FAIL merge: invalid adapted JSON; inspect private evidence")
    except (OSError, KeyError, TypeError, AttributeError, UnicodeError):
        sys.exit("FAIL merge: unreadable or unfamiliar configuration; inspect private evidence")
    except ValueError as error:
        # All explicit errors are fixed messages, never configuration excerpts.
        sys.exit("FAIL merge: " + str(error))
