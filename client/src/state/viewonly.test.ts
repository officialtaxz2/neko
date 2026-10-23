// Unit tests of view-only share links (viewonly.ts): `npm test` (vitest). A fake server keeps the
// sessions /api/login creates and /api/logout deletes, so a run of page loads can be counted; the
// page (pagehide), sessionStorage, fetch and the websocket are faked, the strings are their keys.
import { afterEach, beforeEach, expect, test, vi } from 'vitest'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn() } }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
const TOKEN = 'Abc-def_1234567X'
let stored = fakeBrowser(`http://neko.test/#/${TOKEN}`)
const { createNekoApp } = await import('./neko.ts')
const { shareToken, startShare, watchShare, SHARE_TOKEN_KEY, rememberedName } = await import('./viewonly.ts')

// the server: sessions by token, the login requests it saw
let sessions: Map<string, { name: string; view: boolean }>
let logins: { username: string; password: string }[]
let requests: { path: string; keepalive?: boolean; auth?: string }[]
let n = 0
const tab = new Map<string, string>() // sessionStorage, which survives a reload of the tab
let page: EventTarget

beforeEach(() => {
  sessions = new Map()
  logins = []
  requests = []
  tab.clear()
  stored = fakeBrowser(`http://neko.test/#/${TOKEN}`)
  page = new EventTarget()
  vi.stubGlobal('window', globalThis) // the client's timers
  vi.stubGlobal('addEventListener', page.addEventListener.bind(page))
  vi.stubGlobal('removeEventListener', page.removeEventListener.bind(page))
  vi.stubGlobal('sessionStorage', {
    getItem: (k: string) => tab.get(k) ?? null,
    setItem: (k: string, v: string) => void tab.set(k, String(v)),
    removeItem: (k: string) => void tab.delete(k),
  })
  vi.stubGlobal(
    'WebSocket',
    class {
      static OPEN = 1
      readyState = 0
      close() {}
      send() {}
    },
  )
  vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
    const path = new URL(url).pathname
    const auth = (init.headers as Record<string, string>)?.Authorization?.replace('Bearer ', '')
    requests.push({ path, keepalive: init.keepalive, auth })
    const json = (status: number, body: unknown) => new Response(JSON.stringify(body), { status })
    if (path === '/api/login') {
      const { username, password } = JSON.parse(init.body as string)
      logins.push({ username, password })
      if (password !== TOKEN && password !== 'member') return json(401, { message: 'invalid password' })
      const token = `t${++n}`
      sessions.set(token, { name: username, view: password === TOKEN })
      return json(200, { id: token, token })
    }
    if (!auth || !sessions.has(auth)) return json(401, { message: 'unauthorized' })
    if (path === '/api/logout') return (sessions.delete(auth), json(200, true))
    return json(200, { id: auth })
  })
})
afterEach(() => vi.unstubAllGlobals())

// one load of the share page: a new instance, started as main.tsx and ShareConnect do
async function load(href = `http://neko.test/#/${TOKEN}`) {
  Object.defineProperty(globalThis, 'location', { value: { href }, configurable: true, writable: true })
  const neko = createNekoApp({ transport: fakeTransport() })
  neko.client.setUrl(href)
  await startShare(neko)
  return neko
}
const leave = (persisted = false) => page.dispatchEvent(Object.assign(new Event('pagehide'), { persisted }))
const settle = () => new Promise((r) => setTimeout(r, 0))

test('the share token: exactly 16 URL-safe characters after #/, nothing else', () => {
  expect(shareToken(`http://neko.test/#/${TOKEN}`)).toBe(TOKEN)
  expect(shareToken(`http://neko.test/room/?usr=x#/${TOKEN}`)).toBe(TOKEN)
  for (const href of [
    'http://neko.test/',
    'http://neko.test/#/short',
    `http://neko.test/#/${TOKEN}x`,
    'http://neko.test/#/Abc+def/1234567X',
    `http://neko.test/?token=${TOKEN}`,
    `http://neko.test/#${TOKEN}`,
  ])
    expect(shareToken(href), href).toBeUndefined()
})

test('a share page logs in with the token in the body, under the remembered name, and only shows the picture', async () => {
  stored.set('view_name', 'Gast')
  const neko = await load()
  expect(neko.share).toBe(TOKEN)
  expect(logins).toEqual([{ username: 'Gast', password: TOKEN }])
  expect(requests.every((r) => !r.path.includes(TOKEN))).toBe(true) // never in a request line
  expect(neko.client.state.authenticated).toBe(true)
  expect([...sessions.values()]).toEqual([{ name: 'Gast', view: true }])
})

test('without a name the login screen asks; the name given is remembered for the next visit', async () => {
  const neko = await load()
  expect(logins).toEqual([])
  expect(neko.client.state.authenticated).toBe(false)
  await watchShare(neko, 'Kim')
  expect(rememberedName()).toBe('Kim')
  expect(logins).toEqual([{ username: 'Kim', password: TOKEN }])
})

test('ten reloads leave one session in the room, whether the page logs out on leaving or is killed', async () => {
  stored.set('view_name', 'Gast')
  for (let i = 0; i < 10; i++) {
    await load()
    expect(sessions.size).toBe(1)
    leave() // reload: the old session logs out as the page goes, the next load logs in anew
    await settle()
  }
  expect(sessions.size).toBe(0)
  expect(requests.filter((r) => r.path === '/api/logout').every((r) => r.keepalive)).toBe(true)

  // leaving drops the token at once: the logout's answer rarely arrives on unload, and the next page
  // must not resume the session being logged out
  await load()
  const left = page.dispatchEvent(Object.assign(new Event('pagehide'), { persisted: false }))
  expect(left && tab.has(SHARE_TOKEN_KEY)).toBe(false)
  await settle()
  expect(sessions.size).toBe(0)

  // a tab the system killed (no pagehide) resumes its session from sessionStorage
  await load()
  const first = tab.get(SHARE_TOKEN_KEY)
  for (let i = 0; i < 10; i++) await load()
  expect(tab.get(SHARE_TOKEN_KEY)).toBe(first)
  expect(sessions.size).toBe(1)
})

test('the back/forward cache keeps the session; a member’s own saved session stays untouched', async () => {
  stored.set('neko_session', 'member-token')
  stored.set('view_name', 'Gast')
  const neko = await load()
  leave(true) // into the bfcache: lifecycle.ts disconnects, the session stays
  await settle()
  expect(sessions.size).toBe(1)
  expect(neko.client.state.authenticated).toBe(true)
  expect(stored.get('neko_session')).toBe('member-token')
  expect(tab.has(SHARE_TOKEN_KEY)).toBe(true)
  expect(requests.some((r) => r.auth === 'member-token')).toBe(false) // a member's session is never resumed
})

test('the automatic start happens once per page: after a kick the viewer decides', async () => {
  stored.set('view_name', 'Gast')
  const neko = await load()
  await neko.client.logout() // as after a kick: the session is gone, the login screen shows
  expect(startShare(neko)).toBeNull()
  expect(logins).toHaveLength(1)
})

test('a refused link says so', async () => {
  Object.defineProperty(globalThis, 'location', {
    value: { href: 'http://neko.test/#/WrongWrongWrong1' },
    configurable: true,
    writable: true,
  })
  const neko = createNekoApp({ transport: fakeTransport() })
  neko.client.setUrl(location.href)
  await watchShare(neko, 'Gast')
  expect(neko.client.state.authenticated).toBe(false)
  expect(neko.app.getState().dialog).toMatchObject({ title: 'connect.error', text: 'fork:share.invalid' })
})
