// the fork: what the special keys' bar (KeyboardHelper) sends. Modifiers stay held until tapped
// again, the other keys are tapped (pressed, released 50 ms later), a shortcut wraps its key in
// Ctrl unless Ctrl is already held.
import type { NekoClient } from '@m1k1o/neko'

// X11 keysyms
export const XK = {
  Escape: 0xff1b,
  Tab: 0xff09,
  Left: 0xff51,
  Up: 0xff52,
  Right: 0xff53,
  Down: 0xff54,
  Control_L: 0xffe3,
}
export const MODIFIERS = { ctrl: 0xffe3, alt: 0xffe9, shift: 0xffe1, super: 0xffeb }
export type Modifier = keyof typeof MODIFIERS
export const SHORTCUTS = { all: 0x0061, copy: 0x0063, paste: 0x0076, undo: 0x007a }
export type Shortcut = keyof typeof SHORTCUTS
const STEP_MS = 50

export function specialKeys(client: Pick<NekoClient, 'input'>) {
  const held = new Set<Modifier>()
  const key = (keysym: number, down: boolean) => client.input.key(keysym, down)
  const later = (fn: () => void) => window.setTimeout(fn, STEP_MS)
  return {
    held: () => new Set(held),
    toggle(m: Modifier) {
      const down = !held.has(m)
      if (down) held.add(m)
      else held.delete(m)
      key(MODIFIERS[m], down)
    },
    tap(keysym: number) {
      key(keysym, true)
      later(() => key(keysym, false))
    },
    shortcut(s: Shortcut) {
      const ctrl = !held.has('ctrl')
      if (ctrl) key(XK.Control_L, true)
      later(() => {
        key(SHORTCUTS[s], true)
        later(() => {
          key(SHORTCUTS[s], false)
          if (ctrl) later(() => key(XK.Control_L, false))
        })
      })
    },
    // lets go of every held modifier
    reset() {
      for (const m of held) key(MODIFIERS[m], false)
      held.clear()
    },
  }
}
export type SpecialKeys = ReturnType<typeof specialKeys>
