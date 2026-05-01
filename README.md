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

Design Overhaul **Phase 1–5 vollständig abgeschlossen** ✅ · **R1–R7 vollständig abgeschlossen** ✅ · Bugfix-Session abgeschlossen ✅ (01.05.2026)

| Datei | Wichtigste Änderungen | Status |
|---|---|---|
| `_variables.scss` | Discord-Palette entfernt; Blue-Akzent `hsl(213,90%,62/44%)` Dark/Light; Fluid Type `clamp()`; 8pt-Grid; Shadow/Radius/Transition-Tokens; **R4:** `$controls-height` 125px → 64px (Avatar-Row entfernt) | ✅ |
| `main.scss` | Fontshare CDN (Satoshi + Cabinet Grotesk); globale Resets; `:focus-visible`; `prefers-reduced-motion` | ✅ |
| `app.vue` | Theme-Toggle (`data-theme`); Sidebar-Transition slide+fade; Split-Screen Blue-Glow; room-container Glassmorphism; Sticky Header + Controls (`position:sticky`); `.is-scrolled`-Elevation; **R4:** `neko-members` aus Controls-Bar entfernt (Import + Registrierung bereinigt); **Bugfix:** `.expanded` + `neko-side v-if` an `connected && side` gebunden; **R7:** `<neko-connect>` entfernt; `<transition name="page-fade" mode="out-in">` mit zwei exklusiven Branches (`<neko-login-page v-if="!connected">` / `<div v-else class="neko-room-wrapper">`); Room-DOM nur bei `connected`; Import + Registrierung auf `login.vue` / `neko-login-page` aktualisiert | ✅ |
| `header.vue` | Theme-Toggle-Button (SVG); Icon Micro-Animations (`scale`); funktionierender Glassmorphism-Diagonal-Gradient; CSS Tokens durchgehend | ✅ |
| `side.vue` | Glassmorphism (`backdrop-filter:blur(12px)`); Pill-Tabs; Tab-Inhalt-Transition `out-in`; **R1+R3:** 4-Tab-Leiste (Chat · Users · Files · Settings); Toggle-Logik (Chat+Users unabhängig, Settings exklusiv); 50/50-Split mit `panel-divider` wenn beide aktiv; `panel-grow`-Transition (max-height); **R2:** `<neko-userlist>` ersetzt Placeholder | ✅ |
| `userlist.vue` | **R2 + Bugfix:** Card-Layout (Avatar 32px links; Zeile 1: Name + Crown; Zeile 2: Buttons dauerhaft sichtbar); **Alle User:** Ignore/Unignore (`fa-eye-slash/fa-eye`); **Admin:** Mute/Unmute · Release Controls (`fa-times-circle`) · Force Take Controls (`fa-hand-paper`) · Give Controls (`fa-gamepad`); **Host (non-admin):** Give Controls; **Admin, non-admin Target:** Kick · Ban; Controls-Buttons nur bei `!implicitHosting`; Skeleton Loading 2-zeilig (Name + Actions); `isCurrentHost`-Getter; kein Right-Click-Kontextmenü (alle Aktionen inline) | ✅ |
| `chat.vue` | Pill-Username-Badges; Message-Hover; Textarea Token-Focus; Skeleton Loading (4 Shimmer-Messages) | ✅ |
| `members.vue` | 4-Zustand-Status-Dots (online/away/busy/offline); Avatar Hover-Scale; Skeleton Loading — **R4:** nicht mehr in Controls-Bar gerendert, Datei bleibt vorerst erhalten | ✅ |
| `menu.vue` | Icons auf `--color-text-muted`; `<select>` vollständig auf CSS Custom Properties migriert (Light-Mode-fix) | ✅ |
| `controls.vue` | Touch-Targets ≥44px; Micro-Animations; `.faded` → `--color-text-muted`; `fa-pause`/`fa-play` (Outline); Animated Counters FPS/Bitrate via `getStats()`+Lerp+`requestAnimationFrame` | ✅ |
| `settings.vue` | Bento Grid (2-Spalten); Custom Toggle Switches; LOG OUT Ghost-Button; **R6:** neue „Display"-Card (admin-only) mit Resolution-`<select>` (`W × H @ R fps`); `resValue` getter/setter → `$accessor.video.screenSet()` | ✅ |
| `video.vue` | WebRTC-Video + Maus/Tastatur/Touch-Event-Handler unberührt; **R5:** `fa-desktop`-Overlay-Button entfernt; `<neko-resolution>` entfernt; `Resolution`-Import, `@Ref` + `openResolution()` bereinigt | ✅ |
| `components/login.vue` *(neu, R7)* | Vollbild-Login-Page (ersetzt `connect.vue`-Overlay); Glassmorphism-Card (`.window`); radialer Spotlight-Gradient + Dot-Grid token-basiert; semantisches HTML (`<form>`, `<label for>`, `<input id>`, `<button type="submit">`); Inline-Fehleranzeige via `fieldError` (kein `$swal` mehr); `SystemDialog`-Overlay (`logout`, `DISCONNECT`, `ERROR`); `autoPassword`/`autoUser`-URL-Params in `mounted()`; `removeUrlParam` refactored (`const` + `.filter()`); `prefers-reduced-motion` für Bounce-Loader + Transitions; Touch-Targets ≥44px; Mobile `min(300px, calc(100vw - var(--space-8)))`; PriorityMatrix Kat. 1–5 vollständig | ✅ |
| `components/connect.vue` *(entfernt, R7)* | Vollständig durch `login.vue` ersetzt und gelöscht — kein Dead Code | ✅ |
| `store/client.ts` | **Bugfix:** `systemDialog`-State + `setSystemDialog`-Mutation hinzugefügt | ✅ |
| `neko/index.ts` | **Bugfix:** alle 3 `$swal`-Calls (`logout`, `SYSTEM.DISCONNECT`, `SYSTEM.ERROR`) → `$accessor.client.setSystemDialog(...)` | ✅ |
| `neko/base.ts` | Public getter `peerConnection` auf `BaseClient` (exponiert `protected _peer` sauber für Stats-Polling) | ✅ |
| `client/src/locale/*.ts` | **R1 i18n:** `side.users` Key in allen 15 Locale-Dateien ergänzt | ✅ |

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

**Implementiert in `side.vue` (Commit [2f26180](https://github.com/officialtaxz2/neko/commit/2f26180c6ae861612578b35c827a4d6f8eb531ee)) + i18n (Commits [d6fdd9f](https://github.com/officialtaxz2/neko/commit/d6fdd9f633ccfa17de4bcfe41a6106ca1328e37d) · [ec9ee31](https://github.com/officialtaxz2/neko/commit/ec9ee319ceae7383075b1c2f9fc59c93ab54bd1f))**

**Tab-Leiste:** `fa-comment-alt` Chat · `fa-users` Users · `fa-file` Files (conditional) · `fa-sliders-h` Settings

**Tab-Logik:**

| Zustand | Sidebar-Inhalt |
|---|---|
| Nur Chat aktiv | Chat nimmt 100% der Sidebar-Höhe |
| Nur Users aktiv | Users-Panel nimmt 100% der Sidebar-Höhe |
| Chat + Users aktiv | Chat oben / Users unten — je `flex: 1`; `panel-divider` dazwischen |
| Files aktiv | Files-Panel exklusiv (store-managed); Settings-State wird gecleart |
| Settings aktiv | Settings exklusiv; Chat+Users deaktiviert; beim Verlassen wird `activeChat = true` wiederhergestellt |

**Guards:**
- Letztes aktives Panel (Chat oder Users) kann nicht abgewählt werden
- Files-Aktivierung cleart `activeSettings`; Chat/Users-State bleibt für Rückkehr erhalten

**Transitions:**
- `panel-grow` (max-height-Trick) für Chat + Users (togglebar)
- `tab-fade out-in` für Files + Settings (exklusiv)

**i18n:** `side.users` in allen 15 Locale-Dateien ergänzt.

---

### R2 — `userlist.vue` ✅

**Initial: `userlist.vue` (neu) + `side.vue` (Placeholder ersetzt)**  
**Bugfix (30.04.2026): Layout-Redesign + fehlende Aktionen wiederhergestellt**

#### Layout (aktueller Stand)

Jeder User-Eintrag ist als **Card** aufgebaut:

```
┌─────────────────────────────────────────────┐
│ [Avatar 32px]  Username  [Crown-Icon?]      │  ← Zeile 1: Name-Row
│               [Btn][Btn][Btn][Btn][Btn]     │  ← Zeile 2: Actions (immer sichtbar)
└─────────────────────────────────────────────┘
```

- Avatar (`32px`, `border-radius: full`) links, `margin-top: space-1` für optische Zentrierung
- `.user-info` (flex-column) mit `user-name-row` + `user-actions`
- Buttons sind **dauerhaft sichtbar** — kein Hover-Reveal mehr

#### Self-Row

- Avatar + Name (muted) + `YOU`-Badge rechts
- Keine Action-Buttons

#### Aktions-Buttons (Sichtbarkeits-Logik)

| Button | Icon | Bedingung | Store-Aktion |
|---|---|---|---|
| Ignore / Unignore | `fa-eye-slash` / `fa-eye` | Alle User (immer sichtbar) | `user.setIgnored()` |
| Mute / Unmute | `fa-microphone-slash` / `fa-microphone` | `isAdmin` | `user.mute()` / `user.unmute()` |
| Give Controls | `fa-gamepad` | `!implicitHosting && m.id !== host && (isAdmin \| isCurrentHost)` | `remote.adminGive()` / `remote.give()` |
| Release Controls | `fa-times-circle` | `isAdmin && !implicitHosting && m.id === host` | `remote.adminRelease()` |
| Force Take Controls | `fa-hand-paper` | `isAdmin && !implicitHosting && m.id === host` | `remote.adminControl()` |
| Kick | `fa-user-times` | `isAdmin && !m.admin` | `user.kick()` |
| Ban IP | `fa-ban` | `isAdmin && !m.admin` | `user.ban()` |

**Release vs. Force Take:** Release entfernt den Host-Status (niemand hat danach Kontrolle); Force Take überträgt die Kontrolle direkt an den Admin.

#### Sonstiges

- `isCurrentHost`-Getter: `remote.id === user.id`
- Skeleton Loading: 2-zeilig (`sk-name` + `sk-actions`) passend zum Card-Layout
- Right-Click auf Row: `@contextmenu.stop.prevent` (unterdrückt Browser-Menü; kein Custom-Kontextmenü — alle Aktionen sind inline)
- Mute/Unmute + Kick/Ban jeweils mit `$swal`-Bestätigungsdialog
- `.action-btn--take` (amber) für Release + Force Take; `.action-btn--destructive` (red on hover) für Kick + Ban

---

### R3 — Multiple-Choice Sidebar ✅

In R1-Commit enthalten — siehe R1+R3 oben.

---

### R4 — User-Avatare aus Controls-Bar entfernen ✅

**Implementiert in `app.vue` + `_variables.scss`**

- `app.vue`: `<neko-members />` aus `.room-container` entfernt; `import Members` + Komponentenregistrierung bereinigt [cleanup]
- `_variables.scss`: `$controls-height` 125px → **64px** (war dimensioniert für Members-Row ~72px + Room-Menu ~53px; nach Entfernung unnötig hoch)
- Video-Player gewinnt ~61px Höhe zurück (`.video-container` hat `flex-grow: 1`)
- `members.vue` bleibt im Codebase erhalten (nicht mehr gerendert)

---

### R5 + R6 — Resolution aus Video-Player → Settings ✅

**Implementiert in `video.vue` (R5) + `settings.vue` (R6) — atomarer Commit**

**R5 — `video.vue`:**
- `<li v-if="admin"><i class="fas fa-desktop">` aus `.video-menu.top` entfernt
- `<neko-resolution ref="resolution" v-if="admin" />` aus Player entfernt
- `import Resolution`, `'neko-resolution': Resolution`, `@Ref('resolution')`, `openResolution()` entfernt [cleanup]
- Alle WebRTC/Maus/Tastatur/Touch-Event-Handler **unberührt**

**R6 — `settings.vue`:**
- Neue „Display"-Card (admin-only, `v-if="admin"`) im Bento Grid — nach Input-Card, vor Broadcast-Card
- `<select>` via bestehendem `.select`-Pattern (konsistent mit `keyboard_layout`-Row)
- Optionen: `W × H @ R fps` für jede `ScreenResolution` aus `$accessor.video.configurations`
- `resValue` getter: `"${width}x${height}@${rate}"` — zeigt aktive Auflösung automatisch
- `resValue` setter: `$accessor.video.screenSet(conf)` bei Auswahl
- `ScreenResolution`-Type aus `~/neko/types` importiert

---

### R7 — Dedizierte Login-Page ✅

**Implementiert in 3 Commits (01.05.2026):**
- `app.vue` patch — [0c2bd29](https://github.com/officialtaxz2/neko/commit/0c2bd29540a3998ef5434a591b7f6a1b0fd3be60)
- `login.vue` (neu) — [e380810](https://github.com/officialtaxz2/neko/commit/e380810286f74e9858a02d8b5063e8eae5430334)
- `connect.vue` (gelöscht) — [a91db10](https://github.com/officialtaxz2/neko/commit/a91db105841c3f158ff943ff09fe34e964e6a5ac)

#### Was sich geändert hat

**`app.vue`:**
- `<neko-connect v-if="!connected" />` (Overlay über Room) **entfernt**
- `<transition name="page-fade" mode="out-in">` umschließt zwei exklusive Branches:
  - `<neko-login-page v-if="!connected" key="login" />` — kein Room-DOM wenn nicht eingeloggt
  - `<div v-else key="room" class="neko-room-wrapper">` — gesamter Room (Header, Video, Controls, Sidebar, About, Notifications)
- `connected &&`-Checks auf `neko-side v-if` entfernt (implizit durch `v-else`-Branch)
- `.expanded`-Klasse von `#neko` auf `.neko-room-wrapper` verschoben
- Import + Registrierung: `Connect` → `LoginPage` / `'neko-login-page'`
- SCSS: `.page-fade`-Transition mit `prefers-reduced-motion: reduce`-Guard; `.neko-room-wrapper` trägt Layout + Media-Queries

**`login.vue` (neu, Rewrite von `connect.vue`):**
- Vollbild-Page (`position: fixed; inset: 0`) — kein Overlay mehr, einzige sichtbare Komponente bei `!connected`
- Glassmorphism-Card (`.window`): `--color-surface`, `backdrop-filter`, `--shadow-lg`
- Spotlight-Gradient + Dot-Grid-Hintergrund vollständig token-basiert (keine hardcodierten Farben)
- Semantisches HTML: `<form>`, `<label for="...">`, `<input id="...">`, `<button type="submit">`
- Inline-Fehleranzeige: `fieldError: string | null` unter dem Displayname-Feld (`role="alert"`, `aria-invalid`, `aria-describedby`) — kein `$swal` mehr
- `SystemDialog`-Overlay (`logout`, `DISCONNECT`, `ERROR`) vollständig übernommen
- `autoPassword`/`autoUser`-URL-Param-Logik in `mounted()` unverändert
- `removeUrlParam` refactored: `let`-Mutationen → `const` + `.filter()` [cleanup]
- `prefers-reduced-motion`: Bounce-Loader + alle Transitions statisch/deaktiviert
- Mobile: Card-Breite `min(300px, calc(100vw - var(--space-8)))` — kein Overflow auf 375px
- `box-shadow: var(--shadow-lg)` auf `.window` + `.system-dialog__card`
- `:focus-visible` auf Logo-Button, Submit, Dialog-Button
- PriorityMatrix Kat. 1–5 vollständig umgesetzt

**`connect.vue` (gelöscht):**
- Vollständig durch `login.vue` ersetzt; kein Dead Code verbleibt

#### Ergebnis

| Szenario | Verhalten |
|---|---|
| `!connected` | Nur `<neko-login-page>` im DOM — kein Room, kein Header, keine Controls |
| Login erfolgreich | `page-fade` (300ms) Login → Room |
| Sidebar war offen (`side=true`) bei Logout | Login-Page füllt 100% Breite — kein Layout-Bleed |
| `?pwd=xxx&usr=yyy` URL-Params | Auto-Login via `mounted()` funktioniert |
| `prefers-reduced-motion` | `page-fade` sofort; Bounce-Loader statisch |

---

## Design-System Tokens

```css
/* Art direction: Dark-first tech tool → deep cool-gray surfaces, blue accent
   Palette: HSL-based · Dark default · Light via [data-theme="light"] */

/* === DARK MODE (Standard) === */
--color-bg:                hsl(220, 15%, 9%);
--color-surface:           hsl(220, 13%, 13%);
--color-surface-2:         hsl(220, 11%, 16%);
--color-surface-offset:    hsl(220, 10%, 20%);
--color-surface-dynamic:   hsl(220, 9%, 24%);
--color-primary:           hsl(213, 90%, 62%);   /* Blue-Akzent */
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
--color-link:              hsl(213, 90%, 70%);
--color-link-hover:        hsl(213, 90%, 80%);

/* === LIGHT MODE ([data-theme="light"]) === */
--color-bg:                hsl(220, 20%, 97%);
--color-surface:           hsl(220, 18%, 100%);
--color-border:            hsl(220, 10%, 84%);
--color-text:              hsl(220, 20%, 12%);
--color-text-muted:        hsl(220, 12%, 44%);
--color-text-faint:        hsl(220, 10%, 64%);
--color-primary:           hsl(213, 90%, 44%);   /* dunkler für WCAG auf Weiß */
--color-primary-hover:     hsl(213, 90%, 36%);
--color-primary-active:    hsl(213, 90%, 28%);
--color-primary-highlight: hsl(213, 60%, 92%);
--color-link:              hsl(213, 90%, 38%);
--color-link-hover:        hsl(213, 90%, 28%);

/* === FONTS === */
--font-display:  'Cabinet Grotesk', 'Helvetica Neue', sans-serif;
--font-body:     'Satoshi', 'Inter', 'Helvetica Neue', sans-serif;
--font-mono:     'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
--font-controls: var(--font-mono);

/* === FLUID TYPE === */
--text-xs:   clamp(0.75rem,  0.70rem + 0.25vw, 0.875rem); /*  12–14px */
--text-sm:   clamp(0.875rem, 0.80rem + 0.35vw, 1rem);     /*  14–16px */
--text-base: clamp(1rem,     0.95rem + 0.25vw, 1.125rem); /*  16–18px */
--text-lg:   clamp(1.125rem, 1.00rem + 0.75vw, 1.5rem);   /*  18–24px */
--text-xl:   clamp(1.5rem,   1.20rem + 1.25vw, 2.25rem);  /*  24–36px */

/* === 8PT SPACING === */
--space-1: 0.25rem;  --space-2: 0.5rem;   --space-3: 0.75rem;
--space-4: 1rem;     --space-5: 1.25rem;  --space-6: 1.5rem;
--space-8: 2rem;     --space-10: 2.5rem;  --space-12: 3rem;
--space-16: 4rem;    --space-20: 5rem;    --space-24: 6rem;

/* === LAYOUT === */
$menu-height:     40px;
$controls-height: 64px;   /* R4: reduziert von 125px */
$side-width:      400px;

/* === SHADOWS, RADIUS, TRANSITIONS === */
--shadow-sm: 0 1px 2px rgba(0,0,0,0.35);    /* Light: rgba(30,40,60,0.08) */
--shadow-md: 0 4px 12px rgba(0,0,0,0.45);   /* Light: rgba(30,40,60,0.12) */
--shadow-lg: 0 12px 32px rgba(0,0,0,0.60);  /* Light: rgba(30,40,60,0.16) */

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
