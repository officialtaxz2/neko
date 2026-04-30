# n.eko — Design Overhaul Fork

> **Fork von:** [m1k1o/neko](https://github.com/m1k1o/neko) — ein selbst gehosteter virtueller Browser via Docker + WebRTC.
> **Dieses Repo:** `officialtaxz2/neko` — eigenständiger Fork mit vollständigem UI/UX-Redesign.
> **Upstream-PRs:** Kein komplettes Design-Overhaul ins Upstream-Repo einreichen — dieser Fork lebt als eigenständiges Produkt.

---

## Kontext & Ziel

Das ursprüngliche neko-Frontend ist ein **1:1-Discord-Klon**: Farbpalette `#36393f / #2f3136 / #202225`, Discord-exklusive Font „Whitney", keine eigene Designsprache, keine modernen UX-Patterns.

Ziel dieses Forks: **Komplett eigenständiges Design** auf Basis eines modernen Design-Systems — mit Dark/Light-Mode, fluid Typography, CSS Custom Properties, Glassmorphism-Elementen und sauberer Komponentenhierarchie.

---

## Dev-Umgebung aufsetzen (Einmalig)

Der gesamte Build läuft in Docker — **kein Go, Node.js oder X11 auf dem Host nötig**.

```bash
# 1. Repo klonen
git clone https://github.com/officialtaxz2/neko.git
cd neko

# 2. Backend-Container starten (unverändertes Base-Image)
docker compose up -d

# 3. Frontend-Dev-Server mit Hot Reload starten (kein Node auf Host)
cd client/dev
./serve -i   # -i = node_modules beim ersten Mal installieren
# → http://localhost:3001 mit sofortigem HMR bei jeder CSS/Vue-Änderung
```

Für schnelle Iterationen:
```bash
# Nur Firefox-App neu bauen (Base-Image wiederverwenden)
./build -a firefox -b ghcr.io/m1k1o/neko/base:latest -r my-neko

# Alles neu bauen (Server + Client)
./build -r my-neko
docker compose up --force-recreate
```

---

## Relevante Dateien für das Overhaul

| Datei / Ordner | Inhalt |
|---|---|
| `client/src/assets/styles/_variables.scss` | **Startpunkt:** alle Farben, Fonts, Layout-Größen als SCSS-Variablen (aktuell Discord-Palette) |
| `client/src/assets/styles/main.scss` | Importiert Variablen + Vendor-Styles, setzt body/html-Basis |
| `client/src/assets/styles/_reset.scss` | CSS-Reset |
| `client/src/assets/styles/vendor/` | Fremde Styles (Font Awesome, SweetAlert2, Tooltips) — nicht anfassen |
| `client/src/assets/styles/fonts/` | Lokal eingebundene Fonts (Whitney — wird ersetzt) |
| `client/src/app.vue` | Root-Komponente, Haupt-Layout |
| `client/src/components/header.vue` | Topbar |
| `client/src/components/side.vue` | Seitenleiste (Chat, Members, Files) |
| `client/src/components/chat.vue` | Chat-Panel |
| `client/src/components/controls.vue` | Steuerleiste |
| `client/src/components/connect.vue` | Login/Connect-Dialog |
| `client/src/components/settings.vue` | Einstellungen-Panel |
| `client/src/components/video.vue` | WebRTC-Video + Maus/Tastatur-Overlay — **zuletzt anfassen**, Logik stehen lassen |

**Reihenfolge beim Bearbeiten:**
```
_variables.scss → app.vue → header.vue → side.vue → chat.vue → controls.vue → connect.vue → settings.vue → video.vue
```

---

## Design Roadmap

Die Roadmap folgt der **Prioritätsmatrix** aus dem Design-System (Kat. 0–4).

> **Legende:** ✅ Fertig · 🔄 In Arbeit · ⬜ Offen

---

### Phase 1 — Kat. 0: Technische Basis *(Pflicht, kein Opt-Out)*

| Task | Status | Datei(en) |
|---|---|---|
| Discord-Palette vollständig entfernen | ⬜ | `_variables.scss` |
| HSL-basiertes CSS Custom Property System einführen (`--color-bg`, `--color-surface`, `--color-primary` etc.) | ⬜ | `_variables.scss` |
| Whitney-Font ersetzen durch Variable Font (Fontshare: **Satoshi** Body + **Cabinet Grotesk** Display) | ⬜ | `_variables.scss`, `main.scss`, `fonts/` |
| Fluid Typography mit `clamp()` für alle Schriftgrößen | ⬜ | `_variables.scss` |
| 8pt-Spacing-Grid als Tokens (`--space-1` bis `--space-32`) | ⬜ | `_variables.scss` |
| Dark Mode + Light Mode Token-Sets | ⬜ | `_variables.scss`, `app.vue` |
| Theme-Toggle (Dark/Light) mit `data-theme`-Attribut + `prefers-color-scheme`-Fallback | ⬜ | `app.vue`, `header.vue` |
| Skeleton Screens & Loading States | ⬜ | alle Komponenten |

---

### Phase 2 — Kat. 1: Universal UX Patterns *(immer, außer explizit widerlegt)*

| Task | Status | Datei(en) |
|---|---|---|
| Micro-Animations: Hover-States auf Sidebar-Icons (`transform: scale(1.05)`), Button-Transitions `active:scale-95` | ⬜ | `side.vue`, `controls.vue` |
| Color-Coded Status + Status Dot Indicators bei Nutzern (Online / Away / Busy) | ⬜ | `side.vue`, `chat.vue` |
| Collapsible Sections: Chat, Members, Files als aufklappbare Panels | ⬜ | `side.vue` |
| Toggle Switches im Settings-Panel (custom, animiert) | ⬜ | `settings.vue` |
| Pill-Shaped Elements: Status-Labels, User-Badges | ⬜ | `side.vue`, `chat.vue` |
| Smooth Panel-Ein-/Ausblenden (keine Instant-Show/Hide) | ⬜ | alle Panel-Komponenten |
| Focus-States sichtbar, nie `outline: none` ohne Ersatz | ⬜ | `_variables.scss` |
| Touch-Targets ≥ 44×44px, 8px Gap zwischen Touch-Elementen | ⬜ | alle Komponenten |

---

### Phase 3 — Kat. 2: Layout & Visual Standard *(Standard: implementieren)*

| Task | Status | Datei(en) |
|---|---|---|
| Glassmorphism für Sidebar-Panels + Connect-Dialog (`backdrop-filter: blur(12px)`) | ⬜ | `side.vue`, `connect.vue` |
| Gradient-System für Header-Topbar und Login-Screen | ⬜ | `header.vue`, `connect.vue` |
| Bento Grid Layout für Settings- und Members-Panel | ⬜ | `settings.vue`, `side.vue` |
| Split-Screen Layout: Video-Bereich + Sidebar klarer getrennt | ⬜ | `app.vue` |

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

**Farb-Tokens (Zielzustand):**
```scss
// Dark Mode (Standard)
--color-bg:       hsl(220, 13%, 10%);
--color-surface:  hsl(220, 13%, 13%);
--color-surface-2: hsl(220, 13%, 16%);
--color-primary:  hsl(186, 98%, 28%);   // Teal-Akzent
--color-text:     hsl(220, 10%, 85%);
--color-text-muted: hsl(220, 8%, 55%);
```

**Font-Stack (Zielzustand):**
```scss
--font-display: 'Cabinet Grotesk', 'Helvetica Neue', sans-serif;
--font-body:    'Satoshi', 'Inter', sans-serif;
--font-mono:    'JetBrains Mono', 'Fira Code', monospace;
```

**Spacing (8pt-Grid):**
```scss
--space-1: 0.25rem;  // 4px
--space-2: 0.5rem;   // 8px
--space-4: 1rem;     // 16px
--space-8: 2rem;     // 32px
// ... bis --space-32: 8rem (128px)
```

---

## Design-Abnahme-Checkliste

Vor jedem Commit / PR in diesem Fork prüfen:

```
Design & UX:
  [ ] Alle Views responsive (320px – 2560px) / Mobile First
  [ ] Dark Mode vorhanden und korrekt
  [ ] Kontrast ≥ 4.5:1 (WCAG AA)
  [ ] Focus-States sichtbar
  [ ] Touch-Targets ≥ 44×44px
  [ ] prefers-reduced-motion implementiert
  [ ] Kein Emoji als Icon (SVG only)

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
