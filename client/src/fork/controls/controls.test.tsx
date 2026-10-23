// Unit tests of the fork's controls (fork/controls, fork/tools): what the control pill says and
// offers, the clipboard's recent texts, the More menu's tiles, and a question with its own confirm
// word. `npm test` (vitest, react-dom/server); the strings are their keys.
import { expect, test, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import type { Session } from '@m1k1o/neko'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {}, Toaster: () => null }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init() {} },
  useTranslation: () => ({
    t: (key: string, o?: { name?: string }) => key + (o?.name !== undefined ? `:${o.name}` : ''),
  }),
}))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { NekoProvider } = await import('@/state/provider')
const { ask } = await import('@/state/dialogs')
const { clipboardStore, remember, clearClipboard, CLIPBOARD_HISTORY } = await import('@/state/clipboard')
const { pillState, pillVisible } = await import('./pill')
const { ControlPill } = await import('./ControlPill')
const { MoreTools } = await import('@/fork/tools/MoreTools')

const session = (id: string, name: string, admin = false, canHost = true): Session => ({
  id,
  profile: {
    name,
    is_admin: admin,
    can_login: true,
    can_connect: true,
    can_watch: true,
    can_host: canHost,
    can_share_media: false,
    can_access_clipboard: false,
    sends_inactive_cursor: false,
    can_see_inactive_cursors: false,
  },
  state: { is_connected: true, is_watching: true },
})
interface Room {
  host?: string | null
  admin?: boolean
  canHost?: boolean
  locked?: boolean
  implicit?: boolean
  hls?: boolean
}
function room({
  host = null,
  admin = false,
  canHost = true,
  locked = false,
  implicit = false,
  hls = false,
}: Room = {}) {
  const transport = hls ? { ...fakeTransport(), kind: 'hls' as const } : fakeTransport()
  const neko = createNekoApp({ transport })
  neko.client.store.setState((s) => ({
    session_id: 'me',
    sessions: { me: session('me', 'Matze', admin, canHost), a: session('a', 'Alex') },
    control: { ...s.control, host_id: host },
    settings: { ...s.settings, locked_controls: locked, implicit_hosting: implicit },
    video: { ...s.video, playable: true },
  }))
  return neko
}
const render = (neko: ReturnType<typeof room>, el: React.ReactElement) => {
  for (const store of [neko.client.store, neko.app]) store.getInitialState = store.getState
  return renderToStaticMarkup(<NekoProvider neko={neko}>{el}</NekoProvider>)
}
const state = (r: Room) => pillState(room(r).client.state)

test('control pill state: free, mine, other, locked for members but not admins, implicit hosting', () => {
  expect(state({})).toBe('free')
  expect(state({ host: 'me' })).toBe('mine')
  expect(state({ host: 'a' })).toBe('other')
  expect(state({ host: 'gone' })).toBe('free') // a host who left is nobody
  expect(state({ host: 'a', locked: true })).toBe('locked')
  expect(state({ host: 'a', locked: true, admin: true })).toBe('other')
  expect(state({ implicit: true })).toBe('implicit')
  expect(state({ implicit: true, host: 'me' })).toBe('mine')
  expect(state({ implicit: true, locked: true })).toBe('locked')
  expect(pillVisible(room({ canHost: false }).client.state)).toBe(false)
})

test('control pill: the one action that fits, the input pause for the host only on the desktop', () => {
  expect(render(room(), <ControlPill />)).toContain('data-testid="control-take"')

  const other = render(room({ host: 'a' }), <ControlPill />)
  expect(other).toContain('data-state="other"')
  expect(other).toContain('fork:controls.other_controls:Alex')
  expect(other).toContain('data-testid="control-request"')
  expect(other).not.toContain('control-force')
  expect(render(room({ host: 'a', admin: true }), <ControlPill />)).toContain('data-testid="control-force"')

  const mine = render(room({ host: 'me' }), <ControlPill separator />)
  expect(mine).toContain('data-testid="input-lock"')
  expect(mine).toContain('data-testid="control-release"')
  expect(mine).toMatch(/<\/div><span[^>]*aria-hidden/) // the separator after the pill
  const dock = render(room({ host: 'me' }), <ControlPill compact />)
  expect(dock).not.toContain('input-lock')
  expect(dock).not.toContain('fork:controls.input_active')
  // implicit hosting: nothing to give up
  expect(render(room({ host: 'me', implicit: true }), <ControlPill />)).not.toContain('control-release')
  // a view-only session gets no pill
  expect(render(room({ canHost: false }), <ControlPill />)).toBe('')
  // nor does anyone watching over HLS (receive-only), with implicit hosting or without
  expect(render(room({ hls: true }), <ControlPill />)).toBe('')
  expect(render(room({ hls: true, implicit: true }), <ControlPill />)).toBe('')
})

test('clipboard history: newest first without repeats, at most ten, what the remote copied too, gone on logout', () => {
  const neko = room()
  const history = () => clipboardStore(neko).getState().history
  remember(neko, 'a')
  remember(neko, 'b')
  remember(neko, '  ')
  remember(neko, 'a')
  expect(history()).toEqual(['a', 'b'])
  for (let i = 0; i < 12; i++) remember(neko, `t${i}`)
  expect(history()).toHaveLength(CLIPBOARD_HISTORY)
  expect(history()[0]).toBe('t11')

  neko.client.store.setState((s) => ({ control: { ...s.control, clipboard: { text: 'remote' } } }))
  expect(history()[0]).toBe('remote')
  neko.bus.emit('logout')
  expect(history()).toEqual([])
  remember(neko, 'x')
  clearClipboard(neko)
  expect(history()).toEqual([])
})

test('more menu: the host gets the input tools, the sheet adds what the desktop bar already has', () => {
  // a desktop: a mouse, no touch screen
  vi.stubGlobal('window', { innerWidth: 1920 })
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: !q.includes('coarse') }))
  const tools = (html: string) => [...html.matchAll(/data-testid="tool-([a-z-]+)"/g)].map((m) => m[1])
  const desktop = tools(render(room({ host: 'me', admin: true }), <MoreTools done={() => {}} />))
  expect(desktop).toContain('clipboard')
  expect(desktop).toContain('resolution')
  expect(desktop).toContain('stream')
  expect(desktop).not.toContain('fullscreen')
  expect(desktop).not.toContain('input-pause')
  expect(desktop).not.toContain('keyboard') // no touch screen

  const sheet = tools(render(room({ host: 'me' }), <MoreTools sheet done={() => {}} />))
  expect(sheet).toEqual(expect.arrayContaining(['input-pause', 'clipboard', 'fullscreen', 'stream']))
  expect(sheet).not.toContain('resolution') // admins only
  expect(sheet).not.toContain('microphone') // the profile may not share media

  // over HLS (receive-only): neither the input tools nor the screen size, even as host and admin
  const hls = tools(render(room({ host: 'me', admin: true, hls: true }), <MoreTools sheet done={() => {}} />))
  expect(hls).toEqual(['fullscreen', 'stream'])

  // without control: the picture's tools only
  const watcher = tools(render(room({ host: 'a' }), <MoreTools sheet done={() => {}} />))
  expect(watcher).not.toContain('clipboard')
  expect(watcher).toContain('fullscreen')
  vi.unstubAllGlobals()
})

test('a question with its own confirm word gets the question icon; a plain one stays a warning', async () => {
  const { app } = room()
  const answer = ask(app, 'title', 'text', 'logout')
  expect(app.getState().dialog).toMatchObject({ icon: 'question', confirm: 'logout', cancel: true })
  app.getState().dialog!.resolve(true)
  expect(await answer).toBe(true)

  void ask(app, 'title')
  expect(app.getState().dialog).toMatchObject({ icon: 'warning', confirm: undefined })
})
