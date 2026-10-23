# The fork's GUI

This fork runs upstream's React client with its own look: the Obsidian palette with Cyber Mint, glass
bars and the Whitney font of the fork's earlier Vue client. Everything of that look lives here, so a
sync with upstream touches as few upstream files as possible.

- `theme.css`: the tokens. Imported at the end of `src/index.css`, so its `@theme` values replace
  upstream's (`bg-background-primary`, `text-style-primary`, `h-menu`, ...) and every upstream component
  picks up the palette. It adds the fork's own tokens (`text-pure`, `bg-mint`, `text-mint-ink`,
  `bg-glass`, `border-glass-line`, `bg-bar-room`, `rounded-control`, `ease-fluid`, ...), the font faces,
  the focus ring and the reduced-motion rule. The design thesis is at its top.
- `fonts/`: Whitney, as the Vue client shipped it.
- `layout/`: the pieces of the app shell that `src/app/App.tsx` uses in place of upstream's
  `src/layout/*`: `Header` (logo, status chips, side-panel toggle), `RoomBar` (members, controls,
  emotes in one row), `MobileDock` (the phone's bar under the picture), `Side` (the panel on glass) and
  `Toaster` (glass toasts).
- `room/`: fork components the layout composes (`MemberStrip`).
- Strings: the namespace `fork`, `src/locales/<lang>/fork.json` in every language, used as
  `t('fork:key')`.

Upstream files changed for the fork, each marked with a "the fork" comment: `src/app/App.tsx` (the
shell's classes and the pieces above), `src/index.css` (the import), `src/main.tsx` (the namespace),
`src/features/settings/Settings.tsx` (the language picker and the admin's lock switches),
`src/components/ui/button.tsx` (dark ink on mint), `eslint.config.js` (the layer rule).

Imports follow the layer rules (`eslint.config.js`): `src/fork` sits beside `src/layout`, it may use
features, components, state and the plugin registry; only the app shell imports it.

The visual check (DesignSystem.md, Gate 2) runs against a simulated server:
`node tools/screens.mjs <out-dir>` renders the login, the desktop with and without the side panel, a
narrow window, the phone upright with and without the panel, and the phone sideways.
