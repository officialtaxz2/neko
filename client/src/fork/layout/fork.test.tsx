// Unit tests of the fork's layout pieces: `npm test` (vitest, react-dom/server). The strings are
// their keys; the room is written into the client's store as system/init would.
import { expect, test, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import type { Session } from '@m1k1o/neko'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {}, Toaster: () => null }))
vi.mock('@/i18n', () => ({ t: (key: string) => key, langs: ['en'], htmlLang: (l: string) => l, setLang() {} }))
vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init() {} },
  useTranslation: () => ({
    t: (key: string, o?: { count?: number }) => key + (o?.count !== undefined ? `:${o.count}` : ''),
    i18n: { language: 'en' },
  }),
}))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { NekoProvider } = await import('@/state/provider')
const { Header } = await import('./Header')
const { MemberStrip } = await import('@/fork/room/MemberStrip')
const { MemberList } = await import('@/fork/room/MemberList')
const { Settings } = await import('@/fork/settings/Settings')
const { initPlugins } = await import('@/plugins')
const { store: filesStore } = await import('@/plugins/filetransfer/store')

const session = (id: string, name: string, admin = false, connected = true): Session => ({
  id,
  profile: {
    name,
    is_admin: admin,
    can_login: true,
    can_connect: true,
    can_watch: true,
    can_host: true,
    can_share_media: false,
    can_access_clipboard: false,
    sends_inactive_cursor: false,
    can_see_inactive_cursors: false,
  },
  state: { is_connected: connected, is_watching: connected },
})
function room({
  admin = false,
  lockedControls = false,
  lockedLogins = false,
  files = false,
  lockedFiles = false,
} = {}) {
  const neko = createNekoApp({ transport: fakeTransport() })
  initPlugins(neko)
  if (files)
    filesStore(neko).setState({
      files: { enabled: true, root_dir: '/', user_download: true, user_upload: true, user_delete: false, files: [] },
    })
  neko.client.store.setState((s) => ({
    session_id: 'me',
    sessions: {
      me: session('me', 'Matze', admin),
      a: session('a', 'Alex', true),
      b: session('b', 'Bo'),
      gone: session('gone', 'Gone', false, false),
    },
    control: { ...s.control, host_id: 'a' },
    settings: {
      ...s.settings,
      locked_controls: lockedControls,
      locked_logins: lockedLogins,
      plugins: { 'filetransfer.enabled': !lockedFiles },
    },
  }))
  return neko
}
// server rendering reads a zustand store's initial state (its server snapshot): render the current one
const render = (neko: ReturnType<typeof room>, el: React.ReactElement) => {
  for (const store of [neko.client.store, neko.app, filesStore(neko)]) store.getInitialState = store.getState
  return renderToStaticMarkup(<NekoProvider neko={neko}>{el}</NekoProvider>)
}

test('header: who is in the room counts the connected members only; no lock and no admin chip by default', () => {
  const html = render(room(), <Header />)
  expect(html).toContain('fork:header.in_room:3')
  expect(html).not.toContain('data-testid="lock-')
  expect(html).not.toContain('data-testid="admin-badge"')
  expect(html).toContain('aria-label="fork:header.open_panel"')
})

test('header: a chip for each lock and for an admin, the toggle names what it does', () => {
  const neko = room({ admin: true, lockedControls: true, lockedLogins: true })
  neko.app.setState({ side: true })
  const html = render(neko, <Header />)
  expect(html).toContain('data-testid="lock-control" data-locked="true"')
  expect(html).toContain('data-testid="lock-login" data-locked="true"')
  expect(html).toContain('data-testid="admin-badge"')
  expect(html).toContain('aria-expanded="true"')
  expect(html).toContain('aria-label="fork:header.close_panel"')
})

test('member strip: you first and not a button, the others as buttons with a menu, the host marked', () => {
  const html = render(room(), <MemberStrip />)
  const members = [...html.matchAll(/<(div|button)[^>]*data-testid="member"[^>]*>/g)].map((m) => m[0])
  expect(members).toHaveLength(3) // the disconnected member is left out
  expect(members[0]).toMatch(/^<div[^>]*data-self="true"/)
  expect(members[1]).toMatch(/^<button[^>]*aria-label="Alex"[^>]*aria-haspopup="menu"[^>]*data-host="true"/)
  expect(members[2]).toMatch(/^<button[^>]*aria-label="Bo"/)
  expect(members[2]).not.toContain('data-host')
})

test('header: file transfer locked is a chip too, the plugin header button is left out', () => {
  expect(render(room({ lockedFiles: true }), <Header />)).not.toContain('lock-file_transfer') // no file transfer
  const html = render(room({ admin: true, files: true, lockedFiles: true }), <Header />)
  expect(html).toContain('data-testid="lock-file_transfer" data-locked="true"')
  expect(html).toContain('fork:chip.file_transfer')
  expect(html).not.toContain('data-testid="lock-files"')
  expect(render(room({ admin: true, files: true }), <Header />)).not.toContain('lock-file_transfer')
})

test('settings: the room section has the file-transfer lock next to control and login, once the server offers it', () => {
  vi.stubGlobal('window', { innerWidth: 1920 })
  vi.stubGlobal('matchMedia', () => ({ matches: false }))
  const switches = (html: string) => [...html.matchAll(/data-testid="lock-switch-(\w+)"/g)].map((m) => m[1])
  expect(switches(render(room({ admin: true }), <Settings />))).toEqual(['control', 'login'])
  const html = render(room({ admin: true, files: true, lockedFiles: true }), <Settings />)
  expect(switches(html)).toEqual(['control', 'login', 'file_transfer'])
  expect(html).toContain('files:locks.locked')
  expect(html).toContain('fork:settings.lock_file_transfer_hint')
  expect(html).toMatch(/<input[^>]*data-testid="lock-switch-file_transfer"[^>]*checked=""/)
  expect(switches(render(room({ files: true }), <Settings />))).toEqual([]) // admins only
})

test('member list: everyone connected with their roles, you first as "(you)", the others open the member menu', () => {
  const neko = room()
  neko.app.setState({ ignored: { b: true } })
  const html = render(neko, <MemberList />)
  expect(html).toContain('fork:room.members (3)')
  const rows = html.split(/(?=<(?:div|button)[^>]*data-testid="member-row")/).slice(1)
  expect(rows).toHaveLength(3)
  expect(rows[0]).toMatch(/^<div[^>]*data-self="true"/)
  expect(rows[0]).toContain('Matze <span class="text-text-muted">(you)</span>')
  expect(rows[1]).toMatch(/^<button[^>]*aria-haspopup="menu"/)
  expect(rows[1]).toContain('Alex')
  expect(rows[1]).toContain('aria-label="fork:room.host"')
  expect(rows[1]).toContain('aria-label="fork:header.admin"')
  expect(rows[2]).toContain('Bo')
  expect(rows[2]).toContain('aria-label="fork:room.ignored"')
  expect(rows[2]).not.toContain('fork:room.host')
})
