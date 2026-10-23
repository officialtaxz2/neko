// The stream's path (media.ts): the stored choice and ?media=, the transport an instance gets, what
// the server offers, and a new choice leading to a fresh page load
import { afterEach, beforeEach, expect, test, vi } from 'vitest'
import { fakeBrowser } from '@/test/browser'

vi.mock('sonner', () => ({ toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn() } }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
let stored = fakeBrowser()
const { HLSTransport, WebRTCTransport, WebSocketInput } = await import('@m1k1o/neko')
const { createNekoApp } = await import('./neko.ts')
const media = await import('./media.ts')

const capabilities = {
  version: 1,
  backend: 'hls',
  modes: ['hls'],
  container: 'fmp4',
  video_codec: 'h264-high-3.1',
  audio_codec: 'aac-lc',
  audio_rate: 48000,
  variants: (
    [
      ['high', 1280, 720, 25],
      ['medium', 854, 480, 20],
      ['low', 640, 360, 15],
    ] as const
  ).map(([id, width, height, frame_rate]) => ({
    id,
    width,
    height,
    frame_rate,
    video_codec: 'avc1.64001f',
    container: 'video/mp4',
    bandwidth: 1,
  })),
  limits: {
    idle_expires_in_ms: 30000,
    maximum_playlist_bytes: 65536,
    max_requests_per_lease: 8,
    max_blocking_per_lease: 4,
  },
}

let nav: { assign: ReturnType<typeof vi.fn>; reload: ReturnType<typeof vi.fn> }
function page(href: string) {
  stored = fakeBrowser(href)
  nav = { assign: vi.fn(), reload: vi.fn() }
  Object.assign(location, nav)
}
beforeEach(() => {
  page('https://neko.test/')
  vi.stubGlobal('window', globalThis)
  vi.stubGlobal('addEventListener', () => {})
  vi.stubGlobal('removeEventListener', () => {})
})
afterEach(() => vi.unstubAllGlobals())

test('the stored choice: absent, invalid or the removed webcodecs-ws is WebRTC; only exact values are stored', () => {
  expect(media.readMedia()).toBe('webrtc')
  for (const v of ['', 'HLS', 'webcodecs-ws', 'auto']) {
    stored.set('media_backend', v)
    expect(media.readMedia(), v).toBe('webrtc')
  }
  stored.set('media_backend', 'll-hls')
  expect(media.readMedia()).toBe('ll-hls')
  expect(
    media.readMedia({
      getItem: () => {
        throw new Error('denied')
      },
    }),
  ).toBe('webrtc')
  expect(media.writeMedia('nonsense')).toBe(true)
  expect(stored.get('media_backend')).toBe('webrtc')
  expect(media.writeMedia('hls')).toBe(true)
  expect(stored.get('media_backend')).toBe('hls')
  expect(
    media.writeMedia('hls', {
      setItem: () => {
        throw new Error('denied')
      },
    }),
  ).toBe(false)
})

test('?media=: an exact HLS mode wins for this visit without touching the stored choice; anything else is WebRTC', () => {
  expect(media.resolveMedia('', 'hls')).toEqual({ backend: 'hls', overridden: false, invalid: false })
  expect(media.resolveMedia('?media=ll-hls', 'webrtc')).toEqual({ backend: 'll-hls', overridden: true, invalid: false })
  expect(media.resolveMedia('?usr=x&media=hls', 'll-hls')).toEqual({ backend: 'hls', overridden: true, invalid: false })
  for (const search of ['?media=webrtc', '?media=HLS', '?media=hls&media=hls', '?media='])
    expect(media.resolveMedia(search, 'hls'), search).toEqual({ backend: 'webrtc', overridden: true, invalid: true })
})

test('a new choice re-enters the page without ?media=, a share link’s fragment and other parameters kept', () => {
  expect(media.mediaNavigationURL('https://neko.test/room/?usr=Kim&media=hls#/ShareToken-12345')).toBe(
    'https://neko.test/room/?usr=Kim#/ShareToken-12345',
  )
  const loc = { href: 'https://neko.test/?media=hls', assign: vi.fn(), reload: vi.fn() }
  media.navigateForMedia(loc)
  expect(loc.assign).toHaveBeenCalledWith('https://neko.test/')
  // the same address (a share link's fragment would make assign a mere fragment navigation): reload
  const same = { href: 'https://neko.test/#/ShareToken-12345', assign: vi.fn(), reload: vi.fn() }
  media.navigateForMedia(same)
  expect(same.reload).toHaveBeenCalled()
  expect(same.assign).not.toHaveBeenCalled()
})

test('the instance’s transport: WebRTC by default, HLS as stored or for one visit; input then goes over the socket', () => {
  expect(createNekoApp().client.transport).toBeInstanceOf(WebRTCTransport)

  page('https://neko.test/?media=hls')
  const visit = createNekoApp().client
  expect(visit.transport).toBeInstanceOf(HLSTransport)
  expect((visit.transport as InstanceType<typeof HLSTransport>).mode).toBe('hls')
  expect(visit.input).toBeInstanceOf(WebSocketInput)
  expect(stored.has('media_backend')).toBe(false)

  page('https://neko.test/')
  stored.set('media_backend', 'll-hls')
  const neko = createNekoApp()
  expect((neko.client.transport as InstanceType<typeof HLSTransport>).mode).toBe('ll-hls')
  expect(media.mediaStore(neko).getState().selection).toEqual({ backend: 'll-hls', overridden: false, invalid: false })
})

test('what the server offers: asked for once connected over HTTPS, forgotten when the connection ends', () => {
  const neko = createNekoApp()
  const sent: unknown[] = []
  neko.client.send = (event: string, payload?: unknown) => void sent.push([event, payload])
  neko.client.events.emit('connection.status', 'connected')
  expect(sent).toEqual([
    ['media/hls/capabilities/request', { version: 1, mode: 'hls' }],
    ['media/hls/capabilities/request', { version: 1, mode: 'll-hls' }],
  ])
  neko.client.events.emit('message', 'media/hls/capabilities', capabilities)
  expect(media.mediaStore(neko).getState().modes).toEqual(['hls'])
  neko.client.events.emit('message', 'media/hls/capabilities', { ...capabilities, audio_rate: 44100 })
  expect(media.mediaStore(neko).getState().modes).toEqual([])
  neko.client.events.emit('message', 'media/hls/capabilities', capabilities)
  neko.client.events.emit('connection.status', 'disconnected')
  expect(media.mediaStore(neko).getState().modes).toEqual([])

  // plain HTTP: no lease could work, nothing is asked
  page('http://neko.test/')
  const plain = createNekoApp()
  const none: unknown[] = []
  plain.client.send = (event: string) => void none.push(event)
  plain.client.events.emit('connection.status', 'connected')
  expect(none).toEqual([])
})

test('choosing: an HLS mode the server does not offer cannot be picked; a pick is stored, then the page starts over', () => {
  const neko = createNekoApp()
  const disconnect = vi.spyOn(neko.client, 'disconnect')
  media.chooseMedia(neko, 'hls')
  expect(stored.has('media_backend')).toBe(false)
  expect(nav.assign).not.toHaveBeenCalled()
  expect(nav.reload).not.toHaveBeenCalled()

  neko.client.events.emit('message', 'media/hls/capabilities', capabilities)
  media.chooseMedia(neko, 'hls')
  expect(stored.get('media_backend')).toBe('hls')
  expect(disconnect).toHaveBeenCalled()
  expect(nav.reload).toHaveBeenCalledTimes(1) // https://neko.test/ has no ?media= to drop

  // picking what already runs changes nothing
  media.chooseMedia(neko, 'webrtc')
  expect(nav.reload).toHaveBeenCalledTimes(1)
  expect(stored.get('media_backend')).toBe('webrtc')

  // over ?media=hls, WebRTC (also the failure card's "Use WebRTC") leaves the override behind
  page('https://neko.test/?media=hls')
  const visit = createNekoApp()
  media.chooseMedia(visit, 'webrtc')
  expect(stored.get('media_backend')).toBe('webrtc')
  expect(nav.assign).toHaveBeenCalledWith('https://neko.test/')
  page('https://neko.test/?media=hls')
  const again = createNekoApp()
  media.clearMediaOverride(again)
  expect(nav.assign).toHaveBeenCalledWith('https://neko.test/')
})

test('the HLS transport’s state reaches the store; retry starts it over', async () => {
  page('http://neko.test/?media=hls')
  const neko = createNekoApp()
  const transport = neko.client.transport as InstanceType<typeof HLSTransport>
  expect(media.mediaStore(neko).getState().hls).toMatchObject({ mode: 'hls', status: 'idle' })
  const sent: string[] = []
  const session = {
    url: 'http://neko.test',
    init: { session_id: 's1', sessions: {}, settings: {} } as never,
    send: (event: string) => void sent.push(event),
    on: () => () => {},
  }
  await transport.connect(session)
  expect(media.mediaStore(neko).getState().hls).toMatchObject({
    status: 'terminal',
    detail: 'HLS requires same-origin HTTPS',
  })
  const retry = vi.spyOn(transport, 'retry')
  media.retryHLS(neko)
  expect(retry).toHaveBeenCalled()
  expect(sent).toEqual([])
})

test('the console traces’ handle: states, the element and the peer, read-only, nothing secret', () => {
  page('https://neko.test/?media=ll-hls')
  createNekoApp()
  const trace = window.__nekoTrace!
  expect(Object.keys(trace).sort()).toEqual(['backend', 'connection', 'hls', 'peer', 'video'])
  expect(Object.isFrozen(trace)).toBe(true)
  expect(trace.backend()).toBe('ll-hls')
  expect(trace.connection()).toBe('disconnected')
  expect(trace.hls()).toEqual({ mode: 'll-hls', status: 'idle', detail: '' })
  expect(trace.peer()).toBeNull()
  page('https://neko.test/')
  createNekoApp()
  expect(window.__nekoTrace!.hls()).toBeNull()
  expect(window.__nekoTrace!.peer()).toBeNull() // no peer before connecting
})

test('without WebRTC the page offers HLS for this visit, the share link kept', () => {
  const loc = { href: 'https://neko.test/?usr=Kim#/ShareToken-12345', assign: vi.fn() }
  media.tryHLS(loc)
  expect(loc.assign).toHaveBeenCalledWith('https://neko.test/?usr=Kim&media=hls#/ShareToken-12345')
})
