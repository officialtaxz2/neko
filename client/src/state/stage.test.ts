// Unit tests of the picture's fullscreen (stage.ts): `npm test` (vitest). Each browser is an element
// with the fullscreen calls it has; the iPhone has none, only its video player.
import { expect, test, vi } from 'vitest'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {}, Toaster: () => null }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser()
const { createNekoApp } = await import('./neko')
const { stageStore, requestFullscreen } = await import('./stage')

function stage(player: object, { host = false, video = {} as object } = {}) {
  const transport = { ...fakeTransport(), element: video as HTMLVideoElement }
  const neko = createNekoApp({ transport })
  neko.client.store.setState((s) => ({ session_id: 'me', control: { ...s.control, host_id: host ? 'me' : null } }))
  stageStore(neko).setState({ player: player as HTMLElement })
  return { neko, fallback: () => stageStore(neko).getState().fallback }
}

test('the standard call, else the prefixed one: no fallback', async () => {
  const standard = { requestFullscreen: vi.fn(async () => {}) }
  const s1 = stage(standard)
  await requestFullscreen(s1.neko)
  expect(standard.requestFullscreen).toHaveBeenCalledOnce()
  expect(s1.fallback()).toBe(false)

  const webkit = { webkitRequestFullscreen: vi.fn() }
  const s2 = stage(webkit)
  await requestFullscreen(s2.neko)
  expect(webkit.webkitRequestFullscreen).toHaveBeenCalledOnce()
  expect(s2.fallback()).toBe(false)
})

test("an iPhone: a viewer gets the video's own player, the host the page-filling fallback, which the next call ends", async () => {
  const video = { webkitEnterFullscreen: vi.fn() }
  const watching = stage({}, { video })
  await requestFullscreen(watching.neko)
  expect(video.webkitEnterFullscreen).toHaveBeenCalledOnce()
  expect(watching.fallback()).toBe(false)

  const host = stage({}, { host: true, video })
  await requestFullscreen(host.neko)
  expect(video.webkitEnterFullscreen).toHaveBeenCalledOnce() // not again: its player would take the input
  expect(host.fallback()).toBe(true)
  await requestFullscreen(host.neko)
  expect(host.fallback()).toBe(false)
})

test('a browser that refuses the element fullscreen falls back to filling the page', async () => {
  const s = stage({ requestFullscreen: vi.fn(async () => Promise.reject(new Error('denied'))) })
  await requestFullscreen(s.neko)
  expect(s.fallback()).toBe(true)
})
