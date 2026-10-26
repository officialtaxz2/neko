// The fork's ?debug text (debug.ts): when it is on, what it reads, and that it never shows a secret
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({
  toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn(), dismiss: vi.fn() },
}))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser('http://neko.test/?debug=1')
Object.defineProperty(globalThis, 'window', { value: globalThis, configurable: true, writable: true })
Object.defineProperty(globalThis, 'matchMedia', { value: () => ({ matches: true }), configurable: true })
const { createNekoApp } = await import('./neko.ts')
const debug = await import('./debug.ts')

// a Storage over a Map, with key() and length (the test browser's localStorage has neither)
const storage = (entries: Record<string, string>): Storage => {
  const map = new Map(Object.entries(entries))
  return {
    get length() {
      return map.size
    },
    key: (i: number) => [...map.keys()][i] ?? null,
    getItem: (k: string) => map.get(k) ?? null,
    setItem: (k: string, v: string) => void map.set(k, v),
    removeItem: (k: string) => void map.delete(k),
    clear: () => map.clear(),
  }
}

const counters = () => ({ touches: 0, lastTouch: '', streamAt: 0, firstFrameAt: 0 })

test('?debug turns it on, any value; without it nothing', () => {
  assert.equal(debug.debugEnabled('?debug'), true)
  assert.equal(debug.debugEnabled('?usr=a&debug=1'), true)
  assert.equal(debug.debugEnabled('?usr=a'), false)
  assert.equal(debug.debugEnabled(''), false)
})

test('the stored settings leave out tokens, sessions and passwords and cut long values', () => {
  const text = debug.storedSettings(
    storage({
      neko_session: 'secret-token',
      neko_view_session: 'secret-view',
      password: 'secret-pw',
      access_token: 'secret-at',
      trackpad_mode: '1',
      media_backend: 'll-hls',
      long: 'x'.repeat(40),
    }),
  )
  assert.doesNotMatch(text, /secret/)
  assert.equal(text, `long=${'x'.repeat(24)} media_backend=ll-hls trackpad_mode=1`)
  assert.equal(debug.storedSettings(storage({})), 'empty')
})

test('only the last three page errors are kept, each cut short', () => {
  for (const e of ['one', 'two', 'three', 'x'.repeat(300)]) debug.noteError(e)
  assert.deepEqual(debug.pageErrors().slice(0, 2), ['two', 'three'])
  assert.equal(debug.pageErrors()[2].length, 160)
})

test('the text names the path, the connection, the video and the input, never the session token', () => {
  localStorage.setItem('neko_session', 'secret-token')
  const neko = createNekoApp({ transport: fakeTransport() })
  neko.client.store.setState({
    connection: { ...neko.client.state.connection, status: 'connected', token: 'secret-token' },
  })
  const text = debug.describe(neko, { ...counters(), touches: 4, lastTouch: 'touchend' })
  assert.match(text, /^media webrtc connection connected$/m)
  assert.match(text, /^video none$/m)
  assert.match(text, /^touches 4 touchend sent - skipped -$/m)
  assert.match(text, /^input hosting false locked false/m)
  assert.doesNotMatch(text, /secret/)
})

test('the time from the stream to its first frame', () => {
  const neko = createNekoApp({ transport: fakeTransport() })
  const text = debug.describe(neko, { ...counters(), streamAt: 1_000, firstFrameAt: 2_500 }, 4_000)
  assert.match(text, /^stream since 3\.0s first frame after 1\.5s$/m)
  assert.match(debug.describe(neko, counters()), /^stream since - first frame after -$/m)
})

test('the quality tier the server selected, once it has said so', () => {
  const neko = createNekoApp({ transport: fakeTransport() })
  assert.doesNotMatch(debug.describe(neko, counters()), /^quality/m)
  const text = debug.describe(neko, { ...counters(), video: { id: 'medium', auto: true } })
  assert.match(text, /^quality medium auto$/m)
})
