# `MyNekoProjekt` Local Delta Audit

Completed: 2026-09-09.

This document records the completed, sanitized filesystem comparison between the public fork baseline and the provided local working/deployment tree. It deliberately contains no credential values, browser-profile contents, cookies, downloads, or other runtime payloads.

## Scope and safety

- Fork baseline: `d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6`.
- Safety branch: `safety-pre-local-delta-audit-20260909` at that baseline.
- Local reference root: `MyNekoProjekt/` in the supplied workspace.
- `.git/` was excluded from raw file comparison.
- During the local-delta audit phase, no upstream fetch, merge, rebase, cherry-pick, or source import was performed.
- No filesystem reparse points were present in the local reference.

The baseline was materialized as a clean filesystem tree from the exact Git commit. Both trees were enumerated recursively by relative path. Common files were compared using raw SHA-256; raw mismatches were compared again after replacing only CRLF with LF. This separates transport/checkout line-ending noise from content changes without ignoring other whitespace or bytes.

## Complete inventory result

| Result | Files | Classification |
|---|---:|---|
| Present in both, byte-identical | 138 | No difference |
| Present in both, CRLF/LF only | 585 | Generated/checkout representation; no import |
| Present in both, substantive | 1 | Desired deployment structure plus instance values/credentials; later imported in sanitized form |
| Local-only | 3,929 | Runtime data (3,928) plus instance policy file (1); excluded |
| Fork-only | 0 | Nothing missing from the local tree |
| **Fork baseline total** | **724** | All files accounted for |
| **Local reference total** | **4,653** | All files accounted for |

The 585 line-ending-only differences are exhaustively distributed as follows; every other common file is either one of the 138 byte-identical files or the single substantive compose difference.

| Path area | CRLF/LF-only files |
|---|---:|
| Root project files (`.editorconfig`, `.gitattributes`, `.gitignore`, `build`, `config.yml`, `Dockerfile.tmpl`, `LICENSE`, `neko.code-workspace`, `README.md`, `SECURITY.md`, `tsconfig.json`) | 11 |
| `.github/` | 14 |
| `.vscode/` | 2 |
| `apps/` | 77 |
| `client/` | 107 |
| `runtime/` | 13 |
| `server/` | 164 |
| `utils/` | 41 |
| `webpage/` | 156 |
| **Total** | **585** |

## Classification decisions

### Source feature/fix to import

None. After line-ending normalization, every common source, build, application-image, runtime, server, utility, and webpage file matches the fork baseline.

### Reusable config/example change to import

The initial audit classified the complete local compose file as instance-only because it combined deployment structure with credentials, host-specific ports and runtime mounts. The operator later clarified that this structure is the intended deployment baseline for this fork.

The repository compose now preserves the desired reusable behavior: local `my-neko/brave:latest` image selection, Brave singleton-lock cleanup, persistent profile/download mounts, optional managed-policy mount, loopback HTTP binding, the local UDP range and file transfer. Host paths and ports remain configurable. Passwords are mandatory `.env` inputs and no credential value was copied. The ignored runtime directories and instance policy contents remain outside Git.

### Instance-only deployment config and credential

`docker-compose.yaml` is the only substantive common-file difference. The local version selects a locally named Brave image, enables privileged/container capabilities, mounts the persistent browser profile, downloads and managed policy, maps loopback HTTP and a different UDP range, cleans Chromium/Brave singleton files, enables local file transfer, and embeds member/admin passwords. The raw file was never copied. Its operator-confirmed deployment structure was later reconstructed with parameterized paths/settings and required external password variables.

The local-only root `policy.json` is an empty instance policy mount target. It was not copied; the compose defaults to the reusable `apps/brave/policies.json` and allows an ignored instance policy path through `NEKO_POLICY_FILE`.

The first target-server smoke test found that the sanitized reconstruction had changed the original Brave policy destination from `/etc/brave/policies/managed/policies.json` to the singular filename `policy.json`. The plural destination from the audited local compose, Brave image and inherited browser documentation has been restored. The operator then confirmed that the external policy and persistent profile load as intended.

Any credentials present in the supplied deployment artifact must be considered exposed and rotated operationally if they are still active.

### Runtime data

All 3,928 local-only files under `files/**` are persistent Brave/Chromium profile data. The local `downloads/` directory was empty at audit time. Both areas remain outside Git. The repository root now ignores `MyNekoProjekt/`, `files/`, `downloads/`, and `policy.json` as an additional guard against accidental staging.

### Generated/vendor artifact

The 585 CRLF/LF-only mismatches are checkout/transfer representation differences. Because their normalized SHA-256 values match exactly, importing them would create noise without changing project content.

## Static review and resulting repair

Static manifest/lock inspection found a pre-existing mismatch: `client/package.json` had already migrated to Vite and TypeScript 5.8, while `client/package-lock.json` still described the removed Vue CLI/TypeScript 4 toolchain. Commit `2d89027e` regenerates the lockfile from the manifest, selects TypeScript module resolution compatible with the existing Vue 2 augmentations, and types the existing `cursor-position` event.

Static checks confirm that the lockfile root dependency and dev-dependency declarations now match `package.json`, the obsolete Vue CLI dependency graph is removed, and the existing runtime event name/payload matches its consumers in `video.vue`.

Target-server verification status: **OPERATOR-CONFIRMED PASSED on 2026-09-09** for the target deployment. The reported client commands were:

```bash
cd client
npm ci
npm run lint
npm run build
```

A server build was not required specifically for this earlier local reconciliation because no server, root build, runtime configuration, or application-image source changed in that unit. The later integrated-baseline verification covered the checks applicable to the target deployment; no universal support is claimed for unreported devices or architectures.

## Outcome

The source delta audit is closed: no missing application source change was found. The original all-or-nothing exclusion of the local compose was corrected after the operator identified it as the desired deployment baseline. Its reusable structure is now tracked without copying credentials, browser-profile contents, downloads or instance-policy data. The semantic upstream synchronization is recorded separately in [`UPSTREAM_SYNC_AUDIT.md`](UPSTREAM_SYNC_AUDIT.md); current work status and `NEXT` remain in [`WORKPLAN.md`](WORKPLAN.md#next).
