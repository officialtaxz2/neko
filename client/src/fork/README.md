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
- `room/`: fork components the layout composes (`MemberStrip`, and `MemberList`, the side panel's
  Members tab: everyone with their roles, a tap opening the member menu, so a phone reaches it too).
- `controls/`: `ControlPill` (who controls, in words, with the one action that fits: take, request,
  take over, give up, and the host's input pause; `pill.ts` decides its state), `Playback` (play,
  sound, the host's microphone) and `styles.ts` (the bar's and the dock's buttons).
- `tools/`: the tools that floated on upstream's picture, in one place as the Vue client had them
  (`useTools`): the desktop's `RoomTools` (fullscreen and the More popover), the phone's `MoreSheet`,
  both showing `MoreTools`, and `ClipboardPanel` (the remote clipboard with the recent texts).
- `settings/`: `Settings`, upstream's settings in the Vue client's groups (Stream, Input, Chat, Room
  for admins, Account with logging out after asking); upstream's `features/settings/Settings.tsx`
  stays as it is. A plugin's lock (file transfer) comes through the `room.lock` slot: a switch in the
  Room section and a chip in the `Header`, which leaves out the plugin's own header button.
- `connect/`: `ShareConnect`, the login screen of a view-only share link (`#/<token>`, see
  `src/state/viewonly.ts` and `docs/VIEW_ONLY_SHARING.md`); `layout/PassiveBar` is its playback bar.
- HLS (outside this folder, in their layers): `src/state/media.ts` (the choice, the transport, what
  the server offers), `src/state/diagnostics.ts` (the traces' read-only handle) and
  `src/features/media/` (the status on the picture, the settings' choice, the short select for a
  view-only bar); the transport itself is in the core (`core/src/transport/hls.ts`).
- Touch input (outside this folder too): `src/features/input/` (`useTouch`: a touch device or
  "force touch"; `InputSetting`: force touch, the trackpad and its pointer in the settings;
  `TrackpadCursor`: the trackpad's pointer drawn into the overlay; `KeyboardHelper` and `keys.ts`:
  the special keys' bar over the on-screen keyboard); the trackpad itself is in the core
  (`core/src/overlay.ts`, `setTrackpad`). The settings keep the Vue client's keys (`force_touch`,
  `trackpad_mode`, `trackpad_cursor_hidden`).
- Shared state for these (in `src/state/`, marked "the fork"): `clipboard.ts` (the recent texts, in
  memory, gone on logout), `microphone.ts` (one microphone for the bar and the sheet) and `stage.ts`
  (the picture's element, for fullscreen from the bar, with the Vue client's fallbacks: WebKit's
  prefixed call, the iPhone's video player for a viewer, the page filled with the picture for a host); and `src/components/Switch.tsx`, the on/off
  switch of the settings.
- A new version for an open tab (`src/state/update.ts`, started in `src/main.tsx` for built pages):
  the page compares the hashed names of its built files with those of the server's `index.html`
  when the connection drops or returns, when the tab comes back, and when a lazily loaded file of
  the old build is gone; it loads again when that interrupts nothing (no connection, or one back
  within a minute of an interruption such as the deploy's restart) and otherwise offers it in a
  toast with "Reload" (`layout/Toaster`'s action button). New tabs and reloads are served the new
  build by the server's headers (`no-cache` with an ETag for `index.html`).
- Strings: the namespace `fork`, `src/locales/<lang>/fork.json` in every language, used as
  `t('fork:key')`.

Upstream files changed for the fork, each marked with a "the fork" comment: `src/app/App.tsx` (the
shell's classes and the pieces above), `src/index.css` (the import), `src/main.tsx` (the namespace, the version check),
`src/components/ui/button.tsx` (dark ink on mint), `eslint.config.js` (the layer rule),
`src/features/video/Stage.tsx` (no control request without `can_host`; the trackpad's pointer,
the special keys and "force touch" for the keyboard button; the buttons on the picture only in the
embed view; its element for `state/stage.ts`), `src/features/video/Resolution.tsx` (the menu's
side), `src/features/connect/Connect.tsx` (the login card in the fork's look),
`src/features/media/MediaSetting.tsx` (named by the Stream heading), `src/components/Dialog.tsx` and
`src/state/dialogs.ts` (a question with its own confirm word), `src/state/settings.ts` (the touch
settings into the core), `src/state/app.ts` (also the touch settings and the special keys' and the clipboard panel's flags) and
`src/state/neko.ts` (the share link, the media choice, the traces' handle, the clipboard's history), `src/app/App.tsx` also
for the HLS status, the HLS offer without WebRTC and the `?debug` text (`src/state/debug.ts`,
`src/features/media/DebugOverlay.tsx`), `vite.config.ts` (hls.js's own chunk),
`tools/dist.check.mjs` (that it stays lazy), `src/locales/de/common.json` ("Verbunden", "Abmelden"),
`src/plugins/types.ts`, `src/plugins/index.ts` and `src/plugins/filetransfer/` (the `room.lock`
slot and the file-transfer lock in it), `src/plugins/chat/Chat.tsx` (the host's avatar, "(you)",
the time in the viewer's language), `src/plugins/chat/index.ts` (a chat sound that cannot throw),
`src/features/emotes/Emotes.tsx` (its translated label), `src/features/video/useFullscreen.ts`
(WebKit's event and the fallback), and in the
core `client.ts` (`tokenStorage`/`tokenKey`, `logout({ keepalive })`, `setInputMode`, the Vue
client's touch heuristic in `isTouchDevice`),
`overlay.ts` (trackpad mode), `transport/webrtc.ts` (the read-only `peer` and `playback`),
`input/datachannel.ts` (its `stats` and `channelState`, for `?debug`), `index.ts` and
`dist.check.mjs` (the HLS transport; `samplePlayback`).

Imports follow the layer rules (`eslint.config.js`): `src/fork` sits beside `src/layout`, it may use
features, components, state and the plugin registry; only the app shell imports it (the rules forbid
it below).

The visual check (DesignSystem.md, Gate 2) runs against a simulated server:
`node tools/screens.mjs <out-dir>` renders the login, the desktop with and without the side panel, a
narrow window, the phone upright with and without the panel, the phone sideways, and a share link's
login screen and passive view on desktop and phone, a browser without WebRTC, and HLS starting,
failed, in the settings and on a share link (served over HTTPS with a throwaway certificate), and
the trackpad with the special keys on a phone and the touch settings, the control pill hosting and
free, the More menu on desktop and phone, the phone's emotes and clipboard, the settings' end and
the logout question, the file-transfer lock in the settings, the Members tab on desktop and phone
(with a member's menu) and the page-filling fullscreen of a host's iPhone.
