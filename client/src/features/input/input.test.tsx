// Unit tests of touch input (features/input): the special keys' bar sending keys, the settings
// reaching the core (force touch, the trackpad) and the settings' rows. `npm test` (vitest,
// react-dom/server); the strings are their keys.
import { afterEach, beforeEach, expect, test, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {}, Toaster: () => null }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init() {} },
  useTranslation: () => ({ t: (key: string) => key }),
}))
let stored = fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { NekoProvider } = await import('@/state/provider')
const { setSetting } = await import('@/state/settings')
const { specialKeys, MODIFIERS, XK } = await import('./keys')
const { InputSetting } = await import('./InputSetting')

// a phone: touch, a coarse pointer; or a wide desktop with a mouse
const device = (touch: boolean) => {
  vi.stubGlobal('window', touch ? { ontouchstart: null, innerWidth: 390 } : { innerWidth: 1920 })
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: touch ? q.includes('coarse') : !q.includes('coarse') }))
}
beforeEach(() => {
  stored = fakeBrowser()
  device(false)
  vi.useFakeTimers()
})
afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

const render = (neko: ReturnType<typeof createNekoApp>, el: React.ReactElement) => {
  for (const store of [neko.client.store, neko.app]) store.getInitialState = store.getState
  return renderToStaticMarkup(<NekoProvider neko={neko}>{el}</NekoProvider>)
}

test('special keys: a modifier stays held until tapped again, a key is tapped, a shortcut wraps its key in Ctrl unless Ctrl is held; reset lets go', () => {
  const sent: [number, boolean][] = []
  vi.stubGlobal('window', globalThis)
  const keys = specialKeys({ input: { key: (k: number, d: boolean) => sent.push([k, d]) } } as never)

  keys.toggle('shift')
  expect([...keys.held()]).toEqual(['shift'])
  keys.tap(XK.Tab)
  expect(sent).toEqual([
    [MODIFIERS.shift, true],
    [XK.Tab, true],
  ])
  vi.advanceTimersByTime(50)
  expect(sent.at(-1)).toEqual([XK.Tab, false])
  keys.toggle('shift')
  expect(sent.at(-1)).toEqual([MODIFIERS.shift, false])

  sent.length = 0
  keys.shortcut('copy')
  vi.advanceTimersByTime(200)
  expect(sent).toEqual([
    [XK.Control_L, true],
    [0x63, true],
    [0x63, false],
    [XK.Control_L, false],
  ])

  // with Ctrl held, the shortcut leaves it held
  keys.toggle('ctrl')
  sent.length = 0
  keys.shortcut('paste')
  vi.advanceTimersByTime(200)
  expect(sent).toEqual([
    [0x76, true],
    [0x76, false],
  ])

  keys.toggle('alt')
  sent.length = 0
  keys.reset()
  expect(sent).toEqual([
    [MODIFIERS.ctrl, false],
    [MODIFIERS.alt, false],
  ])
  expect(keys.held().size).toBe(0)
})

test('force touch makes any device a touch device; the trackpad is on only where touch input applies, and survives a reload under the Vue client’s keys', () => {
  const neko = createNekoApp({ transport: fakeTransport() })
  const trackpad = vi.spyOn(neko.overlay, 'setTrackpad')
  expect(neko.client.isTouchDevice).toBe(false)

  setSetting(neko, 'trackpad_mode', true)
  expect(trackpad).toHaveBeenLastCalledWith(false) // a desktop: no trackpad
  setSetting(neko, 'force_touch', true)
  expect(neko.client.isTouchDevice).toBe(true)
  expect(trackpad).toHaveBeenLastCalledWith(true)
  setSetting(neko, 'force_touch', false)
  expect(neko.client.isTouchDevice).toBe(false)
  expect(trackpad).toHaveBeenLastCalledWith(false)
  expect([stored.get('trackpad_mode'), stored.get('force_touch')]).toEqual(['1', '0'])

  device(true) // a phone, loaded with the stored setting
  const phone = createNekoApp({ transport: fakeTransport() })
  expect(phone.client.isTouchDevice).toBe(true)
  expect(phone.app.getState().settings.trackpad_mode).toBe(true)
  expect(phone.overlay.trackpadCursor).toEqual({ x: 0.5, y: 0.5, touching: false })
  expect((phone.overlay as unknown as { trackpad: boolean }).trackpad).toBe(true)
})

test('touch devices as the Vue client counted them: an iPad with keyboard and trackpad, a window up to 1024px', () => {
  const neko = createNekoApp({ transport: fakeTransport() })
  expect(neko.client.isTouchDevice).toBe(false) // a wide desktop with a mouse
  vi.stubGlobal('navigator', {
    ...navigator,
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
    maxTouchPoints: 5,
  })
  expect(neko.client.isTouchDevice).toBe(true) // iPadOS calls itself a Mac, but has touch points
  vi.unstubAllGlobals()
  device(false)
  vi.stubGlobal('window', { innerWidth: 900 })
  expect(neko.client.isTouchDevice).toBe(true) // a narrow window
})

test('settings: force touch always, the trackpad on a touch device, its pointer while the trackpad is on', () => {
  const rows = (neko: ReturnType<typeof createNekoApp>) =>
    [...render(neko, <InputSetting row="" label="" />).matchAll(/data-testid="setting-(\w+)"/g)].map((m) => m[1])
  const desktop = createNekoApp({ transport: fakeTransport() })
  expect(rows(desktop)).toEqual(['force_touch'])
  setSetting(desktop, 'force_touch', true)
  expect(rows(desktop)).toEqual(['force_touch', 'trackpad_mode'])
  setSetting(desktop, 'trackpad_mode', true)
  expect(rows(desktop)).toEqual(['force_touch', 'trackpad_mode', 'trackpad_cursor_hidden'])

  device(true)
  stored = fakeBrowser()
  expect(rows(createNekoApp({ transport: fakeTransport() }))).toEqual(['force_touch', 'trackpad_mode'])
})
