# n.eko — Design Overhaul Fork

> **Fork von:** [m1k1o/neko](https://github.com/m1k1o/neko) — ein selbst gehosteter virtueller Browser via Docker + WebRTC.
> **Dieses Repo:** `officialtaxz2/neko` — eigenständiger Fork mit vollständigem UI/UX-Redesign.
> **Upstream-PRs:** Kein komplettes Design-Overhaul ins Upstream-Repo einreichen — dieser Fork lebt als eigenständiges Produkt.

---

## Kontext & Ziel

Das ursprüngliche neko-Frontend war ein **1:1-Discord-Klon**: Farbpalette `#36393f / #2f3136 / #202225`, Discord-exklusive Font „Whitney", keine eigene Designsprache, keine modernen UX-Patterns.

Ziel dieses Forks: **Komplett eigenständiges Design** auf Basis eines modernen Design-Systems — mit Dark/Light-Mode, fluid Typography, CSS Custom Properties, Glassmorphism-Elementen und sauberer Komponentenhierarchie.

---

## Dev-Umgebung aufsetzen (Einmalig)

Der gesamte Build läuft in Docker — **kein Go, Node.js oder X11 auf dem Host nötig**.

```bash
# 1. Repo klonen
git clone https://github.com/officialtaxz2/neko.git
cd neko

# 2. Backend-Container starten (unverandertes Base-Image reicht für Frontend-Dev)
docker compose up -d

# 3. Frontend-Dev-Server mit Hot Reload starten (kein Node auf Host)
cd client/dev
./serve -i   # -i = node_modules beim ersten Mal installieren
# → http://localhost:3001 — Änderungen an client/src/ sofort sichtbar (HMR)
```

Fertiges Image bauen & testen:
```bash
# Alles neu bauen (Server + Client)
./build -r my-neko

# Nur Firefox-App neu bauen (Base-Image wiederverwenden, schneller)
./build -a firefox -b ghcr.io/m1k1o/neko/base:latest -r my-neko

# In docker-compose.yaml: image: my-neko/firefox:latest (statt ghcr.io/...)
docker compose up --force-recreate
```

---

## Relevante Dateien für das Overhaul

Nur Dateien, die tatsächlich modifiziert wurden oder noch modifiziert werden müssen. Unveränderte Dateien (`about.vue`, `avatar.vue`, `clipboard.vue`, `context.vue`, `emoji.vue`, `emote.vue`, `emotes.vue`, `files.vue`, `markdown.ts`, `resolution.vue`, `_reset.scss`, `vendor/`) sind nicht aufgeführt.

| Datei / Ordner | Inhalt | Status |
|---|---|---|
| `client/src/assets/styles/_variables.scss` | Alle Design-Tokens: Farben, Fonts, Spacing, Shadows, Transitions als CSS Custom Properties; Primär-Akzent **Blue `hsl(213)`** (ersetzt Discord-Teal); `--font-controls` als optionaler Control-Bar-Font-Slot (standardmäßig auf `--font-mono`) | ✅ Rewritten |
| `client/src/assets/styles/main.scss` | Fontshare CDN-Import (Satoshi + Cabinet Grotesk), body/html-Basis, globale Resets, Focus-States, `prefers-reduced-motion` | ✅ Updated |
| `client/src/app.vue` | Root-Komponente: Theme-Initialisierung, `data-theme`-Toggle, Sidebar `<transition name="side">` (slide-from-right + fade; Mobile: slide-from-bottom), **Split-Screen-Glow** (`#neko.expanded .neko-main { box-shadow: 4px 0 20px color-mix(primary 10%) }`), **room-container Glassmorphism** (`backdrop-filter:blur(8px)` + aufwärts-Gradient + Blue-Tint `border-top`), **hardcoded `rgba(0,0,0,0.4)` auf `var(--color-bg)` migriert**, `header-container: background:transparent`, **Sticky Navigation** (`.header-container position:sticky top:0`, `.room-container position:sticky bottom:0`, beide `z-index:10`), **`scroll-behavior:smooth`** auf `.neko-main`, **`.is-scrolled`-Elevation** (passiver scroll-Listener via `mainContainer`-Ref, `scrollTop>4px` → Header `backdrop-filter:blur(8px)` + semi-transparenter BG + `shadow-sm`, `beforeDestroy`-Cleanup) | ✅ Updated |
| `client/src/components/header.vue` | Topbar: Theme-Toggle-Button (Sonne/Mond-SVG, `aria-label`, Blue-Hover), Icon-Micro-Animations (Hover `scale(1.08)`, Active `scale(0.95)`), **funktionierender Glassmorphism-Gradient** (`linear-gradient(135deg, color-mix(primary 12%, transparent) → color-mix(bg 82%, transparent))` — transparenzbasiert, Blur aktiv auch im Ruhezustand), CSS Tokens durchgehend, `border-bottom: color-mix(border 70%, transparent)` | ✅ Updated |
| `client/src/components/side.vue` | Sidebar (`<aside>`): Pill-Tabs (aktiver Tab `background:primary-highlight`, Hover-Scale), Tab-Inhalt `<transition>` fade+slide `mode="out-in"`, **Glassmorphism** (`backdrop-filter:blur(12px)` auf `.neko-menu` + `.tabs-container`, semi-transparente Backgrounds via `color-mix`), **`page-container overflow-y:auto`** (Bento-scroll-support), `border-left: color-mix(border 70%, transparent)` | ✅ Updated |
| `client/src/components/chat.vue` | Chat-Panel: Pill-Username-Badges (`border-radius:var(--radius-full)`), Message-Hover (`background:var(--color-surface-offset)`), Code-Block CSS Tokens, Textarea-Redesign (Token-basierter Focus-Ring), **Skeleton Loading State** (4 Shimmer-Messages via `@keyframes shimmer`) | ✅ Updated |
| `client/src/components/members.vue` | Members-Bar unterhalb Video: **Status-Dots** (4 Zustände: online/away/busy/offline via `--color-success/warning/error/text-muted`), Avatar-Hover (`scale(1.08)`), Host/Admin-Badges (CSS Tokens), **Skeleton Loading State** (4 Shimmer-Circles) | ✅ Updated |
| `client/src/components/menu.vue` | Navigationsmenü (Hamburger): Icons `color:var(--color-text-muted)` + Hover-Blue, Language-`<select>` vollständig von `$background-*/color:white` auf CSS Custom Properties migriert (Light-Mode-fix) | ✅ Updated |
| `client/src/components/controls.vue` | Steuerleiste: Touch-Targets `min-width/height:44px`, Micro-Animations (Hover `scale(1.18)` + Blue, Active `scale(0.88)`, `will-change:transform`), `.faded`-Icons `color:var(--color-text-muted)` (kein `opacity`-Hack), **`fa-pause` / `fa-play`** (Outline-Icons — ersetzen die alten Solid-Circle-Icons `fa-pause-circle`/`fa-play-circle`, die als gefüllter Disc visuell wie ein farbiger Button wirkten), Lock-Switch-Track `--color-border`, Volume-Slider-Thumb mit `box-shadow`, **Animated Counters** (Bitrate/FPS-Badge: `RTCPeerConnection.getStats()` alle 2s via `$client.peerConnection`, Lerp-Animation `displayFps/displayBitrateKbps → targetFps/targetBitrateKbps` per `requestAnimationFrame`, Stat-Flip-`@keyframes` beim Eintreffen neuer Werte via `:key="statsKey"`, Control-Bar-Font via `--font-controls`, `tabular-nums`, `stat-val` in `--color-primary`, nur sichtbar wenn `playing && statsVisible`, `beforeDestroy`-Cleanup) | ✅ Updated |
| `client/src/components/settings.vue` | **Bento Grid Rewrite**: 2-Spalten CSS-Grid, 5 semantisch gruppierte Cards (Scrolling full-width / Chat / Input / Broadcast full-width / Session), Card-Header mit Blue-Icon + uppercase Label, `hover:box-shadow:var(--shadow-md)`, alle bestehenden Controls (Switch, Slider, Select) erhalten, **Custom Toggle Switches** (`width:44px`, Spring-Easing, Blue-Akzent via `--color-primary`, Off-Track `--color-border`), **LOG OUT-Button** als Ghost-Style (transparentes Hintergrund, `--color-border`-Rand, kein solid Fill — neutral wie alle anderen Buttons at rest) | ✅ Rewritten |
| `client/src/components/connect.vue` | Login/Connect-Dialog: Overlay `backdrop-filter:blur(4px)` (semi-transparent), `.window` `backdrop-filter:blur(12px)` + `background:color-mix(surface 75%, transparent)` + `border-top` Glass-Highlight, **2-Layer-Hintergrund** aus radialem Spotlight-Gradient (`radial-gradient(ellipse, color-mix(primary 10%, bg) → bg)`) plus **Dot-Grid-Pattern** als optionaler Terminal-Aesthetic-Accent, Input-Felder `background:color-mix(bg 70%, transparent)`, CSS Tokens, Touch-Targets ≥44px | ✅ Updated |
| `client/src/neko/base.ts` | **`peerConnection` Getter** (`get peerConnection(): RTCPeerConnection \| undefined { return this._peer }`) — exponiert den internen `protected _peer` sauber nach außen; kein `any`-Hack auf protected Fields nötig. Benötigt von `controls.vue` für Stats-Polling | ✅ Updated |
| `client/src/components/video.vue` | WebRTC-Video + Maus/Tastatur-Overlay — **zuletzt anfassen**, Event-Handler nicht verändern | ⬜ Offen |

**Empfohlene Bearbeitungsreihenfolge (von außen nach innen):**
```
_variables.scss → main.scss → app.vue → header.vue → side.vue → chat.vue → members.vue → menu.vue → controls.vue → settings.vue → connect.vue → video.vue
```

---

## Design Roadmap

Die Roadmap folgt der **Prioritätsmatrix** aus dem Design-System (Kat. 0–4).

> **Legende:** ✅ Fertig · 🔄 In Arbeit (teilweise) · ⬜ Offen

---

### Phase 1 — Kat. 0: Technische Basis *(Pflicht, kein Opt-Out)*

| Task | Status | Datei(en) |
|---|---|---|
| Discord-Palette vollständig entfernen | ✅ | `_variables.scss` |
| HSL-basiertes CSS Custom Property System (`--color-bg`, `--color-surface`, `--color-primary` etc.) | ✅ | `_variables.scss` |
| Whitney-Font ersetzen durch Fontshare CDN (**Satoshi** Body + **Cabinet Grotesk** Display) | ✅ | `_variables.scss`, `main.scss` |
| Fluid Typography mit `clamp()` für alle Schriftgrößen (`--text-xs` bis `--text-xl`) | ✅ | `_variables.scss` |
| 8pt-Spacing-Grid als Tokens (`--space-1` bis `--space-24`) | ✅ | `_variables.scss` |
| Dark Mode + Light Mode Token-Sets + `prefers-color-scheme`-Fallback | ✅ | `_variables.scss` |
| Theme-Toggle mit `data-theme`-Attribut auf `<html>`, OS-Präferenz-Listener | ✅ | `app.vue`, `header.vue` |
| Skeleton Screens & Loading States | ✅ | `chat.vue`, `members.vue` |

---

### Phase 2 — Kat. 1: Universal UX Patterns *(immer, außer explizit widerlegt)*

| Task | Status | Datei(en) |
|---|---|---|
| Focus-States global sichtbar (`:focus-visible`, nie `outline: none` ohne Ersatz) | ✅ | `main.scss` |
| `prefers-reduced-motion` global implementiert | ✅ | `main.scss` |
| Micro-Animations auf Header-Icons + Theme-Toggle (Hover `scale(1.08)`, Active `scale(0.95)`) | ✅ | `header.vue` |
| Micro-Animations auf Sidebar-Tabs (Hover `scale(1.15)` Icon, `translateY(-1px)`, Active `scale(0.96)`) | ✅ | `side.vue` |
| Micro-Animations auf Chat- und Members-Avataren (Hover `scale(1.08)`, Active `scale(0.95)`) | ✅ | `chat.vue`, `members.vue` |
| Micro-Animations auf Controls-Icons (Hover `scale(1.18)` + Blue, Active `scale(0.88)`, `will-change: transform`) | ✅ | `controls.vue` |
| Color-Coded Status Dot Indicators bei Nutzern (4 Zustände: online/away/busy/offline) | ✅ | `members.vue` |
| Pill-Shaped Tabs in der Sidebar-Navigation | ✅ | `side.vue` |
| Pill-Shaped Username-Badges im Chat | ✅ | `chat.vue` |
| Smooth Tab-Inhalt-Transition (Vue `<transition>` fade+slide, `mode="out-in"`) | ✅ | `side.vue` |
| Smooth Sidebar open/close Animation (Vue `<transition name="side">`, slide-from-right + fade, Mobile: slide-from-bottom) | ✅ | `app.vue` |
| Custom Toggle Switches im Settings-Panel (Blue-Akzent via `--color-primary`, Spring-Easing `cubic-bezier(0.16,1,0.3,1)`, CSS Tokens) | ✅ | `settings.vue` |
| Touch-Targets ≥ 44×44px: Controls (`min-width/height: 44px` auf allen `li`) | ✅ | `controls.vue` |
| Touch-Targets ≥ 44×44px: Connect, Settings | ✅ | `connect.vue`, `settings.vue` |

---

> **Test/Fix-Durchlauf abgeschlossen (30.04.2026 — Abschluss Phase 1+2):**
> Nach Abschluss von Phase 1 + 2 wurde ein vollständiger Light/Dark-Mode-Test durchgeführt. Behobene Bugs:
> - **Toggle-Switch-Sizing** (`settings.vue`): Label kollabierte auf 0px Breite (kein `width`/`height` → Pill unsichtbar, Thumb flog aus Container). Fix: `display:inline-block + width:44px + height:44px + position:absolute` auf hidden input.
> - **Toggle-Track-Kontrast** (`settings.vue`, `controls.vue`): Off-State-Track `--color-surface-offset` (94% L) nahezu unsichtbar auf weißem Hintergrund. Fix: `--color-border` (84% L, klar sichtbares Grau).
> - **`.faded`-Icons** (`controls.vue`): `color:text + opacity:0.4` kollabierte zu ~`#c5c6ca` in Light Mode. Fix: `color:var(--color-text-muted)` ohne Opacity.
> - **menu.vue `select`**: `color: white` hardcoded → unsichtbarer Text in Light Mode. Fix: Vollständige Migration auf CSS Custom Properties.

---

### Phase 3 — Kat. 2: Layout & Visual Standard *(Standard: implementieren)* ✅ Abgeschlossen (30.04.2026)

| Task | Status | Datei(en) | Hinweis |
|---|---|---|---|
| Glassmorphism für Sidebar + Connect-Dialog (`backdrop-filter:blur(12px)`) | ✅ | `side.vue`, `connect.vue` | Semi-transparente Backgrounds via `color-mix(… transparent)` — Blur effektiv |
| Gradient-System für Header-Topbar + Connect-Login-Screen | ✅ | `header.vue`, `connect.vue` | Header: transparenzbasierter Diagonal-Gradient mit aktivem Blur auch im Ruhezustand. Connect: radiales Spotlight-Gradient |
| Bento Grid Layout für Settings-Panel | ✅ | `settings.vue`, `side.vue` | 2-Spalten CSS-Grid, 5 Cards |
| Split-Screen Layout: Video-Bereich + Sidebar visuell klarer getrennt | ✅ | `app.vue` | Blue-Glow rechter Rand + room-container Glassmorphism + `rgba`→Token-Migration |

---

### Phase 4 — Kat. 3: Interaktion & Storytelling *(kontextabhängig)* ✅ Abgeschlossen (30.04.2026)

| Task | Status | Datei(en) |
|---|---|---|
| Animated Counters für Bitrate/FPS-Anzeige in der Steuerleiste | ✅ | `controls.vue`, `neko/base.ts` |
| Smooth Scroll + Sticky Navigation | ✅ | `app.vue` |

**Animated Counters (Details):**
- `base.ts`: Neuer public getter `peerConnection` auf `BaseClient` — exponiert `protected _peer` ohne `any`-Hack
- `controls.vue`: `RTCPeerConnection.getStats()` Polling alle 2s; FPS aus `r.framesPerSecond`, Bitrate aus `bytesReceived`-Delta × 8 / Zeitdelta (Kbps/Mbps)
- Lerp-Animation: `displayFps/displayBitrateKbps` interpolieren per `requestAnimationFrame` zum Zielwert (Faktor 0.18), stoppt automatisch bei Zielwert
- `stat-flip` `@keyframes` (220ms) beim Eintreffen neuer Poll-Daten via `:key="statsKey"`
- Badge: Control-Bar-Font via `--font-controls`, `tabular-nums`, `stat-val` in `--color-primary`, nur sichtbar wenn `playing && statsVisible`; `@Watch('playing')` reset auf 0 bei Pause

**Sticky Navigation (Details):**
- `.header-container`: `position:sticky; top:0; z-index:10` — bleibt bei Mobile-Scroll sichtbar
- `.room-container`: `position:sticky; bottom:0; z-index:10` — Steuerleiste bleibt beim Scrollen erreichbar
- `scroll-behavior:smooth` auf `.neko-main` — alle Scroll-Events (programmatisch + nativ) laufen smooth
- `.is-scrolled`-State: passiver `scroll`-Listener auf `mainContainer`-Ref; bei `scrollTop > 4px` → Header `backdrop-filter:blur(8px)` + semi-transparenter BG + `shadow-sm`
- `beforeDestroy`: Listener korrekt entfernt, kein Memory Leak

---

### Phase 5 — Kat. 4: Ästhetischer Stil *(Brand-Fit: Tech/Terminal)* ✅ Abgeschlossen (30.04.2026)

| Task | Status | Datei(en) |
|---|---|---|
| Terminal Aesthetic als optionaler Accent (Mono-Font in Controls-Leiste) | ✅ | `controls.vue`, `_variables.scss` |
| Dot-Grid-Hintergrund im Connect-Dialog | ✅ | `connect.vue` |
| Header-Glassmorphism reparieren (semi-transparenter BG für funktionierenden `backdrop-filter` im Ruhezustand) | ✅ | `header.vue` |

---

> **Testphase Bug-Fix-Protokoll (30.04.2026 — Live-Test nach Phase 1–5):**
> Nach Abschluss aller 5 Phasen wurde ein Live-Testlauf durchgeführt. Behobene Bugs:
> - **Primär-Akzentfarbe** (`_variables.scss`): Teal `hsl(174, 72%, 38%)` durch **Modern Blue `hsl(213, 90%, 62%)`** (Dark) / `hsl(213, 90%, 44%)` (Light) ersetzt. Alle abgeleiteten Tokens (`-hover`, `-active`, `-highlight`) sowie Link-Tokens (`hsl(201)` → `hsl(213)`) konsistent mitgezogen. Betrifft: Slider-Fills, Toggle-Tracks (ON-State), Tab-Aktivlinie, Stats-Badge-Werte, Focus-Rings, LOG OUT-Button-Hover.
> - **Controls-Bar Pause/Play-Icon** (`controls.vue`): `fa-pause-circle` / `fa-play-circle` (Solid Filled Circle) → `fa-pause` / `fa-play` (Outline). Die Solid-Circle-Variante renderte als vollständig ausgefüllter farbiger Disc — visuell ununterscheidbar von einem Teal/Blue-Button, inkonsistent mit allen anderen Controls-Bar-Icons.
> - **LOG OUT-Button** (`settings.vue`): Vollständig gefüllter Solid-Primary-Button → Ghost-Style (transparentes Hintergrund, `--color-border`-Rand). Konsistent mit dem Designprinzip: kein persistentes Primärfarbige Fill auf Buttons at rest.

---

## Design-System Kurzreferenz

// Art direction: Dark-first tech tool → deep cool-gray surfaces, blue accent
// Palette: HSL-based, dark default · light via [data-theme="light"]

Dies sind die **tatsächlich implementierten** Token-Werte aus `_variables.scss`.

**Farb-Tokens (Dark Mode, Standard):**
```css
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
--color-border:            hsl(220, 9%, 23%);   /* Toggle-Track off-state */
--color-error:             hsl(4, 68%, 52%);
--color-success:           hsl(142, 50%, 48%);
--color-warning:           hsl(36, 92%, 52%);
--color-link:              hsl(213, 90%, 70%);
--color-link-hover:        hsl(213, 90%, 80%);
```

**Light Mode (via `[data-theme="light"]`):**
```css
--color-bg:                hsl(220, 20%, 97%);
--color-surface:           hsl(220, 18%, 100%);
--color-border:            hsl(220, 10%, 84%);  /* Toggle-Track off-state — 16% Δ zu bg */
--color-text:              hsl(220, 20%, 12%);
--color-text-muted:        hsl(220, 12%, 44%);
--color-text-faint:        hsl(220, 10%, 64%);
--color-primary:           hsl(213, 90%, 44%);  /* Blue — dunkler für WCAG-Kontrast auf Weiß */
--color-primary-hover:     hsl(213, 90%, 36%);
--color-primary-active:    hsl(213, 90%, 28%);
--color-primary-highlight: hsl(213, 60%, 92%);
--color-link:              hsl(213, 90%, 38%);
--color-link-hover:        hsl(213, 90%, 28%);
```

**Font-Stack:**
```css
--font-display:  'Cabinet Grotesk', 'Helvetica Neue', sans-serif;
--font-body:     'Satoshi', 'Inter', 'Helvetica Neue', sans-serif;
--font-mono:     'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
--font-controls: var(--font-mono);
```

**Fluid Typography:**
```css
--text-xs:   clamp(0.75rem,  0.70rem + 0.25vw, 0.875rem); /*  12–14px */
--text-sm:   clamp(0.875rem, 0.80rem + 0.35vw, 1rem);     /*  14–16px */
--text-base: clamp(1rem,     0.95rem + 0.25vw, 1.125rem); /*  16–18px */
--text-lg:   clamp(1.125rem, 1.00rem + 0.75vw, 1.5rem);   /*  18–24px */
--text-xl:   clamp(1.5rem,   1.20rem + 1.25vw, 2.25rem);  /*  24–36px */
```

**Spacing (8pt-Grid, `--space-1` bis `--space-24`):**
```css
--space-1: 0.25rem;  /*  4px */  --space-2: 0.5rem;   /*  8px */
--space-3: 0.75rem;  /* 12px */  --space-4: 1rem;     /* 16px */
--space-5: 1.25rem;  /* 20px */  --space-6: 1.5rem;   /* 24px */
--space-8: 2rem;     /* 32px */  --space-10: 2.5rem;  /* 40px */
--space-12: 3rem;    /* 48px */  --space-16: 4rem;    /* 64px */
--space-20: 5rem;    /* 80px */  --space-24: 6rem;    /* 96px */
```

**Shadows, Radius, Transitions:**
```css
--shadow-sm:  0 1px 2px rgba(30,40,60,0.08);   /* Light Mode */
--shadow-md:  0 4px 12px rgba(30,40,60,0.12);
--shadow-lg:  0 12px 32px rgba(30,40,60,0.16);
/* Dark Mode: rgba(0,0,0,0.35/0.45/0.60) */

--radius-sm: 0.25rem;  --radius-md: 0.5rem;
--radius-lg: 0.75rem;  --radius-xl: 1rem;  --radius-full: 9999px;

--transition-fast:        100ms cubic-bezier(0.16, 1, 0.3, 1);
--transition-interactive: 180ms cubic-bezier(0.16, 1, 0.3, 1);
--transition-slow:        300ms cubic-bezier(0.16, 1, 0.3, 1);
```

---

## Design-Abnahme-Checkliste

Vor jedem Commit / PR in diesem Fork prüfen:

```text
Design & UX:
  [ ] Alle Views responsive (320px – 2560px) / Mobile First
  [ ] Dark Mode vorhanden und korrekt
  [ ] Light Mode: Toggle-Tracks, Icons, Select-Elemente auf Kontrast prüfen
  [ ] Kontrast ≥ 4.5:1 (WCAG AA)
  [ ] Focus-States sichtbar
  [ ] Touch-Targets ≥ 44×44px
  [ ] prefers-reduced-motion implementiert
  [ ] Kein hardcodierter Farbwert (kein color:white, kein #36393f etc.)
  [ ] Keine alten $background-*/color:white SCSS-Vars in neuen Komponenten

Feature-Coverage:
  [ ] Kat. 0–1 vollständig implementiert
  [ ] Kat. 2–4 implementiert oder Opt-Out dokumentiert
```

---

## Upstream-Projekt

Originales Projekt: [m1k1o/neko](https://github.com/m1k1o/neko)
Dokumentation: [neko.m1k1o.net](https://neko.m1k1o.net/)

> Bugfixes und kleinere UX-Verbesserungen können als PR ans Upstream-Projekt eingereicht werden.
> Das komplette Design-Overhaul lebt in diesem Fork als eigenständiges Produkt.
