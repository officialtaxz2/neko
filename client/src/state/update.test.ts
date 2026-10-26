// A new version reaching an open tab: the built files named by the page and by the server's
// index.html, which answers count, and what a new version does in which situation (load it or
// offer it), with the reload guard against a stale copy.
import { beforeEach, test, vi } from 'vitest'
import assert from 'node:assert/strict'
import { fakeBrowser, fakeTransport } from '@/test/browser'

const info = vi.fn()
vi.mock('sonner', () => ({ toast: { info: (...args: unknown[]) => info(...args) } }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser()
const { NekoClient } = await import('@m1k1o/neko')
const { builtFiles, htmlURLs, servedFiles, onNewVersion, bindUpdates, RECENT_MS, RELOAD_KEY, RETRY_MS, VISIBLE_MS } =
  await import('./update.ts')
type UpdateEnv = Parameters<typeof bindUpdates>[2] & object
type NekoApp = Parameters<typeof bindUpdates>[0]

const BASE = 'https://neko.test/'
const index = (hash: string, css = 'CO6vvtZL') => `<!doctype html>
<html lang="en">
  <head>
    <link rel="icon" type="image/png" sizes="32x32" href="favicon-32x32.png" />
    <link rel="manifest" href="site.webmanifest" />
    <script type="module" crossorigin src="./assets/index-${hash}.js"></script>
    <link rel="modulepreload" crossorigin href="./assets/react-D5ydvgTr.js">
    <link rel="stylesheet" crossorigin href="./assets/index-${css}.css">
  </head>
  <body><div id="root"></div></body>
</html>`
const OLD = builtFiles(htmlURLs(index('Dw4TracB')), BASE)
const html = (body: string, init: ResponseInit = {}) =>
  new Response(body, { status: 200, headers: { 'content-type': 'text/html; charset=utf-8' }, ...init })

test('the built files: the paths under assets/, sorted, resolved against the page (a path prefix too)', () => {
  assert.deepEqual(OLD, ['/assets/index-CO6vvtZL.css', '/assets/index-Dw4TracB.js', '/assets/react-D5ydvgTr.js'])
  assert.deepEqual(builtFiles(htmlURLs(index('Dw4TracB')), 'https://neko.test/room/'), [
    '/room/assets/index-CO6vvtZL.css',
    '/room/assets/index-Dw4TracB.js',
    '/room/assets/react-D5ydvgTr.js',
  ])
  // another origin's assets/ (a browser extension built with Vite, a CDN) are not this build
  assert.deepEqual(
    builtFiles(
      ['chrome-extension://abcdef/assets/content-x1.js', 'https://cdn.test/assets/a.js', './assets/b.js'],
      BASE,
    ),
    ['/assets/b.js'],
  )
  // a lazy import's preload link is absolute: the same file
  assert.deepEqual(builtFiles(['https://neko.test/assets/react-D5ydvgTr.js', './assets/react-D5ydvgTr.js'], BASE), [
    '/assets/react-D5ydvgTr.js',
  ])
})

test("the server's answer counts only as this client's index.html", async () => {
  const fetcher = (res: Response) => (async () => res) as typeof fetch
  assert.deepEqual(await servedFiles(BASE, fetcher(html(index('NEW')))), [
    '/assets/index-CO6vvtZL.css',
    '/assets/index-NEW.js',
    '/assets/react-D5ydvgTr.js',
  ])
  assert.equal(await servedFiles(BASE, fetcher(html(index('NEW'), { status: 502 }))), undefined)
  assert.equal(await servedFiles(BASE, fetcher(html('<html>login</html>'))), undefined) // e.g. an OAuth login page
  const json = new Response(index('NEW'), { headers: { 'content-type': 'application/json' } })
  assert.equal(await servedFiles(BASE, fetcher(json)), undefined)
  const redirected = html(index('NEW'))
  Object.defineProperty(redirected, 'redirected', { value: true })
  assert.equal(await servedFiles(BASE, fetcher(redirected)), undefined)
})

test('a new version: reloads unless that interrupts someone', () => {
  const at = (m: Partial<Parameters<typeof onNewVersion>[0]>) =>
    onNewVersion({ status: 'connected', interruptedAt: 0, now: 10 * RECENT_MS, typing: false, loadFailed: false, ...m })
  assert.equal(at({ status: 'disconnected' }), 'reload') // the login, a session gone with the restart
  assert.equal(at({ status: 'connecting' }), 'reload')
  assert.equal(at({ interruptedAt: 10 * RECENT_MS - 5_000 }), 'reload') // back by itself after the restart
  assert.equal(at({}), 'offer') // connected for a while: someone may be in control
  assert.equal(at({ interruptedAt: 8 * RECENT_MS }), 'offer')
  assert.equal(at({ loadFailed: true }), 'reload') // what needed the old file is broken anyway
  assert.equal(at({ status: 'disconnected', typing: true }), 'offer') // someone is writing their name
  assert.equal(at({ loadFailed: true, typing: true }), 'offer')
})

// an instance with a page around it: the server's index.html, the time, the listeners, the storage
function setup({ storage = true }: { storage?: boolean } = {}) {
  const client = new NekoClient({ transport: fakeTransport() })
  const calls: string[] = []
  client.disconnect = () => void calls.push('disconnect')
  const listeners: Record<string, (e: unknown) => void> = {}
  const target = { addEventListener: (ev: string, fn: (e: unknown) => void) => (listeners[ev] = fn) }
  const stored = new Map<string, string>()
  const timers = new Map<number, () => void>()
  let nextTimer = 1
  const page = {
    served: index('Dw4TracB') as string | null, // null: the server is not there
    fetches: 0,
    now: 100 * RECENT_MS,
    activeElement: null as Element | null,
    visibilityState: 'visible' as DocumentVisibilityState,
  }
  const env: UpdateEnv = {
    url: BASE,
    fetch: (async () => {
      page.fetches++
      if (page.served === null) throw new TypeError('Failed to fetch')
      return html(page.served)
    }) as typeof fetch,
    reload: () => void calls.push('reload'),
    storage: storage
      ? { getItem: (k: string) => stored.get(k) ?? null, setItem: (k: string, v: string) => void stored.set(k, v) }
      : undefined,
    window: target,
    document: {
      ...target,
      get visibilityState() {
        return page.visibilityState
      },
      get activeElement() {
        return page.activeElement
      },
    },
    now: () => page.now,
    setInterval: (fn) => (timers.set(nextTimer, fn), nextTimer++),
    clearInterval: (id) => void timers.delete(id),
  }
  bindUpdates({ client } as unknown as NekoApp, OLD, env)
  const status = (s: 'connected' | 'connecting' | 'disconnected') => {
    client.store.setState({ connection: { ...client.state.connection, status: s } })
    client.events.emit('connection.status', s)
  }
  const settle = () => new Promise((resolve) => setTimeout(resolve, 0))
  return { client, calls, listeners, stored, timers, page, status, settle }
}

beforeEach(() => info.mockClear())

test('the same build: nothing happens', async () => {
  const s = setup()
  s.status('connecting')
  s.status('connected')
  await s.settle()
  assert.deepEqual(s.calls, [])
  assert.equal(info.mock.calls.length, 0)
})

test('a deploy while connected: the restart interrupts, the new build loads once the server is back', async () => {
  const s = setup()
  s.status('connected')
  await s.settle()
  s.page.served = null // the container is being replaced
  s.status('connecting')
  await s.settle()
  assert.deepEqual(s.calls, [], 'no server, no decision')
  assert.equal(s.timers.size, 1, 'asked again while reconnecting')
  s.page.served = index('NEW')
  s.page.now += RETRY_MS
  s.timers.values().next().value!()
  await s.settle()
  assert.deepEqual(s.calls, ['disconnect', 'reload'])
  assert.equal(s.stored.get(RELOAD_KEY), '/assets/index-CO6vvtZL.css /assets/index-NEW.js /assets/react-D5ydvgTr.js')
})

test('back by itself within a minute (a share link, a kept session): reloads; later only offers', async () => {
  const s = setup()
  s.status('connecting')
  s.page.now += 10_000
  s.page.served = index('NEW')
  s.status('connected')
  await s.settle()
  assert.equal(s.timers.size, 0, 'no more asking once connected')
  assert.deepEqual(s.calls, ['disconnect', 'reload'])

  const later = setup()
  later.status('connecting')
  later.status('connected')
  await later.settle()
  later.page.now += 2 * RECENT_MS
  later.page.served = index('NEW')
  later.listeners.pageshow({ persisted: true }) // back from the back/forward cache
  await later.settle()
  assert.deepEqual(later.calls, [])
  assert.equal(info.mock.calls.length, 1)
  const [title, options] = info.mock.calls[0] as [
    string,
    { id: string; action: { label: string; onClick: () => void } },
  ]
  assert.equal(title, 'fork:update.title')
  assert.equal(options.id, 'update')
  assert.equal(options.action.label, 'fork:update.reload')
  options.action.onClick()
  assert.deepEqual(later.calls, ['disconnect', 'reload'])
})

test('a stale copy after the reload, or no storage for the guard: only offers', async () => {
  const s = setup()
  s.stored.set(RELOAD_KEY, '/assets/index-CO6vvtZL.css /assets/index-NEW.js /assets/react-D5ydvgTr.js')
  s.page.served = index('NEW')
  s.status('disconnected')
  await s.settle()
  assert.deepEqual(s.calls, [])
  assert.equal(info.mock.calls.length, 1)

  const bare = setup({ storage: false })
  bare.page.served = index('NEW')
  bare.status('disconnected')
  await bare.settle()
  assert.deepEqual(bare.calls, [])
  assert.equal(info.mock.calls.length, 2)
})

test('the login: an empty focused name reloads, a name being written is left alone', async () => {
  const s = setup()
  s.page.served = index('NEW')
  s.page.activeElement = { tagName: 'INPUT', type: 'text', value: 'mat' } as unknown as Element
  s.status('disconnected')
  await s.settle()
  assert.deepEqual(s.calls, [])
  assert.equal(info.mock.calls.length, 1)
  s.page.activeElement = { tagName: 'INPUT', type: 'text', value: '' } as unknown as Element
  s.status('disconnected')
  await s.settle()
  assert.deepEqual(s.calls, ['disconnect', 'reload'])
})

test('a lazy file of the old build is gone: reloads even while connected', async () => {
  const s = setup()
  s.status('connected')
  s.page.now += 5 * RECENT_MS
  s.page.served = index('NEW')
  s.listeners['vite:preloadError']({})
  await s.settle()
  assert.deepEqual(s.calls, ['disconnect', 'reload'])
})

test('coming back to the tab asks at most once a minute', async () => {
  const s = setup()
  s.status('connected')
  s.page.now += 5 * RECENT_MS // connected for a while: a new version is offered
  await s.settle()
  const before = s.page.fetches
  s.page.served = index('NEW')
  s.listeners.visibilitychange({})
  await s.settle()
  assert.equal(s.page.fetches, before + 1)
  assert.equal(info.mock.calls.length, 1)
  s.listeners.visibilitychange({}) // within the minute
  s.page.visibilityState = 'hidden'
  s.listeners.visibilitychange({}) // going away
  await s.settle()
  assert.equal(s.page.fetches, before + 1)
  s.page.now += VISIBLE_MS
  s.page.visibilityState = 'visible'
  s.listeners.visibilitychange({})
  await s.settle()
  assert.equal(s.page.fetches, before + 2)
  assert.deepEqual(s.calls, [], 'offered, not reloaded')
})

test('a page without built files (not a build) binds nothing', async () => {
  const client = new NekoClient({ transport: fakeTransport() })
  let asked = 0
  bindUpdates({ client } as unknown as NekoApp, [], {
    fetch: (async () => (asked++, html(index('NEW')))) as typeof fetch,
  } as unknown as UpdateEnv)
  client.events.emit('connection.status', 'disconnected')
  await new Promise((resolve) => setTimeout(resolve, 0))
  assert.equal(asked, 0)
})
