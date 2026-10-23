// Unit tests of the fork's layout pieces: `npm test` (vitest, react-dom/server). The strings are
// their keys; the room is written into the client's store as system/init would.
import { expect, test, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import type { Session } from '@m1k1o/neko'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {}, Toaster: () => null }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init() {} },
  useTranslation: () => ({
    t: (key: string, o?: { count?: number }) => key + (o?.count !== undefined ? `:${o.count}` : ''),
  }),
}))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { NekoProvider } = await import('@/state/provider')
const { Header } = await import('./Header')
const { MemberStrip } = await import('@/fork/room/MemberStrip')

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
function room({ admin = false, lockedControls = false, lockedLogins = false } = {}) {
  const neko = createNekoApp({ transport: fakeTransport() })
  neko.client.store.setState((s) => ({
    session_id: 'me',
    sessions: {
      me: session('me', 'Matze', admin),
      a: session('a', 'Alex', true),
      b: session('b', 'Bo'),
      gone: session('gone', 'Gone', false, false),
    },
    control: { ...s.control, host_id: 'a' },
    settings: { ...s.settings, locked_controls: lockedControls, locked_logins: lockedLogins },
  }))
  return neko
}
// server rendering reads a zustand store's initial state (its server snapshot): render the current one
const render = (neko: ReturnType<typeof room>, el: React.ReactElement) => {
  for (const store of [neko.client.store, neko.app]) store.getInitialState = store.getState
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
