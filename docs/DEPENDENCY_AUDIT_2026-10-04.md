# Target dependency audit and bounded applicability review

Evidence: the operator supplied the complete npm audit v2 report from exact
`e55bcd7e256c05ca0053d705879ca3adfd8c4e65` on 2026-10-04. The report contains
**20 affected package entries: 11 low, 3 moderate, 5 high and 1 critical**.
These are package counts, including propagated dependency findings, rather
than 20 independent vulnerabilities. Ten packages carry their own advisories;
the other ten inherit Vue findings. Versions below are from the tracked lock.
The raw audit and runtime evidence stay in the private result directory.

This is source/configuration classification, **NOT EXECUTED IN CODEX** for
tests, installation, bundle inspection, requests or exploitation. No packages
or lockfile versions are changed by this review, so the audit count is expected
to remain unchanged. Applicability is not a dependency-security clearance.

## Package inventory

| Package / locked version | Audit severity | Inspected use and disposition |
| --- | --- | --- |
| `@babel/runtime` 7.20.13 | moderate | Browser helpers via v-tooltip/vue-resize. Named-capture RegExp replacement advisory; no `wrapRegExp` use found in those inspected vendor outputs. Update pending; not dismissed solely because it is transitive. |
| `@types/vue-clickaway` 2.2.0 | low | Build types; inherited Vue finding, no separate advisory. |
| `@vitejs/plugin-vue2` 2.3.4 | low | Build plugin; inherited Vue finding. Neko production serves built assets, not the Vite development server. |
| `axios` 1.2.6 | high | Browser HTTP for fixed asset/About URLs and file operations. Its 31 own advisories cover Node HTTP/proxy/redirect/body handling, browser XSRF and configuration/header/prototype/form parsing. Node-specific paths are separated below; remaining configuration/prototype findings lack a demonstrated entry in the inspected fixed-config calls and remain update work. XSRF cookie auto-reading is disabled as bounded containment. |
| `braces` 3.0.2 | high | Build watcher/glob path via chokidar; resource/stack exhaustion on crafted patterns. Not a room-message handler; build dependency repair pending. |
| `follow-redirects` 1.15.2 | moderate | Axios Node adapter; redirect URL/auth advisories. Browser Axios does not use this Node adapter. Lock repair pending. |
| `form-data` 4.0.0 | critical | Axios Node platform; predictable multipart boundary and field/header escaping advisories. Browser uploads use native `FormData`. No deployed Node request path found; lock repair pending, not a demonstrated critical HLS-server exposure. |
| `immutable` 4.2.2 | high | Sass build dependency; prototype/hash/trie advisories. Not a deployed room data structure; build dependency repair pending. |
| `lodash` 4.17.21 | high | v-tooltip imports `merge` and `isEqual`. Reported `template` imports and `unset`/`omit` paths were not found in that inspected bundle or direct client calls. Update pending; no assertion about every Lodash use in future code. |
| `picomatch` 2.3.1 | high | Chokidar/anymatch/readdirp build globs; crafted-pattern injection/RegExp advisories. This is the reported 2.x installation, not the separate 4.x Vite installation. Build dependency repair pending. |
| `typed-vuex` 0.1.22 | low | Browser store wrapper; inherited Vue finding, no separate advisory. |
| `v-tooltip` 2.1.3 | low | Browser tooltips; inherited Vue finding. Separately confirmed unsafe HTML handling of member names is contained with `html: false`. |
| `vue` 2.7.14 | low | Browser framework plus runtime compiler through `vue/dist/vue.esm.js`. Participant-controlled chat reached runtime compilation; the repair removes that path. Static application templates remain build input. Vue 2 maintenance decision remains open. |
| `vue-class-component` 7.2.6 | low | Browser component wrapper; inherited Vue finding. |
| `vue-clickaway` 2.2.2 | low | Browser directive; inherited Vue finding. |
| `vue-context` 5.2.0 | low | Browser context menu; inherited Vue/clickaway finding. |
| `vue-notification` 1.3.20 | low | Browser notifications; inherited Vue finding. |
| `vue-resize` 1.0.1 | low | Browser resize component; inherited Vue finding. |
| `vue-template-compiler` 2.7.14 | moderate | Direct development dependency; prototype-pollution-dependent compiler XSS advisory. No direct source import found. Do not label the related runtime compiler safe merely because this package is dev-only; the unsafe chat compilation path is removed. Vue 2 maintenance remains open. |
| `vuex` 3.6.2 | low | Browser store; inherited Vue finding. Vuex 4 is not a routine Vue 2 patch. |

The report contains no `hls.js` entry. That only describes this report and does
not prove that the pinned player or the Go HLS implementation is vulnerability
free. npm audit does not audit the Go server, Caddy or browser image packages.

## Reachability evidence and repairs

**Node dependencies.** Axios's exact
[1.2.6 package configuration](https://raw.githubusercontent.com/axios/axios/v1.2.6/package.json)
maps its Node HTTP adapter to a null helper and the Node platform to the browser
platform for browser builds. The inspected client imports Axios normally;
`files.vue` creates native `FormData` and explicitly disables credentials for
uploads/downloads. Neko serves the built browser client through Go; no deployed
Node API/SSR process is configured. Therefore non-reachability of `form-data`
and `follow-redirects` in that browser path is a source-level conclusion,
pending confirmation against the generated bundle. The
[critical multipart advisory](https://github.com/advisories/GHSA-fjxv-7rqg-78g4)
requires an application actually using that multipart generator and an
observable/predictable boundary; the package being installed is not proof of
that runtime path.

**Browser Axios.** The
[XSRF advisory](https://github.com/advisories/GHSA-wf5p-g6vw-rhxx)
also concerns the browser. Exact installed 1.2.6 `lib/adapters/xhr.js` reads the
configured cookie when `withCredentials || isURLSameOrigin(fullPath)`, provided
`xsrfCookieName` is truthy. No Neko use of the Axios XSRF-cookie mechanism was
found in client/server source. `plugins/axios.ts` now sets that name to the empty
string when installing the shared instance, preventing this automatic cookie
read/header insertion, including the fixed cross-origin About requests. This
does not patch the other Axios advisories or replace an eventual reviewed
dependency update. Recheck file upload/download and About on the target.

**Chat compilation.** `markdown.ts` originally returned a component whose
`template` concatenated Markdown output from room messages. HTML escaping does
not neutralize Vue interpolation; this violates Vue's
[non-trusted template rule](https://v2.vuejs.org/v2/guide/security.html).
The component now renders a fixed `div` with escaped Markdown HTML rather than
compiling a message as Vue source. Custom emoji/link attributes are escaped;
emoji labels use native `title` instead of generated executable tooltip
directives. Sprite rendering, Markdown formatting, spoilers and delegated
open-in-app links remain represented in the output. This removes the inspected
participant-controlled compiler path associated with the
[Vue parseHTML risk](https://github.com/advisories/GHSA-5j4c-8p2g-v4jx).
It does not patch Vue globally or establish a fix for the
[prototype-dependent compiler advisory](https://github.com/advisories/GHSA-g3ch-rx76-35fx).

**Member tooltip.** v-tooltip 2.1.3 has
[`defaultHtml: true`](https://raw.githubusercontent.com/Akryum/v-tooltip/v2.1.3/src/directives/v-tooltip.js).
`members.vue` passed another participant's display name without overriding it.
The binding now explicitly uses `html: false`; member names remain literal text.
This application defect is separate from the propagated npm entry.

These are confirmed source defects/containment changes, not a reproduction or
diagnosis of the unavailable television's event-triggered disconnects. They
change no media transport, codec, estimator or bitrate.

## Required follow-up and release boundary

1. Pull the repair commit and rerun the exact Phase 4 automated/image helper in
   a new private result directory. The earlier e55 images do not contain these
   repairs; retain their evidence rather than reusing their pass marker.
2. Target `npm test` includes the new actual-parser/Vue-component regressions
   for literal interpolation, raw tags, hostile emoji/URL attributes and normal
   formatting. Type/build and the image gate remain required. No target result
   is claimed for this repair yet.
3. On available browsers, verify chat/code/emoji/spoilers and open-in-app links,
   literal member-name tooltips, file operations and About; inspect cross-origin
   requests locally for absence of automatic `X-XSRF-TOKEN`. Record conclusions
   without sharing cookies, URLs containing credentials or header dumps.
4. Prepare a separate reviewed lockfile refresh for Axios and the vulnerable
   transitive/build packages, check all report ranges against the then-current
   registry, inspect the resulting graph and run exact target checks. Resolve
   Vue 2/compiler maintenance separately. No `npm audit fix --force`, Vuex 4
   migration or suggested wrapper downgrade is authorized by an audit hint.
5. Bounded HLS compatibility testing can continue after the repair image and
   Caddy gates. **Final dependency-security acceptance and master promotion
   remain open** until the outstanding dependency decisions/repairs and
   grouped runtime evidence are resolved. Do not mark findings fixed merely
   because a currently inspected call path does not use the vulnerable API.
