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

Design Overhaul **Phase 1–5 vollständig abgeschlossen** ✅ · Testphase abgeschlossen ✅ (30.04.2026)

| Datei | Wichtigste Änderungen | Status |
|---|---|---|
| `_variables.scss` | Discord-Palette entfernt; Blue-Akzent `hsl(213,90%,62/44%)` Dark/Light; Fluid Type `clamp()`; 8pt-Grid; Shadow/Radius/Transition-Tokens | ✅ |
| `main.scss` | Fontshare CDN (Satoshi + Cabinet Grotesk); globale Resets; `:focus-visible`; `prefers-reduced-motion` | ✅ |
| `app.vue` | Theme-Toggle (`data-theme`); Sidebar-Transition slide+fade; Split-Screen Blue-Glow; room-container Glassmorphism; Sticky Header + Controls (`position:sticky`); `.is-scrolled`-Elevation | ✅ |
| `header.vue` | Theme-Toggle-Button (SVG); Icon Micro-Animations (`scale`); funktionierender Glassmorphism-Diagonal-Gradient; CSS Tokens durchgehend | ✅ |
| `side.vue` | Glassmorphism (`backdrop-filter:blur(12px)`); Pill-Tabs; Tab-Inhalt-Transition `out-in` | ✅ |
| `chat.vue` | Pill-Username-Badges; Message-Hover; Textarea Token-Focus; Skeleton Loading (4 Shimmer-Messages) | ✅ |
| `members.vue` | 4-Zustand-Status-Dots (online/away/busy/offline); Avatar Hover-Scale; Skeleton Loading | ✅ |
| `menu.vue` | Icons auf `--color-text-muted`; `<select>` vollständig auf CSS Custom Properties migriert (Light-Mode-fix) | ✅ |
| `controls.vue` | Touch-Targets ≥44px; Micro-Animations; `.faded` → `--color-text-muted`; `fa-pause`/`fa-play` (Outline); Animated Counters FPS/Bitrate via `getStats()`+Lerp+`requestAnimationFrame` | ✅ |
| `settings.vue` | Bento Grid (2-Spalten, 5 Cards); Custom Toggle Switches (Spring-Easing, Blue-Akzent, `--color-border` Off-Track); LOG OUT Ghost-Button | ✅ |
| `connect.vue` | Glassmorphism-Dialog; radialer Spotlight-Gradient; Dot-Grid-Hintergrund; Touch-Targets ≥44px | ✅ |
| `neko/base.ts` | Public getter `peerConnection` auf `BaseClient` (exponiert `protected _peer` sauber für Stats-Polling) | ✅ |
| `video.vue` | WebRTC-Video + Maus/Tastatur-Overlay — Event-Handler unberührt | ⬜ |

---

## Roadmap

> ⬜ = Geplant · 🔄 = In Arbeit · ✅ = Fertig

### Übersicht

| # | Feature | Betroffene Dateien | Status |
|---|---|---|---|
| R1 | Sidebar: 3. Tab „Users“ ergänzen | `side.vue` | ⬜ |
| R2 | Neue `userlist.vue` — User-Einträge + Moderationsaktionen | `userlist.vue` (neu erstellen) | ⬜ |
| R3 | Multiple-Choice Sidebar: Chat + Users gleichzeitig (50/50) | `side.vue` | ⬜ |
| R4 | User-Avatare aus Controls-Bar entfernen | `members.vue`, `app.vue` | ⬜ |
| R5 | Resolution-Button aus Video-Player entfernen | `video.vue` | ⬜ |
| R6 | Resolution-Dropdown in Settings ergänzen | `settings.vue` | ⬜ |
| R7 | *(Platzhalter — noch zu definieren)* | — | ⬜ |
| R8 | *(Platzhalter — noch zu definieren)* | — | ⬜ |
| R9 | *(Platzhalter — noch zu definieren)* | — | ⬜ |

---

### R1–R3: Sidebar Redesign + Userlist

**Ziel:** Die Sidebar erhält einen dritten Tab „Users“. Chat und Users können gleichzeitig aktiv sein (Multiple-Choice). Settings ist exklusiv — entweder ist Chat/Users aktiv, oder Settings, nie beides.

**Tab-Logik in `side.vue`:**

| Zustand | Sidebar-Inhalt |
|---|---|
| Nur Chat aktiv | Chat nimmt 100% der Sidebar-Höhe |
| Nur Users aktiv | Userlist nimmt 100% der Sidebar-Höhe |
| Chat + Users aktiv | Chat 50% oben / Userlist 50% unten (`flex-direction: column`, beide `flex: 1`) |
| Settings aktiv | Settings nimmt 100% — Chat/Users-State wird nicht beibehalten (oder eingefroren) |

- Tab-Leiste: 3 Tabs — `fa-comment` Chat · `fa-users` Users · `fa-cog` Settings
- Chat und Users-Tabs sind **Toggle-Buttons** (unabhängig an/abwählbar), Settings-Tab ist **exklusiv** (deaktiviert Chat+Users-State beim Wechsel)
- Tab-Zustand: `activeChat: boolean`, `activeUsers: boolean`, `activeSettings: boolean` — Settings schließt die anderen aus
- Der Sidebar-Content-Bereich nutzt CSS `display: flex; flex-direction: column` — aktive Panels erhalten `flex: 1`
- Übergang zwischen Split und Single via `<transition>` (Höhe smooth mit `max-height`-Trick oder Vue `<transition-group>`)

**Neue `userlist.vue`:**

- Pro User eine Zeile: **Avatar** (kleines Pill-Badge wie in `chat.vue`) + **Username** + **Aktions-Icons** rechts
- Aktions-Icons (nur sichtbar für Admin/Host oder per Hover): `fa-eye-slash` Ignore · `fa-microphone-slash` Mute · `fa-gamepad` Give Controls · `fa-user-times` Kick · `fa-ban` Ban IP
- Icons: Touch-Target ≥44px, Micro-Animation (`scale(1.18)` Hover / `scale(0.88)` Active), `--color-text-muted` at rest → `--color-error` bei destruktiven Aktionen (Kick/Ban) on Hover
- Skeleton Loading: 4 Shimmer-Rows (konsistent mit `chat.vue` und `members.vue`)
- Bestehende Right-Click-Logik aus `members.vue` (Context-Menu) in `userlist.vue`-Aktionen übertragen; Right-Click in Controls-Bar entfällt damit
- CSS Tokens durchgehend, kein hardcodierter Wert

---

### R4: User-Avatare aus Controls-Bar entfernen

**Ziel:** Die User-Avatar-Bubbles, die aktuell in der Controls-Bar unterhalb des Video-Players angezeigt werden, werden komplett entfernt — User-Übersicht lebt ausschließlich in der Sidebar (Userlist, R2).

- `members.vue`: Komponente bleibt vorerst bestehen, wird aber nicht mehr in der Controls-Bar gerendert (Einbindung aus `app.vue` oder `controls.vue` entfernen)
- Prüfen ob `members.vue` danach noch benötigt wird oder in `userlist.vue` aufgeht
- Controls-Bar (`controls.vue` / `app.vue`): Avatar-Render-Slot entfernen; kein visuelles Loch — Spacing anpassen

---

### R5–R6: Resolution aus Video-Player → Settings

**Ziel:** Der Monitor-Icon-Button auf dem Video-Player (`video.vue`) wird entfernt. Die Auflösungsauswahl wird als neue Card in `settings.vue` integriert.

**R5 — `video.vue`:**
- Monitor-Icon-Button (Auflösungsauswahl) aus dem Player-Overlay entfernen
- Nur Button + zugehöriges Dropdown/Modal entfernen — Video-Event-Handler (Maus, Tastatur, WebRTC) nicht anfassen

**R6 — `settings.vue`:**
- Neue 6. Card im Bento Grid: **„Display“** (Icon: `fa-desktop`)
- Inhalt: Custom Dropdown-Select mit der Liste verfügbarer Auflösungen (aus vorhandener Neko-API)
- Preview-Label: zeigt aktuell aktive Auflösung im geschlossenen Zustand an (z.B. `1920 × 1080 @ 30 fps`)
- Styling: konsistent mit bestehenden Settings-Cards — CSS Custom Properties, kein hardcodierter Wert
- Dropdown-State und API-Aufruf analog zur bestehenden Resolution-Logik aus `video.vue` übernehmen

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

/* === SHADOWS, RADIUS, TRANSITIONS === */
--shadow-sm: 0 1px 2px rgba(30,40,60,0.08);   /* Dark: rgba(0,0,0,0.35) */
--shadow-md: 0 4px 12px rgba(30,40,60,0.12);  /* Dark: rgba(0,0,0,0.45) */
--shadow-lg: 0 12px 32px rgba(30,40,60,0.16); /* Dark: rgba(0,0,0,0.60) */

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
