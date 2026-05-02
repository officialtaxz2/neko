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

> **Status legend:** ⬜ Planned · 🔄 In Progress · ✅ Done · ⏳ Needs decision
>
> **Implementation order:** Phase 1 → Phase 2 → Phase 3 → Phase 4.
> Each phase is atomic: commit every feature separately. Phase 3.3 requires 3.2. Phase 4.2 requires an architecture decision (see below).
>
> **Ground rules for agents & developers:**
> - All styles **must** use existing CSS Custom Properties (`var(--color-*)`, `var(--space-*)`, `var(--radius-*)`, `var(--transition-interactive)`). No hardcoded hex, px, or rem values.
> - Every animated feature **must** respect `prefers-reduced-motion: reduce`. CSS animations are globally covered by `base.css`. JS-driven loops (`requestAnimationFrame`, `mousemove`) must be guarded manually:
>   ```js
>   const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
>   if (!prefersReduced) { /* start animation */ }
>   ```
> - All new components must be tested in **both dark and light mode**.
> - Dark mode: Glassmorphism opacity slightly lower; shadows need lower opacity or a subtle light-glow for elevation.
> - Light mode: Glassmorphism opacity slightly higher (~20% vs. 15% for self-bubbles).

---

### Phase 1 — Quick Wins ⬜

All decisions finalized. No open questions. Each feature is independently committable.

| # | Feature | File | Effort | Notes |
|---|---------|------|--------|-------|
| 1.1 | Animated Counters FPS/Bitrate — Verify & Debug | `controls.vue` | ~0 | Already implemented. Diagnose only: verify counters appear when `playing === true` and `displayFps > 0`. Some browsers hide them without an active video track — no code fix needed. |
| 1.2 | Kinetic Typography — Login | `login.vue` | S | One-time weight-pulse animation on the `n.eko` wordmark at pageload. Cabinet Grotesk variable font already loaded. Wrap each letter in `<span>`, staggered `animation-delay` (40ms each), `font-variation-settings: 'wght' 900 → 400`, 600ms ease-out. Trigger: on-mount only. |
| 1.3 | Dot Grid Background — Settings Panel | `settings.vue` | S | SVG dot grid via `::before` pseudo-element on `.neko-settings`. `opacity: 0.04` — keep it subtle. No screen-wide grid. |
| 1.4 | Parallax / Depth — Login Screen | `login.vue` | S | `mousemove` on `.login` container → normalised `{ x, y }` (−0.5 to 0.5) written as `--mx` / `--my`. Dot grid translates 12px, login card tilts max ±4° via `perspective: 1000px`. |
| 1.5 | Gamification — Controls Handoff | `controls.vue` | S | On receive (`hosting: false → true`): `burst` keyframe on keyboard icon + 2s inline badge "You have the controls" fades out. On release: opacity flash `1 → 0.4 → 1`. Do **not** reuse the existing `shake` keyframe. |

<details>
<summary>Phase 1 — CSS/JS snippets</summary>

**1.2 — Weight-Pulse keyframe**
```css
@keyframes weight-pulse {
  0%   { font-variation-settings: 'wght' 900; }
  100% { font-variation-settings: 'wght' 400; }
}
.neko-logo span {
  display: inline-block;
  animation: weight-pulse 600ms ease-out both;
}
.neko-logo span:nth-child(1) { animation-delay:   0ms; }
.neko-logo span:nth-child(2) { animation-delay:  40ms; }
.neko-logo span:nth-child(3) { animation-delay:  80ms; }
.neko-logo span:nth-child(4) { animation-delay: 120ms; }
.neko-logo span:nth-child(5) { animation-delay: 160ms; }
```

**1.3 — Dot Grid**
```css
.neko-settings { position: relative; }
.neko-settings::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, var(--color-border) 1px, transparent 1px);
  background-size: 20px 20px;
  opacity: 0.04;
  pointer-events: none;
  z-index: 0;
}
```

**1.4 — Parallax JS**
```js
const login = document.querySelector('.login');
login.addEventListener('mousemove', (e) => {
  const x = e.clientX / window.innerWidth  - 0.5;
  const y = e.clientY / window.innerHeight - 0.5;
  login.style.setProperty('--mx', x);
  login.style.setProperty('--my', y);
});
```
```css
@media (prefers-reduced-motion: reduce) {
  .dot-grid, .login-card { transform: none !important; }
}
```

**1.5 — Burst & Fade keyframes**
```css
@keyframes burst {
  0%   { transform: scale(1);   filter: drop-shadow(0 0 0px  var(--color-primary)); }
  40%  { transform: scale(1.4); filter: drop-shadow(0 0 8px  var(--color-primary)); }
  100% { transform: scale(1);   filter: drop-shadow(0 0 0px  var(--color-primary)); }
}
@keyframes fade-out {
  0%   { opacity: 1; transform: translateY(0); }
  100% { opacity: 0; transform: translateY(-6px); }
}
```
</details>

---

### Phase 2 — Core Features ⬜

All decisions finalized. Implement in order: 2.1 → 2.2 → 2.3 → 2.4.

| # | Feature | File | Effort | Notes |
|---|---------|------|--------|-------|
| 2.1 | Collapsible Sections — Settings | `settings.vue` | M | ~20 bento cards grouped into 4 collapsible sections: **Video & Audio**, **Input & Controls**, **Appearance**, **Admin** (admin-only). Default: all open. State persisted in `localStorage` per group (`neko_settings_group_video`, etc.). Chevron icon on each header. |
| 2.2 | Glassmorphism — Chat Bubbles | `chat.vue` | M | Self-bubbles: `--color-primary` 15% tint + `backdrop-filter: blur(8px)`. Others: `--color-surface-2` 80% + blur. System messages (join/leave): flat, `--color-text-muted`, no blur. Sidebar already has `blur(12px)` — increase bubble opacity slightly to avoid readability issues. |
| 2.3 | Split-Screen Sidebar — Transition Animation | `app.vue` | M | Animate the Chat/Users height split via `grid-template-rows: 0fr/1fr` transition (280ms, spring easing). Replaces the current instant size switch. |
| 2.4 | Kinetic Typography — Header | `header.vue` | S | `letter-spacing: -0.05em → 0` animation (600ms ease-out) on the room name. Triggers on every room join via `@Watch('roomName')` — add/remove CSS class with `$nextTick`. |

<details>
<summary>Phase 2 — CSS/JS snippets</summary>

**2.1 — Collapsible grid animation**
```css
.settings-group-content {
  display: grid;
  grid-template-rows: 1fr;
  overflow: hidden;
  transition: grid-template-rows 280ms cubic-bezier(0.16, 1, 0.3, 1);
}
.settings-group-content.collapsed { grid-template-rows: 0fr; }
.settings-group-content > * { min-height: 0; }
```
```js
const key = `neko_settings_group_${groupId}`;
const isOpen = localStorage.getItem(key) !== 'false'; // default: true
localStorage.setItem(key, String(!isOpen));
```

**2.2 — Chat bubble glass**
```css
.chat-message.is-self .bubble {
  background: color-mix(in srgb, var(--color-primary) 15%, transparent);
  backdrop-filter: blur(8px);
  border: 1px solid color-mix(in srgb, var(--color-primary) 25%, transparent);
}
[data-theme="light"] .chat-message.is-self .bubble {
  background: color-mix(in srgb, var(--color-primary) 20%, transparent);
}
.chat-message:not(.is-self) .bubble {
  background: color-mix(in srgb, var(--color-surface-2) 80%, transparent);
  backdrop-filter: blur(8px);
  border: 1px solid color-mix(in srgb, var(--color-border) 40%, transparent);
}
.chat-message.is-system .bubble {
  background: transparent;
  color: var(--color-text-muted);
  border: none;
  backdrop-filter: none;
}
.chat-input-wrapper {
  background: color-mix(in srgb, var(--color-surface) 70%, transparent);
  backdrop-filter: blur(8px);
  border: 1px solid color-mix(in srgb, var(--color-border) 40%, transparent);
}
```

**2.3 — Sidebar split animation**
```css
.page-container {
  display: grid;
  transition: grid-template-rows 280ms cubic-bezier(0.16, 1, 0.3, 1);
}
.page-container.chat-only  { grid-template-rows: 0fr 1fr; }
.page-container.split       { grid-template-rows: 1fr 1fr; }
.page-container.users-only  { grid-template-rows: 1fr 0fr; }
.page-container > * { min-height: 0; overflow: hidden; }
```

**2.4 — Letter-open keyframe**
```css
@keyframes letter-open {
  0%   { letter-spacing: -0.05em; opacity: 0.7; }
  100% { letter-spacing: 0;       opacity: 1; }
}
.room-name.animate {
  animation: letter-open 600ms ease-out both;
}
```
```js
@Watch('roomName')
onRoomNameChange() {
  this.animating = false;
  this.$nextTick(() => { this.animating = true; });
}
```
</details>

---

### Phase 3 — Layout Rewrites ⬜

Larger structural changes. Implement in order: 3.1 → 3.2 → 3.3. Feature 3.3 depends on 3.2.

| # | Feature | File | Effort | Notes |
|---|---------|------|--------|-------|
| 3.1 | Split-Screen Login | `login.vue` | L | Desktop: 50/50 CSS Grid — left branding panel (`--color-surface` bg + `n.eko` wordmark + tagline "Your browser. Your session." + "Self-hosted · Open Source"), right: existing glassmorphism card unchanged. Spotlight gradient + dot grid on left panel only. Mobile ≤768px: single column, branding panel hidden. |
| 3.2 | User-Count Badge — Tab Bar | `sidebar.vue` | M | Pill badge next to the "Users" tab label. Background `--color-surface-offset`, text `--color-text-muted`, `--radius-full`. On user join/leave: `scale(1.3) → scale(1)` bump animation (300ms spring). |
| 3.3 | Animated Counter — User Count | `sidebar.vue` | S | **Requires 3.2.** Lerp-interpolation (factor 0.18) on the badge count, consistent with FPS counters in `controls.vue`. Fallback: `scale` bump via `.bump` class is acceptable for small numbers. `prefers-reduced-motion` → instant jump. |

<details>
<summary>Phase 3 — CSS/JS snippets</summary>

**3.1 — Login layout**
```css
.login-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 100dvh;
}
@media (max-width: 768px) {
  .login-layout { grid-template-columns: 1fr; }
  .login-branding-panel { display: none; }
}
.login-branding-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: var(--space-16);
  background: var(--color-surface);
  position: relative;
  overflow: hidden;
}
```

**3.1 — Branding panel content**
```
[n.eko wordmark]

Your browser.
Your session.

🔒  Self-hosted · Open Source
```

**3.2 — Badge CSS**
```css
.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 var(--space-1);
  border-radius: var(--radius-full);
  background: var(--color-surface-offset);
  color: var(--color-text-muted);
  font-size: var(--text-xs);
  font-variant-numeric: tabular-nums;
}
.tab-badge.bump {
  animation: badge-bump 300ms cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes badge-bump {
  0%   { transform: scale(1); }
  40%  { transform: scale(1.3); }
  100% { transform: scale(1); }
}
```
</details>

---

### Phase 4 — Complex Features ⏳

Technically analysed. A final decision is required before implementation.

#### 4.1 Connection Quality Indicator

**File:** `controls.vue` · **Effort:** M

3-step signal bar (like a Wi-Fi icon) in the controls bar showing live connection quality. Uses the existing `RTCPeerConnection.getStats()` call — no additional network requests needed. Rendered as inline SVG.

**Proposed thresholds — confirm before implementing:**

| Status | Condition |
|--------|-----------|
| 🟢 Good | RTT < 80ms **and** packetLoss < 1% |
| 🟡 Fair | RTT 80–200ms **or** packetLoss 1–5% |
| 🔴 Poor | RTT > 200ms **or** packetLoss > 5% |

> ⏳ **Open decision:** Confirm or adjust thresholds.

---

#### 4.2 Chat Message Reactions

**File:** `chat.vue` (+ possibly Go backend) · **Effort:** M–L

Hover over a message reveals a mini reaction bar (SVG icons: 👍 / ❤️ / 😂 — not emoji strings). Counter animates via `count-up` (synergistic with animated counters from Phase 1/3).

**Architecture options:**

| Option | Description | Effort | Trade-off |
|--------|-------------|--------|----------|
| **A — Client-only** | Reactions local only, no broadcast | XS | Useless in real group sessions |
| **B — Chat-channel simulation** | Reaction = special message `__reaction:messageId:emoji`, filtered client-side and rendered as badge on the original message | M | Hack, but no backend changes needed |
| **C — Dedicated WS message type** | Clean protocol extension | L | Requires Go backend changes |

**Recommendation for this fork:** Option B — no backend changes required.

> ⏳ **Open decisions:**
> 1. Choose Option B or C.
> 2. Should the host-queue display be part of this feature or separate?

---

### Dropped Features

| Feature | Reason |
|---------|--------|
| **Bento Grid 3-column extension** | Sidebar has a fixed width — 3 columns would squeeze cards or require a wider sidebar. Both undesirable. Optimise the existing grid within the fixed sidebar width instead. |
| **Mobile Fullscreen Toggle** | Already present as an on-screen button. No action needed. |

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
