# n.eko — Design Overhaul Fork

> **Fork of:** [m1k1o/neko](https://github.com/m1k1o/neko) — self-hosted virtual browser via Docker + WebRTC.
> **This repo:** `officialtaxz2/neko` — independent fork with a complete UI/UX redesign.
> **Upstream PRs:** None — this fork lives as a standalone product.

---

## What was improved?

| Area | Improvements | Status |
|---|---|---|
| **Design System** | CSS Custom Properties throughout; fluid type `clamp()`; 8pt spacing; shadow/radius/transition tokens; dark & light mode with theme toggle | ✅ |
| **Typography** | Fontshare CDN — Cabinet Grotesk (display) + Satoshi (body) + JetBrains Mono; Whitney fully removed | ✅ |
| **Glassmorphism** | Sidebar, header, login card, notifications — `backdrop-filter: blur()` + token-based surfaces | ✅ |
| **Login Page** | Full-screen page; glassmorphism card; spotlight gradient + dot grid; inline error display; SystemDialog overlay; auto-login via URL params | ✅ |
| **Sidebar** | 4-tab bar (Users · Chat · Files · Settings); Users + Chat simultaneously (50/50 split); exclusive Settings/Files logic | ✅ |
| **User List** | Card layout; permanently visible action buttons; full admin moderation (Kick/Ban/Mute/Give/Release/Take Controls); Ignore | ✅ |
| **Controls Bar** | Height 125px → 64px; member avatars removed; touch targets ≥44px; animated FPS/bitrate counters | ✅ |
| **Settings** | Bento grid (2-column); custom toggles; resolution dropdown (admin-only); scroll sensitivity 1–20 (default 5); live tooltip on slider | ✅ |
| **Video Player** | Resolution button removed from player (→ Settings); WebRTC/event handlers untouched | ✅ |
| **Notifications** | Floating card design; type-specific accent border + FA icon | ✅ |
| **SweetAlert2** | Fully migrated to CSS Custom Properties | ✅ |
| **Trackpad Mode** | Relative touch cursor for mobile; tap = left click, long-press (600ms) = right click, two fingers = scroll; toggle in Settings with touch-only badge; i18n in all 15 locales | ✅ |
| **Bug Fixes** | Kick/Ban/Mute cancel bug; i18n `side.users` in all 15 locales; mobile right dead-space strip; TS2339 (`changeResolution→screenSet`, `disconnect→logout`); clipped select dropdowns in bento half-width cards | ✅ |

---

## Design System Tokens

```css
/* Dark (default) */
--color-bg:             hsl(220, 15%,  9%);
--color-surface:        hsl(220, 13%, 13%);
--color-surface-2:      hsl(220, 11%, 16%);
--color-surface-offset: hsl(220, 10%, 20%);
--color-primary:        hsl(213, 90%, 62%);
--color-text:           hsl(220, 10%, 86%);
--color-text-muted:     hsl(220,  8%, 50%);
--color-border:         hsl(220,  9%, 23%);
--color-error:          hsl(  4, 68%, 52%);
--color-success:        hsl(142, 50%, 48%);
--color-warning:        hsl( 36, 92%, 52%);

/* Light ([data-theme="light"]) */
--color-bg:      hsl(220, 20%, 97%);
--color-surface: hsl(220, 18%, 100%);
--color-border:  hsl(220, 10%, 84%);
--color-text:    hsl(220, 20%, 12%);
--color-primary: hsl(213, 90%, 44%);

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

/* Spacing (8pt), shadows, radius, transitions — see _variables.scss */
```

---

## Dev Setup

```bash
git clone https://github.com/officialtaxz2/neko.git && cd neko
docker compose up -d

# Frontend dev server (hot reload)
cd client/dev && ./serve -i   # → http://localhost:3001

# Build final image
./build -r my-neko
docker compose up --force-recreate
```

---

## Roadmap

> ⬜ = Planned · 🔄 = In Progress · ✅ = Done

*New roadmap items to follow.*

---

## Acceptance Checklist

```
[ ] All views responsive (320px – 2560px)
[ ] Dark mode · light mode — check contrast, icons, selects
[ ] Contrast ≥ 4.5:1 (WCAG AA) · focus states · touch targets ≥ 44×44px
[ ] prefers-reduced-motion implemented
[ ] No hardcoded color values (no #36393f, no legacy $background-* vars)
```

---

Original: [m1k1o/neko](https://github.com/m1k1o/neko) · Docs: [neko.m1k1o.net](https://neko.m1k1o.net/)
