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

Design Overhaul **Phase 1–5 vollständig abgeschlossen** ✅ · **R1–R6 vollständig abgeschlossen** ✅ · Bugfix-Session abgeschlossen ✅ (01.05.2026)

| Datei | Wichtigste Änderungen | Status |
|---|---|---|
| `_variables.scss` | Discord-Palette entfernt; Blue-Akzent `hsl(213,90%,62/44%)` Dark/Light; Fluid Type `clamp()`; 8pt-Grid; Shadow/Radius/Transition-Tokens; **R4:** `$controls-height` 125px → 64px (Avatar-Row entfernt) | ✅ |
| `main.scss` | Fontshare CDN (Satoshi + Cabinet Grotesk); globale Resets; `:focus-visible`; `prefers-reduced-motion` | ✅ |
| `app.vue` | Theme-Toggle (`data-theme`); Sidebar-Transition slide+fade; Split-Screen Blue-Glow; room-container Glassmorphism; Sticky Header + Controls (`position:sticky`); `.is-scrolled`-Elevation; **R4:** `neko-members` aus Controls-Bar entfernt (Import + Registrierung bereinigt); **Bugfix:** `.expanded` + `neko-side v-if` an `connected && side` gebunden | ✅ |
| `header.vue` | Theme-Toggle-Button (SVG); Icon Micro-Animations (`scale`); funktionierender Glassmorphism-Diagonal-Gradient; CSS Tokens durchgehend | ✅ |
| `side.vue` | Glassmorphism (`backdrop-filter:blur(12px)`); Pill-Tabs; Tab-Inhalt-Transition `out-in`; **R1+R3:** 4-Tab-Leiste (Chat · Users · Files · Settings); Toggle-Logik (Chat+Users unabhängig, Settings exklusiv); 50/50-Split mit `panel-divider` wenn beide aktiv; `panel-grow`-Transition (max-height); **R2:** `<neko-userlist>` ersetzt Placeholder | ✅ |
| `userlist.vue` | **R2 + Bugfix:** Card-Layout (Avatar 32px links; Zeile 1: Name + Crown; Zeile 2: Buttons dauerhaft sichtbar); **Alle User:** Ignore/Unignore (`fa-eye-slash/fa-eye`); **Admin:** Mute/Unmute · Release Controls (`fa-times-circle`) · Force Take Controls (`fa-hand-paper`) · Give Controls (`fa-gamepad`); **Host (non-admin):** Give Controls; **Admin, non-admin Target:** Kick · Ban; Controls-Buttons nur bei `!implicitHosting`; Skeleton Loading 2-zeilig (Name + Actions); `isCurrentHost`-Getter; kein Right-Click-Kontextmenü (alle Aktionen inline) | ✅ |
| `chat.vue` | Pill-Username-Badges; Message-Hover; Textarea Token-Focus; Skeleton Loading (4 Shimmer-Messages) | ✅ |
| `members.vue` | 4-Zustand-Status-Dots (online/away/busy/offline); Avatar Hover-Scale; Skeleton Loading — **R4:** nicht mehr in Controls-Bar gerendert, Datei bleibt vorerst erhalten | ✅ |
| `menu.vue` | Icons auf `--color-text-muted`; `<select>` vollständig auf CSS Custom Properties migriert (Light-Mode-fix) | ✅ |
| `controls.vue` | Touch-Targets ≥44px; Micro-Animations; `.faded` → `--color-text-muted`; `fa-pause`/`fa-play` (Outline); Animated Counters FPS/Bitrate via `getStats()`+Lerp+`requestAnimationFrame` | ✅ |
| `settings.vue` | Bento Grid (2-Spalten); Custom Toggle Switches; LOG OUT Ghost-Button; **R6:** neue „Display"-Card (admin-only) mit Resolution-`<select>` (`W × H @ R fps`); `resValue` getter/setter → `$accessor.video.screenSet()` | ✅ |
| `video.vue` | WebRTC-Video + Maus/Tastatur/Touch-Event-Handler unberührt; **R5:** `fa-desktop`-Overlay-Button entfernt; `<neko-resolution>` entfernt; `Resolution`-Import, `@Ref` + `openResolution()` bereinigt | ✅ |
| `connect.vue` | Glassmorphism-Dialog; radialer Spotlight-Gradient; Dot-Grid-Hintergrund; Touch-Targets ≥44px; **Bugfix:** `$swal` durch In-App `SystemDialog`-Overlay ersetzt; store-gesteuert via `client.systemDialog` | ✅ |
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
| R7 | Dedizierte Login-Page — `connect.vue` als Vollbild-Route gemäß DesignSystem.md | `connect.vue` → `login.vue`, `app.vue` | ⬜ |
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

### R7 — Dedizierte Login-Page ⬜

**Ziel:** `connect.vue` wird von einem modalen Overlay (das über dem Room-Layout schwimmt) zu einer eigenständigen Vollbild-Seite umgebaut. Der User sieht den Room **niemals**, solange er nicht eingeloggt ist — keine Layout-Bleed-Artefakte, keine Hintergrund-UI.

#### Motivation & Problem-Analyse

Aktuell (`connect.vue` als `position: fixed`-Overlay über `app.vue`):

- Der gesamte Room wird im DOM gerendert und geladen, auch wenn niemand eingeloggt ist
- Sidebar-State (`side = true` im Store/LocalStorage) beeinflusst das Layout hinter dem Overlay — sichtbar als Breiten-Artefakt in Header und Controls-Bar (bereits in Bugfix-Commit 01.05.2026 entschärft, aber nicht vollständig gelöst, da DOM-Struktur identisch bleibt)
- Das Login-Overlay schwimmt über einem bereits initialisierten Room — konzeptuell falsch und potenziell sicherheitsrelevant (DOM-Inspektion ermöglicht Einblick in Room-Struktur vor Login)
- `$swal`-Dialoge (Logged Out, Disconnected, Error) wurden durch In-App `SystemDialog` ersetzt — die Basis für eine saubere Entkopplung ist damit gelegt

**Gewünschter Zielzustand:** Solange `!connected`, rendert `app.vue` **ausschließlich** `<neko-login-page>` — kein Room-DOM, kein Header, keine Controls, keine Sidebar. Nach erfolgreichem Login wird `<neko-login-page>` ausgeblendet und der Room vollständig eingeblendet.

---

#### Architektur-Entscheidung: kein Vue Router

neko verwendet **keinen Vue Router** — das gesamte App-State-Management läuft über den Vuex-Store (`$accessor`). R7 folgt diesem Pattern: Der Login-Zustand wird ausschließlich über `$accessor.connected` gesteuert. Kein `vue-router`, keine Hash-Routing-Änderungen, keine neuen Abhängigkeiten.

---

#### Design-Anforderung: DesignSystem.md Prioritätsmatrix Kat. 1–5

> **Verbindlich:** Die Login-Page wird gemäß `DesignSystem.md` — **PriorityMatrix Kategorien 1–5** — vollständig umgesetzt. Kein Abweichen von diesen Vorgaben ohne explizite Freigabe.

Die fünf Kategorien der PriorityMatrix sind bei der Implementierung von `login.vue` **in dieser Reihenfolge** abzuarbeiten und zu gewährleisten:

| Kat. | Thema | Verbindliche Anforderungen für `login.vue` |
|---|---|---|
| **1** | **Farbe & Oberflächen** | Ausschließlich CSS Custom Properties aus `_variables.scss` (`--color-bg`, `--color-surface`, `--color-border`, etc.). Kein einziger hardcodierter Hex-/RGB-Wert. Dark Mode (Standard) + Light Mode via `[data-theme="light"]` — beide Zustände vollständig und korrekt. Glassmorphism-Card (`.window`) nutzt `--color-surface` + `backdrop-filter` + `--shadow-lg`. Spotlight-Gradient und Dot-Grid-Hintergrund ebenfalls token-basiert. |
| **2** | **Typografie** | `--font-display` (Cabinet Grotesk) für den Login-Titel (≥ `--text-xl`). `--font-body` (Satoshi) für Labels, Inputs, Buttons. Fluid Type Scale via `clamp()` — alle Größen referenzieren `--text-*`-Tokens. Keine hardcodierten `px`-Werte für Schriftgrößen. Mindestens 12px Floor (`--text-xs`). |
| **3** | **Spacing & Layout** | Alle Abstände über `--space-*`-Tokens (8pt-Grid). Kein `margin: 20px` o.ä. Die Login-Card ist vertikal + horizontal zentriert (`display: flex; align-items: center; justify-content: center` auf dem `.login-page`-Root). Auf Mobile (375px) füllt die Card maximal `calc(100vw - 2 × --space-4)` Breite ohne Overflow. |
| **4** | **Interaktion & Transitions** | Alle Hover-/Focus-States via `--transition-interactive` (180ms). `:focus-visible` mit `outline: 2px solid var(--color-primary)` sichtbar. Touch-Targets ≥ 44×44px (Buttons, Inputs). `prefers-reduced-motion`: alle Animationen (Bounce-Loader, Spotlight-Pulse, Dot-Grid-Fade) deaktivieren auf `@media (prefers-reduced-motion: reduce)`. Page-Transition (`page-fade`) zwischen Login-Page und Room ebenfalls motion-safe. |
| **5** | **Zugänglichkeit & Semantik** | Semantisches HTML: `<form>`, `<label>`, `<input>`, `<button type="submit">`. Jedes `<input>` hat ein assoziiertes `<label>` (kein `placeholder` als Ersatz). WCAG AA: Kontrast ≥ 4.5:1 für alle Texte auf allen Surfaces (Dark + Light). Fehlermeldungen (`--color-error`) inline unter dem betroffenen Feld. Keyboard-Navigation: Tab-Reihenfolge logisch (Displayname → Passwort → Submit). `aria-label` auf Icon-Buttons (Theme-Toggle falls vorhanden). |

> **Hinweis für Implementierer:** Die `DesignSystem.md` liegt nicht im Repository, sondern wird separat als Projekt-Referenz geführt. Bei Widerspruch zwischen dieser Spec und `DesignSystem.md` gilt stets `DesignSystem.md` als Primat-Dokument. Im Zweifelsfall vor der Implementierung klären.

---

#### Betroffene Dateien

| Datei | Aktion | Beschreibung |
|---|---|---|
| `client/src/components/connect.vue` | **Rewrite** | Vollbild-Login-Page (kein Overlay mehr); gemäß PriorityMatrix Kat. 1–5 |
| `client/src/app.vue` | **Patch** | Bedingte Render-Logik: `v-if="connected"` auf gesamte Room-Struktur; `<neko-login-page>` als Ersatz wenn `!connected` |
| `client/src/store/client.ts` | **Keine Änderung nötig** | `systemDialog`-State bereits vorhanden (Bugfix 01.05.2026) |
| `client/src/neko/index.ts` | **Keine Änderung nötig** | `setSystemDialog`-Calls bereits implementiert (Bugfix 01.05.2026) |

> **Umbenennung (empfohlen):** `connect.vue` → `login.vue`; Komponenten-Name `neko-connect` → `neko-login-page`; Import in `app.vue` entsprechend anpassen.

---

#### Schritt-für-Schritt-Implementierung

##### Schritt 1 — `app.vue`: Room hinter `v-if="connected"` legen

**Aktueller Stand in `app.vue` (vereinfacht):**

```html
<template>
  <div id="neko" :class="[!videoOnly && connected && side ? 'expanded' : '']">
    <template v-else>  <!-- v-if="!$client.supported" ... -->
      <main class="neko-main">       <!-- Header + Video + Controls -->
        ...
      </main>
      <transition name="side">
        <neko-side v-if="!videoOnly && connected && side" />
      </transition>
      <neko-connect v-if="!connected" />   <!-- Overlay über dem Room -->
      <neko-about v-if="about" />
      <notifications ... />
    </template>
  </div>
</template>
```

**Ziel-Stand nach R7:**

```html
<template>
  <div id="neko">
    <template v-if="!$client.supported">
      <neko-unsupported />
    </template>

    <!-- Login-Page: exklusiv wenn nicht verbunden -->
    <transition name="page-fade" mode="out-in">
      <neko-login-page v-if="!connected" key="login" />

      <!-- Room: nur wenn verbunden -->
      <div
        v-else
        key="room"
        class="neko-room-wrapper"
        :class="[!videoOnly && side ? 'expanded' : '']"
      >
        <main class="neko-main">  <!-- Header + Video + Controls unverändert -->
          ...
        </main>
        <transition name="side">
          <neko-side v-if="!videoOnly && side" />
        </transition>
        <neko-about v-if="about" />
        <notifications ... />
      </div>
    </transition>
  </div>
</template>
```

**Wichtige Detail-Änderungen in `app.vue`:**

- `id="neko"` bleibt auf dem Root-Div — CSS-Selektoren bleiben stabil
- `<neko-connect v-if="!connected" />` wird **entfernt**
- `<neko-login-page />` tritt als eigenständige Branch an (kein Overlay)
- `connected`-Checks auf `neko-side v-if` entfallen (implizit im `v-else`-Branch)
- Import: `import LoginPage from '~/components/login.vue'`
- Komponenten-Registrierung: `'neko-login-page': LoginPage`

**SCSS in `app.vue` — Page-Transition + Room-Wrapper:**

```scss
// Sanfter Fade: Login-Page ↔ Room
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity var(--transition-slow);
}
.page-fade-enter,
.page-fade-leave-to {
  opacity: 0;
}

// Room-Wrapper (trägt .expanded wenn Sidebar offen)
.neko-room-wrapper {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: row;
  overflow: hidden;

  &.expanded .neko-main {
    box-shadow: 4px 0 20px color-mix(in srgb, var(--color-primary) 10%, transparent);
  }
}
```

> **Alternativ (minimal invasiv):** `#neko` behält seine bestehende SCSS-Struktur identisch. Nur Template und `.expanded`-Bedingung ändern. Kein neuer Wrapper nötig wenn `#neko` bereits `flex-direction: row` + `overflow: hidden` trägt.

---

##### Schritt 2 — `connect.vue` → `login.vue` Rewrite

Die Datei wird von einem Overlay-Pattern zu einer **eigenständigen Vollbild-Page** umgebaut. Visuelle Sprache (Glassmorphism, Spotlight-Gradient, Dot-Grid) bleibt erhalten. Die PriorityMatrix Kat. 1–5 ist vollständig umzusetzen.

**Was bleibt (unverändert übernehmen):**

- `position: fixed; inset: 0` auf `.login-page` (Root-Element)
- Spotlight-Gradient + Dot-Grid-Hintergrund — token-basiert
- Glassmorphism `.window`-Card — `--color-surface`, `backdrop-filter`, `--shadow-lg`
- `SystemDialog`-Overlay (aus Bugfix 01.05.2026) — bleibt vollständig
- `autoPassword`/`autoUser`-URL-Param-Logik in `mounted()` — bleibt vollständig
- Bounce-Loader-Animation — bleibt; `prefers-reduced-motion` ergänzen falls fehlend
- i18n-Keys (`connect.login_title`, `connect.displayname`, etc.) — alle unverändert

**Was sich ändert:**

- Dateiname: `connect.vue` → `login.vue`
- Komponenten-Name: `neko-connect` → `neko-login-page` (Template + `@Component`-Decorator)
- Konzeptuell: kein Overlay mehr — die Komponente **ist** die einzige sichtbare Seite
- Alle Farben, Abstände, Schriftgrößen auf Token-Compliance prüfen (PriorityMatrix Kat. 1–2–3)
- Semantisches HTML verifizieren: `<form>`, `<label for>`, `<input id>`, `<button type="submit">` (Kat. 5)
- `backdrop-filter: blur(4px)` auf `.login-page`-Root optional (war Overlay-Effekt; als Page beibehalten oder entfernen)

---

##### Schritt 3 — Verifikation & Abnahme

Nach Implementierung folgende Szenarien manuell testen:

| Szenario | Erwartetes Verhalten |
|---|---|
| Frischer Seitenaufruf (nicht eingeloggt) | Nur Login-Page sichtbar; kein Room-DOM im DevTools Inspector |
| Login erfolgreich | Sanfter Fade-Out Login → Fade-In Room (`page-fade`, 300ms) |
| Sidebar war offen (LocalStorage `side=true`), dann ausgeloggt | Login-Page füllt **100% Breite** — kein Layout-Bleed |
| Logout via Settings | `systemDialog` (info) über Login-Page; OK → Login-Card zurück |
| Server-Disconnect / Kick | `systemDialog` (error) über Login-Page; OK → Login-Card zurück |
| `?pwd=xxx&usr=yyy` URL-Params | Auto-Login funktioniert (Logik in `mounted()` unverändert) |
| Light Mode | Login-Page in `[data-theme="light"]` korrekt — alle Tokens greifen |
| Mobile (375px) | Login-Card zentriert, kein Overflow, Touch-Targets ≥44px |
| `prefers-reduced-motion` aktiv | Bounce-Loader + Spotlight-Pulse statisch; `page-fade` sofort |
| WCAG-Kontrast | Alle Texte ≥ 4.5:1 auf ihren Surfaces (Dark + Light) |
| Keyboard-Navigation | Tab: Displayname → Passwort → Submit; Enter löst Submit aus |

---

#### Abgrenzung: Was R7 **nicht** ist

- **Kein eigener URL-Pfad / Route** (`/login` o.ä.) — neko hat keinen Router; bleibt so
- **Kein Server-Side Rendering / Auth-Guard** — Login-State bleibt rein client-seitig über WebSocket
- **Keine Änderung am Backend** — nur Frontend
- **Keine Änderung an i18n-Keys** — bestehende Schlüssel werden weiterverwendet
- **Kein visuelles Redesign des Formulars** — Glassmorphism, Spotlight-Gradient, Dot-Grid bleiben; nur Token-Compliance + Semantik sicherstellen

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
