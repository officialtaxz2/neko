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

| Datei / Ordner | Inhalt | Status |
|---|---|---|
| `client/src/assets/styles/_variables.scss` | Alle Design-Tokens: Farben, Fonts, Spacing, Shadows, Transitions als CSS Custom Properties | ✅ Rewritten |
| `client/src/assets/styles/main.scss` | Fontshare CDN-Import (Satoshi + Cabinet Grotesk), body/html-Basis, globale Resets, Focus-States, `prefers-reduced-motion` | ✅ Updated |
| `client/src/assets/styles/_reset.scss` | CSS-Reset | Unverändert |
| `client/src/assets/styles/vendor/` | Font Awesome, SweetAlert2, Tooltips, Emoji — nicht anfassen | Unverändert |
| `client/src/assets/styles/fonts/` | Lokale Whitney-Fonts (veraltet, nicht mehr importiert) | Inaktiv |
| `client/src/app.vue` | Root-Komponente: Theme-Initialisierung, `data-theme`-Toggle, Layout, Sidebar `<transition name="side">` Slide-Animation | ✅ Updated |
| `client/src/components/header.vue` | Topbar: Theme-Toggle-Button, CSS Tokens, Micro-Animations (Hover/Active) | ✅ Updated |
| `client/src/components/side.vue` | Sidebar: Pill-Tabs, Hover/Active Micro-Animations, Vue Tab-Transition, Touch-Targets, CSS Tokens, **Glassmorphism** (`backdrop-filter: blur(12px)`) | ✅ Updated |
| `client/src/components/chat.vue` | Chat-Panel: Pill-Username-Badges, Avatar-Hover, Message-Hover, Code-Block-Tokens, Textarea-Redesign, Skeleton Loading State (4 Shimmer-Messages) | ✅ Updated |
| `client/src/components/members.vue` | Members-Bar: Status-Dots (online/away/busy/offline), Avatar-Hover, Host/Admin-Badges, CSS Tokens, Skeleton Loading State (4 Shimmer-Circles) | ✅ Updated |
| `client/src/components/menu.vue` | Navigationsmenü: Icons auf `color: var(--color-text)` + Hover-Animation migriert, Language-Select vollständig von `$background-*`/`color: white` auf CSS Custom Properties umgestellt (Light-Mode-kompatibel) | ✅ Updated |
| `client/src/components/controls.vue` | Steuerleiste: Touch-Targets ≥ 44px, Micro-Animations (Hover `scale(1.18)` + Teal, Active `scale(0.88)`), `.faded`-Icons nutzen `--color-text-muted` (kein `opacity`-Hack), Lock-Switch-Track `--color-border`, Volume-Slider-Thumb mit `box-shadow`, CSS Tokens | ✅ Updated |
| `client/src/components/settings.vue` | Einstellungen-Panel: Custom Toggle Switches (Teal-Akzent, Spring-Easing), Touch-Target-Sizing via `display:inline-block + width:44px + height:44px + position:absolute` auf hidden input (kein Padding-Hack), Track-Farbe `--color-border` (sichtbar in Light Mode), Slider/Select/Input/Button Redesign, CSS Tokens durchgehend | ✅ Updated |
| `client/src/components/connect.vue` | Login/Connect-Dialog: CSS Tokens, `color-mix`-Overlay, Touch-Targets ≥ 44×44px auf Input + Button, Hover/Active Micro-Animations, **Glassmorphism** (`backdrop-filter: blur(12px)` auf `.window`, `blur(4px)` auf Overlay) | ✅ Updated |
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
| Micro-Animations auf Controls-Icons (Hover `scale(1.18)` + Teal, Active `scale(0.88)`, `will-change: transform`) | ✅ | `controls.vue` |
| Color-Coded Status Dot Indicators bei Nutzern (4 Zustände: online/away/busy/offline) | ✅ | `members.vue` |
| Pill-Shaped Tabs in der Sidebar-Navigation | ✅ | `side.vue` |
| Pill-Shaped Username-Badges im Chat | ✅ | `chat.vue` |
| Smooth Tab-Inhalt-Transition (Vue `<transition>` fade+slide, `mode="out-in"`) | ✅ | `side.vue` |
| Smooth Sidebar open/close Animation (Vue `<transition name="side">`, slide-from-right + fade, Mobile: slide-from-bottom) | ✅ | `app.vue` |
| Custom Toggle Switches im Settings-Panel (Teal-Akzent, Spring-Easing `cubic-bezier(0.16,1,0.3,1)`, CSS Tokens) | ✅ | `settings.vue` |
| Touch-Targets ≥ 44×44px: Controls (`min-width/height: 44px` auf allen `li`) | ✅ | `controls.vue` |
| Touch-Targets ≥ 44×44px: Connect, Settings | ✅ | `connect.vue`, `settings.vue` |

---

> **Test/Fix-Durchlauf abgeschlossen (30.04.2026):**
> Nach Abschluss von Phase 1 + 2 wurde ein vollständiger Light/Dark-Mode-Test durchgeführt. Behobene Bugs:
> - **Toggle-Switch-Sizing** (`settings.vue`): Label kollabierte auf 0px Breite (kein `width`/`height` → Pill unsichtbar, Thumb flog aus Container). Fix: `display:inline-block + width:44px + height:44px + position:absolute` auf hidden input.
> - **Toggle-Track-Kontrast** (`settings.vue`, `controls.vue`): Off-State-Track `--color-surface-offset` (94% L) nahezu unsichtbar auf weißem Hintergrund. Fix: `--color-border` (84% L, klar sichtbares Grau).
> - **`.faded`-Icons** (`controls.vue`): `color:text + opacity:0.4` kollabierte zu ~`#c5c6ca` in Light Mode. Fix: `color:var(--color-text-muted)` ohne Opacity.
> - **menu.vue `select`**: `color: white` hardcoded → unsichtbarer Text in Light Mode. Fix: Vollständige Migration auf CSS Custom Properties.
>
> **Phase 3 gestartet.**

---

### Phase 3 — Kat. 2: Layout & Visual Standard *(Standard: implementieren)*

| Task | Status | Datei(en) |
|---|---|---|
| Glassmorphism für Sidebar-Panels + Connect-Dialog (`backdrop-filter: blur(12px)`) | ✅ | `side.vue`, `connect.vue` |
| Gradient-System für Header-Topbar und Login-Screen | ⬜ | `header.vue`, `connect.vue` |
| Bento Grid Layout für Settings- und Members-Panel | ⬜ | `settings.vue`, `side.vue` |
| Split-Screen Layout: Video-Bereich + Sidebar visuell klarer getrennt | ⬜ | `app.vue` |

---

### Phase 4 — Kat. 3: Interaktion & Storytelling *(kontextabhängig)*

| Task | Status | Datei(en) |
|---|---|---|
| Animated Counters für Bitrate/FPS-Anzeige in der Steuerleiste | ⬜ | `controls.vue` |
| Smooth Scroll + Sticky Navigation | ⬜ | `app.vue` |

---

### Phase 5 — Kat. 4: Ästhetischer Stil *(Brand-Fit: Tech/Terminal)*

| Task | Status | Datei(en) |
|---|---|---|
| Terminal Aesthetic als optionaler Accent (Mono-Font in Controls-Leiste) | ⬜ | `controls.vue`, `_variables.scss` |
| Dot-Grid-Hintergrund im Connect-Dialog | ⬜ | `connect.vue` |

---

## Design-System Kurzreferenz

Dies sind die **tatsächlich implementierten** Token-Werte aus `_variables.scss`.

**Farb-Tokens (Dark Mode, Standard):**
```css
--color-bg:                hsl(220, 15%, 9%);
--color-surface:           hsl(220, 13%, 13%);
--color-surface-2:         hsl(220, 11%, 16%);
--color-surface-offset:    hsl(220, 10%, 20%);
--color-surface-dynamic:   hsl(220, 9%, 24%);
--color-primary:           hsl(174, 72%, 38%);   /* Teal-Akzent */
--color-primary-hover:     hsl(174, 72%, 30%);
--color-primary-highlight: hsl(174, 30%, 18%);
--color-text:              hsl(220, 10%, 86%);
--color-text-muted:        hsl(220, 8%, 50%);
--color-text-faint:        hsl(220, 7%, 34%);
--color-border:            hsl(220, 9%, 23%);   /* Toggle-Track off-state */
--color-error:             hsl(4, 68%, 52%);
--color-success:           hsl(142, 50%, 48%);
--color-warning:           hsl(36, 92%, 52%);
--color-link:              hsl(201, 90%, 60%);
```

**Light Mode (via `[data-theme="light"]`):**
```css
--color-bg:                hsl(220, 20%, 97%);
--color-surface:           hsl(220, 18%, 100%);
--color-border:            hsl(220, 10%, 84%);  /* Toggle-Track off-state — 16% Δ zu bg */
--color-text:              hsl(220, 20%, 12%);
--color-text-muted:        hsl(220, 12%, 44%);
--color-text-faint:        hsl(220, 10%, 64%);
--color-primary:           hsl(174, 72%, 30%);
```

**Font-Stack:**
```css
--font-display: 'Cabinet Grotesk', 'Helvetica Neue', sans-serif;
--font-body:    'Satoshi', 'Inter', 'Helvetica Neue', sans-serif;
--font-mono:    'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
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

```
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
