// The back/forward cache: a page cached while connected closes its connection and connects the same
// session again when it comes back; a page that was not connected, or is logged out, stays so.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({
  toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn(), dismiss: vi.fn(), getToasts: () => [] },
}))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser()
const { NekoClient } = await import('@m1k1o/neko')
const { bindPageLifecycle } = await import('./lifecycle.ts')

const page = () => {
  const listeners: Record<string, (e: unknown) => void> = {}
  return {
    addEventListener: (ev: string, fn: (e: unknown) => void) => (listeners[ev] = fn),
    removeEventListener: (ev: string) => delete listeners[ev],
    fire: (ev: string, persisted: boolean) => listeners[ev]?.({ persisted }),
    listeners,
  }
}
const client = (status: 'connected' | 'disconnected', authenticated = true) => {
  const c = new NekoClient({ transport: fakeTransport() })
  const calls: string[] = []
  c.store.setState({ authenticated, connection: { ...c.state.connection, status } })
  c.disconnect = () => {
    calls.push('disconnect')
    c.store.setState({ connection: { ...c.state.connection, status: 'disconnected' } })
  }
  c.connect = () => void calls.push('connect')
  return { c, calls }
}

test('cached while connected: disconnect on pagehide, connect again on pageshow', () => {
  const { c, calls } = client('connected')
  const p = page()
  const unbind = bindPageLifecycle(c, p)
  p.fire('pagehide', false) // an ordinary unload: the browser closes everything itself
  assert.deepEqual(calls, [])
  p.fire('pagehide', true)
  assert.deepEqual(calls, ['disconnect'])
  p.fire('pageshow', true)
  assert.deepEqual(calls, ['disconnect', 'connect'])
  p.fire('pageshow', true) // nothing left to resume
  assert.deepEqual(calls, ['disconnect', 'connect'])
  unbind()
  assert.deepEqual(Object.keys(p.listeners), [])
})

test('not connected, or logged out meanwhile: nothing is resumed', () => {
  const idle = client('disconnected')
  const p = page()
  bindPageLifecycle(idle.c, p)
  p.fire('pagehide', true)
  p.fire('pageshow', true)
  assert.deepEqual(idle.calls, [])

  const out = client('connected')
  const q = page()
  bindPageLifecycle(out.c, q)
  q.fire('pagehide', true)
  out.c.store.setState({ authenticated: false })
  q.fire('pageshow', true)
  assert.deepEqual(out.calls, ['disconnect'])
})
