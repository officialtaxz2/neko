# n.eko — Design Overhaul Fork

> **Fork von:** [m1k1o/neko](https://github.com/m1k1o/neko) — selbst gehosteter virtueller Browser via Docker + WebRTC.
> **Dieses Repo:** `officialtaxz2/neko` — eigenständiger Fork mit vollständigem UI/UX-Redesign.
> **Upstream-PRs:** Keine — dieser Fork lebt als eigenständiges Produkt.

---

## Was wurde verbessert?

Das Original-Frontend war ein **Discord-Klon** (`#36393f`-Palette, Whitney-Font, keine eigene Designsprache). Dieser Fork ersetzt es vollständig durch ein modernes Design-System.

| Bereich | Verbesserungen | Status |
|---|---|---|
| **Design-System** | CSS Custom Properties durchgehend; Fluid Type `clamp()`; 8pt-Spacing; Shadow/Radius/Transition-Tokens; Dark- & Light-Mode mit Theme-Toggle | ✅ |
| **Typografie** | Fontshare CDN — Cabinet Grotesk (Display) + Satoshi (Body) + JetBrains Mono; Whitney vollständig entfernt | ✅ |
| **Glassmorphism** | Sidebar, Header, Login-Card, Notifications — `backdrop-filter: blur()` + token-basierte Oberflächen | ✅ |
| **Login-Page** | Vollbild-Page (ersetzt Modal-Overlay); Glassmorphism-Card; Spotlight-Gradient + Dot-Grid; semantisches HTML; Inline-Fehleranzeige; SystemDialog-Overlay; Auto-Login via URL-Params | ✅ |
| **Sidebar** | 4-Tab-Leiste (Users · Chat · Files · Settings); Users + Chat gleichzeitig (50/50-Split); beide Panels standardmäßig offen; exklusive Settings/Files-Logik | ✅ |
| **Userlist** | Card-Layout mit Avatar + Name + dauerhaft sichtbaren Aktions-Buttons; vollständige Admin-Moderation (Kick/Ban/Mute/Give/Release/Take Controls); Ignore für alle User | ✅ |
| **Controls-Bar** | Höhe 125px → 64px; Member-Avatare entfernt; Touch-Targets ≥44px; Animated FPS/Bitrate-Counters | ✅ |
| **Settings** | Bento Grid (2-Spalten); Custom Toggles; Resolution-Dropdown (admin-only) direkt in Settings; Scroll-Sensitivity Range 1–20 (war 1–100), Default 5; Drag-Tooltip zeigt aktuellen Wert live über dem Slider-Thumb | ✅ |
| **Video-Player** | Resolution-Button aus Player entfernt (→ Settings); alle WebRTC/Event-Handler unberührt | ✅ |
| **Notifications** | Floating Card-Design (ersetzt Vollfarb-Banner); typ-spezifische Accent-Border + FA-Icon | ✅ |
| **SweetAlert2** | Vollständig auf CSS Custom Properties migriert — Dark/Light-Mode-Tokens wirken zur Laufzeit | ✅ |
| **Trackpad-Modus** | Virtueller relativer Touch-Cursor für Mobile; Tippen = Linksklick, Lang-halten (600ms) = Rechtsklick, Zwei Finger = Scroll; Toggle in Settings mit Touch-Only-Badge; i18n in allen 15 Locales | ✅ |
| **Bugfixes** | Cancel-Bug bei Kick/Ban/Mute (alle `$swal`-Dialoge); `SystemDialog` via Store statt `$swal`; i18n `side.users` in allen 15 Locale-Dateien; mobiler rechter Leerstreifen (`100vw` → `100%` + `overflow-x: hidden`); TS2339 `changeResolution→screenSet` + `disconnect→logout`; abgeschnittene Select-Dropdowns in Bento-Half-Cards (`.row--select` stacking, 2× behoben) | ✅ |

---

## Dev-Umgebung

```bash
# Repo klonen + Backend starten
git clone https://github.com/officialtaxz2/neko.git && cd neko
docker compose up -d

# Frontend Dev-Server (Hot Reload)
cd client/dev && ./serve -i   # → http://localhost:3001

# Fertiges Image bauen
./build -r my-neko
./build -a firefox -b ghcr.io/m1k1o/neko/base:latest -r my-neko
docker compose up --force-recreate
```

---

## Design-System Tokens

```css
/* Dark (Standard) */
--color-bg:             hsl(220, 15%, 9%);
--color-surface:        hsl(220, 13%, 13%);
--color-surface-2:      hsl(220, 11%, 16%);
--color-surface-offset: hsl(220, 10%, 20%);
--color-primary:        hsl(213, 90%, 62%);   /* Blue-Akzent */
--color-text:           hsl(220, 10%, 86%);
--color-text-muted:     hsl(220,  8%, 50%);
--color-border:         hsl(220,  9%, 23%);
--color-error:          hsl(  4, 68%, 52%);
--color-success:        hsl(142, 50%, 48%);
--color-warning:        hsl( 36, 92%, 52%);

/* Light ([data-theme="light"]) */
--color-bg:             hsl(220, 20%, 97%);
--color-surface:        hsl(220, 18%, 100%);
--color-border:         hsl(220, 10%, 84%);
--color-text:           hsl(220, 20%, 12%);
--color-primary:        hsl(213, 90%, 44%);   /* dunkler für WCAG auf Weiß */

/* Fonts */
--font-display: 'Cabinet Grotesk', 'Helvetica Neue', sans-serif;
--font-body:    'Satoshi', 'Inter', sans-serif;
--font-mono:    'JetBrains Mono', 'Fira Code', monospace;

/* Fluid Type */
--text-xs:   clamp(0.75rem,  0.70rem + 0.25vw, 0.875rem);
--text-sm:   clamp(0.875rem, 0.80rem + 0.35vw, 1rem);
--text-base: clamp(1rem,     0.95rem + 0.25vw, 1.125rem);
--text-lg:   clamp(1.125rem, 1.00rem + 0.75vw, 1.5rem);
--text-xl:   clamp(1.5rem,   1.20rem + 1.25vw, 2.25rem);

/* Layout */
$menu-height:     40px;
$controls-height: 64px;
$side-width:      400px;

/* Spacing (8pt), Shadows, Radius, Transitions — siehe _variables.scss */
```

---

## Roadmap

> ⬜ = Geplant · 🔄 = In Arbeit · ✅ = Fertig

### Virtueller Trackpad-Modus (Mobile)

**Ziel:** Touch-Events werden im Trackpad-Modus relativ interpretiert — wie ein Laptop-Touchpad. Tippen = Linksklick, Lang-halten (600ms) = Rechtsklick, Zwei Finger = Scroll. Der originale absolute Modus (1:1 Touch → Mausposition) bleibt als Default und wird nicht angetastet.

**Schwierigkeit:** Niedrig–Mittel · **Aufwand:** ~3–5h

| # | Aufgabe | Datei(en) | Status |
|---|---|---|---|
| M1 | `trackpad_mode: boolean`-State + Mutation + Getter in Settings-Store | `client/src/store/settings.ts` | ✅ |
| M2 | `virtualTrackpad.ts` — isoliertes Utility-Modul (State, Init, Touch-Handlers, Clamp-Logik) | `client/src/utils/virtualTrackpad.ts` | ✅ |
| M3 | `onTouchHandler` in `video.vue` um Trackpad-Zweig erweitern; Original-Zweig bleibt unverändert | `client/src/components/video.vue` | ✅ |
| M4 | Toggle-Switch in Settings-Panel einfügen (nach Scroll-Invert); Badge signalisiert Touch-Only | `client/src/components/settings.vue` | ✅ |
| M5 | i18n-Strings (`trackpad_mode`, `trackpad_mode_description`, `trackpad_mode_mobile_hint`, `mobile`, `desktop_only_inactive`) in alle 15 Locale-Dateien | `client/src/locale/*.ts` (15 Dateien) | ✅ |

#### Implementierungsdetails

**Store (M1)** — analog zu `scroll_invert`; accessor-Pattern: `this.$accessor.settings.trackpad_mode`. LocalStorage-Persistenz via `get`/`set` aus `~/utils/localstorage`.

**`virtualTrackpad.ts` (M2)** — exportiert:
- `initVirtualCursor(srcW, srcH)` — setzt virtuellen Cursor in Quellauflösungs-Mitte (lazy, beim ersten `touchstart`)
- `handleVirtualTouchStart(touch, onMouseMove, onMouseDown, onMouseUp)` — startet Long-Press-Timer (600ms), sendet initiale Cursor-Position
- `handleVirtualTouchMove(touch, sensitivity, onMouseMove)` — Delta dx/dy berechnen, Position clamp auf `[0, srcW] / [0, srcH]`, `touchTotalDistance` akkumulieren
- `handleVirtualTouchEnd(onMouseDown, onMouseUp)` — Tap (<10px) → Linksklick; Long-Press (600ms, kein Move >10px) → Rechtsklick; Drag → nur MouseUp
- `resetVirtualTrackpad()`, `getVirtualCursorPosition()` — für späteres Cursor-Overlay
- Konstanten: `LONG_PRESS_DELAY_MS = 600`, `CLICK_THRESHOLD_PX = 10`

**`video.vue` (M3)** — Guard-Reihenfolge im neuen `onTouchHandler`:
```
1. !hosting || locked → return
2. trackpad_mode && is_touch_device → Trackpad-Zweig (sendData direkt in Quellkoordinaten)
3. else → original absoluter Zweig (unverändert)
```
Zwei-Finger-Scroll im Trackpad-Zweig: `touchmove` mit `e.touches.length === 2` → `sendData('wheel', { x:0, y:±1 })`. Klassenfeld `_lastTwoFingerY = 0` nötig.

> **Koordinaten:** Im Trackpad-Modus wird `$client.sendData('mousemove', {x, y})` direkt aufgerufen (virtuelle Cursor-Koordinaten liegen bereits in Quellauflösung). Kein `sendMousePos`-Umweg.

**Settings-UI (M4)** — Toggle nach dem Scroll-Invert-Block; `is_touch_device`-Getter via `'ontouchstart' in window && navigator.maxTouchPoints > 0`; bei Desktop: `opacity: 0.75`, grauer „Nur Touch"-Badge; bei Touch: grüner „Mobil"-Badge.

**Nicht in Scope (absichtlich):** Visueller Cursor-Overlay, einstellbare Sensitivität-Slider, Pinch-to-Zoom, PointerEvent-Refactoring — alle als optionale Follow-ups definiert.

#### Testfälle

| Szenario | Erwartetes Verhalten |
|---|---|
| Desktop, Toggle OFF | Maus funktioniert unverändert |
| Desktop, Toggle ON | Keine Auswirkung (Guard `!is_touch_device`) |
| Mobil, Toggle OFF | Touch 1:1 auf Position gemappt (Original) |
| Mobil, Toggle ON — Wischen | Cursor bewegt sich relativ; startet mittig |
| Mobil, Toggle ON — Tap | Linksklick an virtueller Cursor-Position |
| Mobil, Toggle ON — Lang-Halten (600ms) | Rechtsklick wird gesendet |
| Mobil, Toggle ON — Zwei Finger | Scroll-Event wird gesendet |
| Mobil, Toggle ON — über Rand wischen | Cursor stoppt am Rand (Clamp) |

---

## Bug-Fix-Phase

> Laufend dokumentierte Bugfixes nach Abschluss der Feature-Roadmap.

| # | Bug | Ursache | Fix | Commit |
|---|---|---|---|---|
| B1 | Kick/Ban/Mute Cancel-Bug | `$swal`-Promise warf bei Abbrechen, kein Guard | Guard auf `.isDismissed` + `SystemDialog` via Store | [7d92b21](https://github.com/officialtaxz2/neko/commit/7d92b21f738552317e0ad8d83ab5bb0d12e728cc) |
| B2 | i18n `side.users` fehlte in 14 Locales | Key nur in `en.ts` vorhanden | Key in alle 15 Locale-Dateien nachgetragen | [7d92b21](https://github.com/officialtaxz2/neko/commit/7d92b21f738552317e0ad8d83ab5bb0d12e728cc) |
| B3 | Mobiler rechter Leerstreifen | `max-width: 100vw` schließt Scrollbar-Gutter ein | `width/max-width: 100%` + `overflow-x: hidden` auf `html` + `body` | [770fd81](https://github.com/officialtaxz2/neko/commit/770fd81fb0e9e6d89073f07b6aca832222b2f4e6) |
| B4 | TS2339: `changeResolution` + `disconnect` existieren nicht | Store-Accessor-API hatte sich geändert | `changeResolution → screenSet`, `disconnect → logout` | [9f050be](https://github.com/officialtaxz2/neko/commit/9f050be665133472de7e19241ae124f50ae71546) |
| B5 | Select-Dropdowns (Keyboard Layout, Resolution) in Bento-Half-Cards abgeschnitten | `justify-content: space-between` in halber Kartenbreite; `:has(.select)` fiel bei Rewrite weg | `.row--select` explizit im Template; `flex-direction: column` + `width: 100%` auf `.select` | [c301d46](https://github.com/officialtaxz2/neko/commit/c301d46d8c986ea9eb774eda1909ffb25a902f02) → re-fix [1cd38e8](https://github.com/officialtaxz2/neko/commit/1cd38e8e8208ef11634bfab69fde9627a6c5ba1b) |

---

## Abnahme-Checkliste

```
[ ] Alle Views responsive (320px – 2560px)
[ ] Dark Mode · Light Mode — Kontrast, Icons, Selects prüfen
[ ] Kontrast ≥ 4.5:1 (WCAG AA) · Focus-States · Touch-Targets ≥ 44×44px
[ ] prefers-reduced-motion implementiert
[ ] Kein hardcodierter Farbwert (kein #36393f, keine alten $background-*-Vars)
```

---

## Upstream

Original: [m1k1o/neko](https://github.com/m1k1o/neko) · Docs: [neko.m1k1o.net](https://neko.m1k1o.net/)
