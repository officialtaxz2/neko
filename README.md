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
| **Settings** | Bento Grid (2-Spalten); Custom Toggles; Resolution-Dropdown (admin-only) direkt in Settings | ✅ |
| **Video-Player** | Resolution-Button aus Player entfernt (→ Settings); alle WebRTC/Event-Handler unberührt | ✅ |
| **Notifications** | Floating Card-Design (ersetzt Vollfarb-Banner); typ-spezifische Accent-Border + FA-Icon | ✅ |
| **SweetAlert2** | Vollständig auf CSS Custom Properties migriert — Dark/Light-Mode-Tokens wirken zur Laufzeit | ✅ |
| **Bugfixes** | Cancel-Bug bei Kick/Ban/Mute (alle `$swal`-Dialoge); `SystemDialog` via Store statt `$swal`; i18n `side.users` in allen 15 Locale-Dateien | ✅ |

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

| # | Feature | Betroffene Dateien | Status |
|---|---|---|---|
| — | *(Neue Features hier eintragen)* | — | ⬜ |

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
