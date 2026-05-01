# n.eko — Design Overhaul Fork

> **Fork von:** [m1k1o/neko](https://github.com/m1k1o/neko) — selbst gehosteter virtueller Browser via Docker + WebRTC.
> **Dieses Repo:** `officialtaxz2/neko` — eigenständiger Fork mit vollständigem UI/UX-Redesign.
> **Upstream-PRs:** Kein komplettes Design-Overhaul ins Upstream-Repo — dieser Fork lebt als eigenständiges Produkt.

---

## Kontext

Das ursprüngliche neko-Frontend war ein **1:1-Discord-Klon**: Farbpalette `#36393f / #2f3136`, Discord-exklusive Font „Whitney", keine eigene Designsprache. Ziel dieses Forks: **komplett eigenständiges Design** auf Basis eines modernen Design-Systems — Dark/Light-Mode, Fluid Typography, CSS Custom Properties, Glassmorphism und sauberer Komponentenhierarchie.

---

## Dev-Umgebung

```bash
# Repo klonen + Backend starten
git clone https://github.com/officialtaxz2/neko.git && cd neko
docker compose up -d

# Frontend Dev-Server (Hot Reload, kein Node auf Host)
cd client/dev && ./serve -i   # → http://localhost:3001

# Fertiges Image bauen
./build -r my-neko
# Nur Firefox (schneller):
./build -a firefox -b ghcr.io/m1k1o/neko/base:latest -r my-neko
docker compose up --force-recreate
```

---

## Implementierter Stand

Design Overhaul **Phase 1–5 vollständig abgeschlossen** ✅ · **R1–R7 vollständig abgeschlossen** ✅ · **Bugfix-Session B1–B6 vollständig abgeschlossen** ✅ (01.05.2026)

| Datei | Wichtigste Änderungen | Status |
|---|---|---|
| `_variables.scss` | Discord-Palette entfernt; Blue-Akzent `hsl(213,90%,62/44%)` Dark/Light; Fluid Type `clamp()`; 8pt-Grid; Shadow/Radius/Transition-Tokens; **R4:** `$controls-height` 125px → 64px (Avatar-Row entfernt) | ✅ |
| `main.scss` | Fontshare CDN (Satoshi + Cabinet Grotesk); globale Resets; `:focus-visible`; `prefers-reduced-motion`; **B3:** `@import vendor/notification` ergänzt; `vendor/font-whitney` auskommentiert (durch Fontshare ersetzt) | ✅ |
| `app.vue` | Theme-Toggle (`data-theme`); Sidebar-Transition slide+fade; Split-Screen Blue-Glow; room-container Glassmorphism; Sticky Header + Controls (`position:sticky`); `.is-scrolled`-Elevation; **R4:** `neko-members` aus Controls-Bar entfernt; **Bugfix:** `.expanded` + `neko-side v-if` an `connected && side` gebunden; **R7:** `<neko-connect>` entfernt; `<transition name="page-fade" mode="out-in">` mit zwei exklusiven Branches | ✅ |
| `header.vue` | Theme-Toggle-Button (SVG); Icon Micro-Animations (`scale`); funktionierender Glassmorphism-Diagonal-Gradient; CSS Tokens durchgehend | ✅ |
| `side.vue` | Glassmorphism (`backdrop-filter:blur(12px)`); Pill-Tabs; Tab-Inhalt-Transition `out-in`; **R1+R3:** 4-Tab-Leiste; Toggle-Logik (Chat+Users unabhängig, Settings exklusiv); 50/50-Split mit `panel-divider`; **B5:** Tab-Reihenfolge Users → Chat (Users links); **B6:** `activeUsers = true` — beide Panels standardmäßig offen nach Login | ✅ |
| `userlist.vue` | **R2:** Card-Layout (Avatar 36px links; Name + Crown; Buttons dauerhaft sichtbar); alle Aktionen inline; Skeleton Loading; **B2:** Cancel-Bug `result?.isConfirmed`-Guard auf Kick/Ban/toggleMute | ✅ |
| `context.vue` | Right-Click-Kontextmenü; Avatar + Name im Header; alle Admin-Aktionen (Kick/Ban/Mute/Unmute/Give/Release/Take); **B1:** Cancel-Bug `result?.isConfirmed`-Guard auf alle 4 `$swal`-Calls; SCSS vollständig auf CSS Custom Properties migriert | ✅ |
| `chat.vue` | Pill-Username-Badges; Message-Hover; Textarea Token-Focus; Skeleton Loading (4 Shimmer-Messages) | ✅ |
| `members.vue` | 4-Zustand-Status-Dots; Avatar Hover-Scale; Skeleton Loading — **R4:** nicht mehr in Controls-Bar gerendert, Datei bleibt erhalten | ✅ |
| `menu.vue` | Icons auf `--color-text-muted`; `<select>` vollständig auf CSS Custom Properties migriert | ✅ |
| `controls.vue` | Touch-Targets ≥44px; Micro-Animations; `.faded` → `--color-text-muted`; Animated Counters FPS/Bitrate | ✅ |
| `settings.vue` | Bento Grid (2-Spalten); Custom Toggle Switches; LOG OUT Ghost-Button; **R6:** „Display"-Card mit Resolution-`<select>` | ✅ |
| `video.vue` | WebRTC-Video + Event-Handler unberührt; **R5:** `fa-desktop`-Overlay-Button + `<neko-resolution>` entfernt | ✅ |
| `login.vue` *(neu, R7)* | Vollbild-Login-Page; Glassmorphism-Card; Spotlight-Gradient + Dot-Grid; semantisches HTML; Inline-Fehleranzeige; `SystemDialog`-Overlay; `autoPassword`/`autoUser` URL-Params | ✅ |
| `connect.vue` *(entfernt, R7)* | Vollständig durch `login.vue` ersetzt und gelöscht | ✅ |
| `vendor/_swal.scss` | **B1:** Komplett-Rewrite — alle `$background-*` Legacy-Variablen → CSS Custom Properties; Post-import-Override-Block für Runtime-Token-Support (Dark/Light); Icon-Farben (`warning`/`error`/`success`/`info`) über Design-System-Tokens; Backdrop `blur(4px)` | ✅ |
| `vendor/_notification.scss` *(neu, B3)* | **B3:** Neue Datei — floatendes Card-Design für vue-notification; ersetzt das alte farbige Banner-Layout; Icon per FA `::before` je Typ (warn/error/success/info); 3px accent-border-left statt Legacy-Vollfarb-Hintergrund; vollständig token-basiert | ✅ |
| `store/client.ts` | `systemDialog`-State + `setSystemDialog`-Mutation | ✅ |
| `neko/index.ts` | `$swal`-Calls (`logout`, `DISCONNECT`, `ERROR`) → `$accessor.client.setSystemDialog()` | ✅ |
| `neko/base.ts` | Public getter `peerConnection` auf `BaseClient` | ✅ |
| `client/src/locale/*.ts` | **R1 i18n:** `side.users` Key in allen 15 Locale-Dateien ergänzt | ✅ |

---

## Bugfix-Session (01.05.2026)

| # | Bug / Verbesserung | Betroffene Dateien | Status |
|---|---|---|---|
| B1 | Cancel-Bug: Kick/Ban/Mute/Unmute führten Aktion aus, wenn Dialog per ESC oder Cancel geschlossen wurde | `context.vue`, `userlist.vue` | ✅ |
| B2 | `result?.isConfirmed`-Guard in `userlist.vue` für Kick/Ban/toggleMute fehlte | `userlist.vue` | ✅ |
| B3 | `_notification.scss` fehlte komplett — vue-notification hatte kein Design-Override | `vendor/_notification.scss` *(neu)*, `main.scss` | ✅ |
| B4 | `_swal.scss` nutzte veraltete `$background-*` SCSS-Vars statt CSS Custom Properties → Dark/Light-Mode-Tokens wirkten nicht zur Laufzeit | `vendor/_swal.scss` | ✅ |
| B5 | Tab-Reihenfolge: Chat war links, Users rechts — gewünschte Reihenfolge: Users links, Chat rechts | `side.vue` | ✅ |
| B6 | Default-Zustand nach Login: nur Chat aktiv — gewünscht: beide Panels (Users + Chat) standardmäßig offen | `side.vue` | ✅ |

### B1 / B2 — Cancel-Bug (`context.vue` + `userlist.vue`)

SweetAlert2 v11 gibt bei Cancel/ESC kein truthy `value` zurück, sondern `{ isDismissed: true }`. Der alte Code prüfte `if (value)` — dadurch wurde die Aktion trotzdem ausgeführt.

**Fix:** Alle 4 `$swal`-Calls in `context.vue` (Kick, Ban, Mute, Unmute) sowie Kick, Ban, toggleMute in `userlist.vue` verwenden jetzt `result?.isConfirmed`:

```ts
// Vorher (buggy)
const value = await this.$swal({ ... })
if (value) this.$accessor.user.kick(member)

// Nachher (korrekt)
const result = await this.$swal({ ... })
if (result?.isConfirmed) this.$accessor.user.kick(member)
```

### B3 — `_notification.scss` (neu)

Neue Vendor-Override-Datei für vue-notification. Ersetzt das alte farbige Banner-Layout durch floating Cards:

- `background: var(--color-surface)` + `border: 1px solid var(--color-border)` + `var(--shadow-md)`
- `backdrop-filter: blur(8px)` für Glassmorphism-Konsistenz
- Typ-spezifische 3px-`border-left` (warn/error/success/info) + FA-Icon per `::before`
- `main.scss` importiert `vendor/notification`

### B4 — `_swal.scss` Rewrite

SweetAlert2 kompiliert SCSS-Variablen zu statischen Werten — CSS Custom Properties, die nach dem Import gesetzt werden, greifen nicht automatisch. Lösung: Post-import-Block mit `!important`-Overrides direkt auf `.swal2-popup`, `.swal2-title`, `.swal2-confirm` etc., sodass Design-Token-Werte zur Laufzeit korrekt angewendet werden.

### B5 / B6 — Tab-Reihenfolge + Default-State (`side.vue`)

- **B5:** Users-`<li>` im Template vor Chat-`<li>` verschoben → Users-Tab erscheint links
- **B6:** `activeUsers = true` (war `false`) → nach Login sind beide Panels sofort aktiv; der `panel-divider` + Split-Layout greifen direkt

---

## Roadmap

> ⬜ = Geplant · 🔄 = In Arbeit · ✅ = Fertig

### Übersicht

| # | Feature | Betroffene Dateien | Status |
|---|---|---|---|
| R1 | Sidebar: 3. Tab „Users" ergänzen + Toggle-Logik | `side.vue`, `client/src/locale/*.ts` (15 Dateien) | ✅ |
| R2 | Neue `userlist.vue` — User-Einträge + Moderationsaktionen | `userlist.vue` (neu), `side.vue` (Placeholder ersetzt) | ✅ |
| R3 | Multiple-Choice Sidebar: Chat + Users gleichzeitig (50/50) | `side.vue` — **in R1-Commit enthalten** | ✅ |
| R4 | User-Avatare aus Controls-Bar entfernen + Höhe reduzieren | `app.vue`, `_variables.scss` | ✅ |
| R5 | Resolution-Button aus Video-Player entfernen | `video.vue` | ✅ |
| R6 | Resolution-Dropdown in Settings ergänzen | `settings.vue` | ✅ |
| R7 | Dedizierte Login-Page — Vollbild-Route, Room-DOM nur bei `connected` | `app.vue` (patch), `login.vue` (neu), `connect.vue` (gelöscht) | ✅ |
| R8 | *(Platzhalter — noch zu definieren)* | — | ⬜ |
| R9 | *(Platzhalter — noch zu definieren)* | — | ⬜ |

---

### R1 + R3 — Sidebar Redesign ✅

**Tab-Leiste:** `fa-users` Users · `fa-comment-alt` Chat · `fa-file` Files (conditional) · `fa-sliders-h` Settings

**Tab-Logik:**

| Zustand | Sidebar-Inhalt |
|---|---|
| Nur Users aktiv | Users-Panel nimmt 100% der Sidebar-Höhe |
| Nur Chat aktiv | Chat nimmt 100% der Sidebar-Höhe |
| Users + Chat aktiv *(Standard nach Login)* | Users oben / Chat unten — je `flex: 1`; `panel-divider` dazwischen |
| Files aktiv | Files-Panel exklusiv (store-managed); Settings-State wird gecleart |
| Settings aktiv | Settings exklusiv; Chat+Users deaktiviert; beim Verlassen wird `activeChat = true` wiederhergestellt |

**Guards:**
- Letztes aktives Panel (Users oder Chat) kann nicht abgewählt werden
- Files-Aktivierung cleart `activeSettings`; Chat/Users-State bleibt für Rückkehr erhalten

**Transitions:**
- `panel-from-top` / `panel-from-bottom` (transform+opacity, GPU-composited) für Users + Chat
- `tab-fade out-in` für Files + Settings (exklusiv)

**i18n:** `side.users` in allen 15 Locale-Dateien ergänzt.

---

### R2 — `userlist.vue` ✅

Jeder User-Eintrag ist als **Card** aufgebaut:

```
┌─────────────────────────────────────────────┐
│ [Avatar 36px]  Username  [Crown-Icon?]      │  ← Zeile 1: Name-Row
│               [Btn][Btn][Btn][Btn][Btn]     │  ← Zeile 2: Actions (immer sichtbar)
└─────────────────────────────────────────────┘
```

#### Aktions-Buttons (Sichtbarkeits-Logik)

| Button | Icon | Bedingung | Store-Aktion |
|---|---|---|---|
| Ignore / Unignore | `fa-eye-slash` / `fa-eye` | Alle User | `user.setIgnored()` |
| Mute / Unmute | `fa-microphone-slash` / `fa-microphone` | `isAdmin` | `user.mute()` / `user.unmute()` |
| Give Controls | `fa-gamepad` | `!implicitHosting && m.id !== host && (isAdmin \| isCurrentHost)` | `remote.adminGive()` / `remote.give()` |
| Release Controls | `fa-times-circle` | `isAdmin && !implicitHosting && m.id === host` | `remote.adminRelease()` |
| Force Take Controls | `fa-hand-paper` | `isAdmin && !implicitHosting && m.id === host` | `remote.adminControl()` |
| Kick | `fa-user-times` | `isAdmin && !m.admin` | `user.kick()` |
| Ban IP | `fa-ban` | `isAdmin && !m.admin` | `user.ban()` |

---

### R4 — User-Avatare aus Controls-Bar entfernen ✅

- `app.vue`: `<neko-members />` + Import + Registrierung entfernt [cleanup]
- `_variables.scss`: `$controls-height` 125px → **64px**
- `members.vue` bleibt im Codebase erhalten (nicht mehr gerendert)

---

### R5 + R6 — Resolution aus Video-Player → Settings ✅

**R5 — `video.vue`:** `fa-desktop`-Button + `<neko-resolution>` + alle zugehörigen Imports/Refs entfernt.

**R6 — `settings.vue`:** Neue „Display"-Card (admin-only) mit `<select>` (`W × H @ R fps`); `resValue` getter/setter → `$accessor.video.screenSet()`.

---

### R7 — Dedizierte Login-Page ✅

| Szenario | Verhalten |
|---|---|
| `!connected` | Nur `<neko-login-page>` im DOM — kein Room, kein Header, keine Controls |
| Login erfolgreich | `page-fade` (300ms) Login → Room |
| `?pwd=xxx&usr=yyy` URL-Params | Auto-Login via `mounted()` |
| `prefers-reduced-motion` | `page-fade` sofort; Bounce-Loader statisch |

---

## Design-System Tokens

```css
/* Art direction: Dark-first tech tool → deep cool-gray surfaces, blue accent */

/* === DARK MODE (Standard) === */
--color-bg:                hsl(220, 15%, 9%);
--color-surface:           hsl(220, 13%, 13%);
--color-surface-2:         hsl(220, 11%, 16%);
--color-surface-offset:    hsl(220, 10%, 20%);
--color-surface-dynamic:   hsl(220, 9%, 24%);
--color-primary:           hsl(213, 90%, 62%);
--color-primary-hover:     hsl(213, 90%, 52%);
--color-primary-active:    hsl(213, 90%, 42%);
--color-primary-highlight: hsl(213, 35%, 18%);
--color-text:              hsl(220, 10%, 86%);
--color-text-muted:        hsl(220, 8%, 50%);
--color-text-faint:        hsl(220, 7%, 34%);
--color-border:            hsl(220, 9%, 23%);
--color-error:             hsl(4, 68%, 52%);
--color-success:           hsl(142, 50%, 48%);
--color-warning:           hsl(36, 92%, 52%);

/* === LIGHT MODE ([data-theme="light"]) === */
--color-bg:                hsl(220, 20%, 97%);
--color-surface:           hsl(220, 18%, 100%);
--color-border:            hsl(220, 10%, 84%);
--color-text:              hsl(220, 20%, 12%);
--color-text-muted:        hsl(220, 12%, 44%);
--color-primary:           hsl(213, 90%, 44%);

/* === FONTS === */
--font-display:  'Cabinet Grotesk', 'Helvetica Neue', sans-serif;
--font-body:     'Satoshi', 'Inter', 'Helvetica Neue', sans-serif;
--font-mono:     'JetBrains Mono', 'Fira Code', monospace;

/* === FLUID TYPE === */
--text-xs:   clamp(0.75rem,  0.70rem + 0.25vw, 0.875rem);
--text-sm:   clamp(0.875rem, 0.80rem + 0.35vw, 1rem);
--text-base: clamp(1rem,     0.95rem + 0.25vw, 1.125rem);
--text-lg:   clamp(1.125rem, 1.00rem + 0.75vw, 1.5rem);
--text-xl:   clamp(1.5rem,   1.20rem + 1.25vw, 2.25rem);

/* === LAYOUT === */
$menu-height:     40px;
$controls-height: 64px;   /* R4: reduziert von 125px */
$side-width:      400px;

/* === SPACING (8pt) === */
--space-1: 0.25rem;  --space-2: 0.5rem;   --space-3: 0.75rem;
--space-4: 1rem;     --space-6: 1.5rem;   --space-8: 2rem;
--space-12: 3rem;    --space-16: 4rem;

/* === SHADOWS / RADIUS / TRANSITIONS === */
--shadow-sm: 0 1px 2px rgba(0,0,0,0.35);
--shadow-md: 0 4px 12px rgba(0,0,0,0.45);
--shadow-lg: 0 12px 32px rgba(0,0,0,0.60);

--radius-sm: 0.25rem;  --radius-md: 0.5rem;
--radius-lg: 0.75rem;  --radius-xl: 1rem;  --radius-full: 9999px;

--transition-fast:        100ms cubic-bezier(0.16, 1, 0.3, 1);
--transition-interactive: 180ms cubic-bezier(0.16, 1, 0.3, 1);
--transition-slow:        300ms cubic-bezier(0.16, 1, 0.3, 1);
```

---

## Abnahme-Checkliste

```
Design & UX:
  [ ] Alle Views responsive (320px – 2560px) / Mobile First
  [ ] Dark Mode korrekt · Light Mode: Toggle-Tracks, Icons, Selects auf Kontrast prüfen
  [ ] Kontrast ≥ 4.5:1 (WCAG AA) · Focus-States sichtbar · Touch-Targets ≥ 44×44px
  [ ] prefers-reduced-motion implementiert
  [ ] Kein hardcodierter Farbwert (kein color:white, kein #36393f, keine alten $background-*-Vars)
```

---

## Upstream

Original: [m1k1o/neko](https://github.com/m1k1o/neko) · Docs: [neko.m1k1o.net](https://neko.m1k1o.net/)
