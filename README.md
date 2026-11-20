# officialtaxz2/neko

A customized fork of [m1k1o/neko](https://github.com/m1k1o/neko): a self-hosted virtual browser that several people use together over the web.

The fork keeps Neko's shared session and roles (one person controls, everyone watches; admins can lock and hand over control). It adds a redesigned client for desktop and phone, per-viewer adaptive quality, HLS as an alternative stream, view-only share links and a one-command deployment with Brave or Helium.

## What the fork adds

**Client** (upstream's React client with the fork's design in [`client/src/fork/`](client/src/fork/README.md))

- Redesigned interface for desktop, phone and tablet, with a control pill that says who has control.
- Touch input: trackpad mode with its own pointer, "force touch" and a bar of special keys above the on-screen keyboard.
- Playback watchdog: when the picture stops for eight seconds in the foreground, the stream is re-attached, up to three times.
- Open tabs notice a new deployment: they reload when that interrupts nothing, otherwise they offer a reload.

**Server and media**

- Per-viewer adaptive quality: four VP8 tiers, from `high` (1080p) down to `lowest` (540p). Each viewer's tier is chosen on its own: it drops a tier only when its bandwidth falls short and its receiver still reports loss after retransmission, and it probes back up one tier at a time. See [`docs/ADAPTIVE_QUALITY.md`](docs/ADAPTIVE_QUALITY.md).
- HLS and LL-HLS as a manual alternative to WebRTC for watching (H.264/AAC). WebRTC stays the default for everyone. See [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md).
- View-only share links (`#/<16-character token>`), enforced by the server, not just hidden in the UI. See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).
- A media layer that separates the shared capture from per-viewer delivery, so one slow viewer cannot block the others. See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

**Deployment**

- `setup.sh` installs and updates everything with one command and keeps the previous build for a rollback.
- Brave (default) or Helium, both configured close to a normal desktop browser.

## Deployment

You need Docker with Compose v2 and a reverse proxy with HTTPS in front of Neko (required for HLS). On the server:

```bash
git clone https://github.com/officialtaxz2/neko.git /opt/docker/neko
cd /opt/docker/neko
bash setup.sh
```

Every update is the same command:

```bash
git pull && bash setup.sh
```

`setup.sh` is idempotent. On each run it:

1. creates `.env` from `.env.example` on first run and prints the two generated passwords once; later runs only add new keys and never change existing values,
2. prepares the profile and download directories,
3. builds the base and browser images with the Docker cache,
4. checks the Compose configuration, starts it and waits for the health check,
5. keeps the images it replaced as `:previous` (for example `my-neko/brave:previous`) and removes the older ones.

| Option | Effect |
| --- | --- |
| `--check` | Runs the server and client tests in containers first (a few minutes) and stops before building if they fail. Use it after a pull that changed `server/` or `client/`. |
| `--no-cache` | Rebuilds without the Docker cache. This also installs the newest browser release. |
| `--no-build` | Skips the build and recreates the containers, so a changed policy file, `.env` value or other mounted file takes effect. The stream restarts briefly. |
| `--rollback` | Starts the previous build again without touching `.env`. The next run without it returns to the newest build. |

A plain `docker compose up -d` keeps a running container as it is, including the old policy file. Use `bash setup.sh --no-build` instead.

<details>
<summary>Manual steps without <code>setup.sh</code></summary>

```bash
./build my-neko/base:latest -y
./build my-neko/brave:latest -y   # or my-neko/helium:latest
cp .env.example .env              # set both passwords and the server-specific values
docker compose config --quiet
docker compose up -d
```

</details>

## Configuration

All settings live in `.env` (ignored by Git, mode `600`). The most important ones:

| Key | Meaning |
| --- | --- |
| `NEKO_BROWSER` | `brave` (default) or `helium`. Leave `NEKO_IMAGE` and `NEKO_POLICY_FILE` empty so they follow it. |
| `NEKO_MEMBER_MULTIUSER_USER_PASSWORD`, `..._ADMIN_PASSWORD` | Login passwords; generated on first run. If both are the same, everyone becomes admin. |
| `NEKO_HTTP_BIND`, `NEKO_HTTP_PORT` | Where the web interface listens (default `127.0.0.1:8082`, for a reverse proxy on the same host). |
| `NEKO_WEBRTC_EPR` | UDP port range for WebRTC (default `51000-51100`); open it in the firewall. |
| `NEKO_DESKTOP_SCREEN` | Resolution and frame rate of the shared desktop (default `1920x1080@30`). |
| `NEKO_PROFILE_DIR`, `NEKO_DOWNLOADS_DIR` | Browser profile and downloads on the host (default `./files`, `./downloads`). |
| `NEKO_POLICY_FILE` | Optional own policy file instead of `apps/<browser>/policies.json`. |
| `NEKO_VIEW_ONLY_TOKEN` | Enables view-only links. Generate one with `docker compose -f docker-compose.validation.yaml run --rm token-generator`; empty disables them. |
| `NEKO_WEBRTC_NAT1TO1` | Public server address, only if clients cannot reach the discovered one (also uncomment it in `docker-compose.yaml`). |
| `COMPOSE_FILE` | Which Compose files are used, see below. |

### Adaptive quality and HLS

Both are Compose overlays that this deployment enables by default:

```bash
COMPOSE_FILE=docker-compose.yaml:docker-compose.adaptive.yaml:docker-compose.hls.yaml
```

Because the list is in `.env`, plain `docker compose` commands (for example `docker compose logs neko`) use the overlays too. `COMPOSE_FILE=docker-compose.yaml` runs without them. HLS needs the adaptive overlay, because all its variants are encoded from the `high` tier.

HLS additionally needs:

- `NEKO_MEDIA_HLS_ALLOWED_ORIGINS`: the exact public `https://` origin.
- `NEKO_MEDIA_HLS_TRUSTED_PROXIES`: the reverse proxy's address as Neko sees it. For a proxy on the same host that is the Docker network gateway; `setup.sh` stops when either value is missing and suggests it.
- `NEKO_MEDIA_HLS_MODES`: `hls ll-hls` (default) or only `hls`. Conventional HLS runs about 20 seconds behind, LL-HLS about 3 to 5 seconds and needs HTTP/2 at the proxy (`deploy/check-ll-hls.sh` checks it).

If the running container uses overlays that `COMPOSE_FILE` no longer lists, `setup.sh` stops before building and prints the line to add. Reverse proxy setup: [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md).

## Browsers

Both browsers keep their profile (history, bookmarks, cookies, settings, extensions) across restarts, and extensions install from the Chrome Web Store as usual. The managed policy only blocks what would endanger the shared session:

- no developer tools (they would show every viewer the session cookies of logged-in sites),
- no password manager or autofill, no extra profiles or guest windows,
- no `file://` pages except the Downloads folder, no notifications,
- for Brave also no Brave Sync (its sync code is visible to every viewer).

Because these are managed policies, the browser shows "Managed by your organization"; that refers only to this file. Downloads land in `NEKO_DOWNLOADS_DIR`, which Neko's file transfer serves.

Stale profile locks are removed on start (`runtime/chromium-unlock`). An image built before #37 lacks this: if the browser then refuses to start, delete `Singleton{Lock,Socket,Cookie}` in `NEKO_PROFILE_DIR`.

### Brave (default)

`apps/brave` installs the current stable release from Brave's signed apt repository; `bash setup.sh --no-cache` updates it. The profile lives directly in `NEKO_PROFILE_DIR`, and Brave restores the open tabs on start ("Continue where you left off"; this needs `stopasgroup=true` in `apps/brave/supervisord.conf`, because `/usr/bin/brave-browser` does not forward the stop signal). Brave Shields blocks ads; no extension is preinstalled. Brave allows file upload dialogs; they show the container's files, so upload a file with Neko's file transfer first and then pick it from `~/Downloads`.

### Helium (optional)

[Helium](https://helium.computer) (`apps/helium`, amd64/arm64) comes with uBlock Origin built in. To switch, set `NEKO_BROWSER=helium` in `.env` and run `bash setup.sh`; set it back to `brave` to return. Helium keeps its profile in the `helium/` subdirectory of `NEKO_PROFILE_DIR`, so Brave's profile stays untouched.

Helium has no file upload dialogs. To install extensions, enable Helium services once under `helium://settings`, because the first-run page that normally offers this is skipped. Not yet verified: Widevine/DRM and whether the Openbox window rule matches Helium's window.

## Repository layout

```text
client/    React + TypeScript/Vite client (upstream), the fork's design in client/src/fork/
server/    Go server and plugins
apps/      browser and application images (apps/brave, apps/helium, ...)
runtime/   base runtime image
deploy/    adaptive-quality config, HLS proxy and check helpers
docs/      the fork's architecture, media and sync notes
webpage/   upstream's documentation site
setup.sh   install, update and rollback
```

Never commit `.env`, `files/`, `downloads/` or `policy.json`; they hold credentials and browser data and are ignored by Git.

## Documentation

| Document | Content |
| --- | --- |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Architecture, media delivery and view-only links |
| [`docs/ADAPTIVE_QUALITY.md`](docs/ADAPTIVE_QUALITY.md) | Adaptive quality: tiers, diagnostics, rollback |
| [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md) | HLS/LL-HLS design and reverse proxy setup |
| [`docs/UPSTREAM_SYNC_AUDIT.md`](docs/UPSTREAM_SYNC_AUDIT.md) | Record of every upstream merge and where the fork deviates |
| [`client/src/fork/README.md`](client/src/fork/README.md) | The fork's client code |
| [`AGENTS.md`](AGENTS.md) | Rules, invariants and open items for coding agents |

The fork follows upstream `m1k1o/neko` by merging its `master`, never by overwriting; each merge is reviewed and recorded in the sync audit.

## License

Apache 2.0, as upstream. See [`LICENSE`](LICENSE).
