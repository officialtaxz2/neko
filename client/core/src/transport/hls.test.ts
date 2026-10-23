// The HLS transport (hls.ts) with deterministic HTTP, video and timer doubles: negotiation, the lease,
// private mode, autoplay, the readiness deadline and the stall watch. Ported from the fork's Vue
// client tests; no browser or player starts (the MSE player is replaced).
import { afterEach, test, vi } from 'vitest'
import assert from 'node:assert/strict'
import { Emitter } from '../emitter.ts'
import { capabilities, offer } from '../test/hls.ts'

const mse = vi.hoisted(() => ({ attach: undefined as undefined | ((video: any) => void) }))
vi.mock('./hls/player.ts', () => ({
  mseSupported: () => true,
  createMSEPlayer(video: any) {
    video.src = 'blob:synthetic-hls'
    mse.attach?.(video)
    return { destroy() {} }
  },
}))
const { HLSTransport } = await import('./hls.ts')

const masterPath = '/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8'
const manifest =
  '#EXTM3U\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",URI="audio/index.m3u8"\n#EXT-X-STREAM-INF:BANDWIDTH=650000,CODECS="avc1.64001f,mp4a.40.2",AUDIO="audio"\nlow/index.m3u8\n'
const child = '#EXTM3U\n#EXT-X-MAP:URI="init-1.mp4"\n#EXTINF:6,\nseg-1.m4s\n'
const lease = () =>
  new Response(JSON.stringify({ mode: 'hls', master: masterPath, idle_expires_in_ms: 30000 }), { status: 201 })
const flush = async () => {
  await new Promise(setImmediate)
  await new Promise(setImmediate)
}

// a <video> as far as the transport uses it; handlers holds the player's listeners, one per event
class Video {
  src = ''
  srcObject: unknown = null
  paused = true
  seeking = false
  volume = 1
  currentTime = 0
  readyState = 0
  videoWidth = 640
  videoHeight = 360
  playsInline = false
  style: Record<string, string> = {}
  playCalls = 0
  loads = 0
  handlers = new Map<string, () => void>()
  private _muted = false
  private volumeListeners: (() => void)[] = []
  get muted() {
    return this._muted
  }
  set muted(v: boolean) {
    this._muted = v
    this.volumeListeners.forEach((fn) => fn())
  }
  canPlayType() {
    return 'maybe'
  }
  addEventListener(event: string, callback: () => void) {
    if (event === 'volumechange') this.volumeListeners.push(callback)
    else this.handlers.set(event, callback)
  }
  removeEventListener(event: string, callback: () => void) {
    if (this.handlers.get(event) === callback) this.handlers.delete(event)
  }
  removeAttribute(name: string) {
    if (name === 'src') this.src = ''
  }
  load() {
    this.loads++
  }
  play(): Promise<void> {
    this.playCalls++
    this.paused = false
    return Promise.resolve()
  }
  pause() {
    this.paused = true
  }
  remove() {}
  fire(event: string) {
    this.handlers.get(event)?.()
  }
}

type Options = {
  autoplay?: boolean
  fetcher?: (url: string, init: RequestInit) => Promise<Response>
  mse?: boolean
  mseAttach?: (video: any) => void
  page?: string
  fetch?: boolean
}

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
  mse.attach = undefined
})

function harness({ autoplay = true, fetcher, mse: useMSE = false, mseAttach, page, fetch = true }: Options = {}) {
  let now = 0
  let id = 0
  let revoked = false
  const timers = new Map<number, { at: number; callback: () => void }>()
  const status: { state: string; detail: string }[] = []
  const playing: boolean[] = []
  const playable: boolean[] = []
  const transport: string[] = []
  const requests: { url: string; init: RequestInit }[] = []
  const video = new Video()
  mse.attach = mseAttach
  vi.stubGlobal('window', {
    setTimeout(callback: () => void, delay: number) {
      const key = ++id
      timers.set(key, { at: now + delay, callback })
      return key
    },
    clearTimeout(key: number) {
      timers.delete(key)
    },
    ...(useMSE && { MediaSource: { isTypeSupported: () => true } }),
  })
  vi.spyOn(Date, 'now').mockImplementation(() => now)
  vi.stubGlobal('location', { href: page ?? 'https://neko.example/prefix/' })
  vi.stubGlobal(
    'navigator',
    useMSE ? { userAgent: 'Chrome', vendor: 'Google Inc.' } : { userAgent: 'iPhone', vendor: 'Apple Computer, Inc.' },
  )
  vi.stubGlobal('document', { createElement: () => video, addEventListener() {} })
  vi.stubGlobal(
    'fetch',
    fetch
      ? async (url: string, init: RequestInit) => {
          requests.push({ url, init })
          if (fetcher) return fetcher(url, init)
          if (url.endsWith('/session')) return lease()
          if (revoked) return new Response(null, { status: 404 })
          if (url.endsWith('/index.m3u8')) return new Response(child)
          return url.endsWith('/keepalive') ? new Response(null, { status: 204 }) : new Response(manifest)
        }
      : undefined,
  )

  // the session as the client hands it over, its socket recorded
  const events: { event: string; payload: any }[] = []
  const messages = new Emitter<Record<string, (p: any) => void>>()
  const session = {
    url: 'https://neko.example/prefix',
    token: 'tok',
    init: {
      session_id: 's1',
      sessions: { s1: { id: 's1', profile: { is_admin: false } } },
      settings: { private_mode: false },
    } as any,
    send: (event: string, payload?: unknown) => events.push({ event, payload }),
    on: (event: string, cb: (p: any) => void) => messages.on(event, cb),
  }
  const receive = (event: string, payload?: unknown) => messages.emit(event, payload)

  const t = new HLSTransport({ mode: 'hls', autoplay: () => autoplay })
  t.on('hls', (s) => status.push({ state: s.status, detail: s.detail }))
  t.on('state', (s) => {
    playing.push(s.video.playing)
    playable.push(s.video.playable)
    transport.push(s.status)
  })
  t.attach({ append() {} } as unknown as HTMLElement)

  const advance = async (ms: number) => {
    const end = now + ms
    for (;;) {
      const entries = [...timers].filter(([, timer]) => timer.at <= end).sort((a, b) => a[1].at - b[1].at)
      if (!entries.length) break
      const [key, timer] = entries[0]
      timers.delete(key)
      now = timer.at
      timer.callback()
      await flush()
    }
    now = end
    await flush()
  }
  const start = () => void t.connect(session)
  const negotiate = async () => {
    start()
    receive('media/hls/capabilities', capabilities)
    receive('media/hls/offer', offer)
    await flush()
  }
  // the room's private mode, as system/settings announces it
  const privateMode = (on: boolean) => receive('system/settings', { id: 's2', private_mode: on })
  return {
    t,
    video,
    timers,
    status,
    events,
    requests,
    playing,
    playable,
    transport,
    session,
    receive,
    advance,
    start,
    negotiate,
    privateMode,
    revoke: () => (revoked = true),
  }
}
const count = (requests: { url: string }[], suffix: string) => requests.filter(({ url }) => url.endsWith(suffix)).length

test('an unadvertised mode, a page without fetch or a plain-HTTP page stop without a lease or another backend', async () => {
  const disabled = harness()
  disabled.start()
  assert.deepEqual(disabled.events, [{ event: 'media/hls/capabilities/request', payload: { version: 1, mode: 'hls' } }])
  await disabled.advance(5000)
  assert.equal(disabled.status.at(-1)!.state, 'terminal')
  assert.equal(disabled.requests.length, 0)
  assert.equal(disabled.events.length, 1)
  assert.equal(disabled.transport.at(-1), 'unavailable', 'the room stays usable without a picture')

  const noFetch = harness({ fetch: false })
  noFetch.start()
  assert.deepEqual(noFetch.status.at(-1), {
    state: 'terminal',
    detail: 'HLS requires an authenticated session and HTTP playback support',
  })
  assert.equal(noFetch.events.length, 0)

  const http = harness({ page: 'http://neko.example/prefix/' })
  http.start()
  assert.deepEqual(http.status.at(-1), { state: 'terminal', detail: 'HLS requires same-origin HTTPS' })
  assert.equal(http.events.length, 0)
  assert.equal(http.requests.length, 0)
})

test('negotiation: capabilities, create, the ticket only in the bootstrap body, then the lease is watched', async () => {
  const h = harness()
  h.start()
  assert.equal(h.status.at(-1)!.state, 'negotiating')
  assert.equal(h.transport.at(-1), 'connecting')
  h.receive('media/hls/capabilities', { ...capabilities, modes: ['ll-hls'] }) // not this transport's mode
  assert.equal(h.status.at(-1)!.state, 'terminal')

  const g = harness()
  await g.negotiate()
  assert.deepEqual(
    g.events.map((e) => e.event),
    ['media/hls/capabilities/request', 'media/hls/create'],
  )
  assert.deepEqual(g.events[1].payload, { version: 1, backend: 'hls', mode: 'hls' })
  const bootstrap = g.requests[0]
  assert.equal(bootstrap.url, 'https://neko.example/prefix/api/media/hls/session')
  assert.equal(bootstrap.init.body, JSON.stringify({ ticket: offer.ticket }))
  assert.ok(g.requests.every(({ url }) => !url.includes(offer.ticket)))
  assert.equal(bootstrap.init.credentials, 'same-origin')
  assert.equal(bootstrap.init.redirect, 'error')
  assert.equal(g.transport.at(-1), 'connected', 'the server granted the stream: the room is up')
  assert.equal(g.t.hls.player, 'native')
  assert.match(g.video.src, /\/prefix\/api\/media\/hls\/[A-Za-z0-9_-]{22}\/master\.m3u8$/)
})

test('private mode: the room’s setting (not for admins) clears native buffers and resumes the same lease', async () => {
  const h = harness()
  await h.negotiate()
  assert.match(h.video.src, /\/prefix\/api\/media\/hls\/[A-Za-z0-9_-]{22}\/master\.m3u8$/)
  const requests = h.requests.length
  h.privateMode(true)
  assert.equal(h.video.src, '')
  assert.equal(h.video.srcObject, null)
  assert.equal(h.video.handlers.size, 0)
  assert.equal(h.status.at(-1)!.state, 'paused')
  h.privateMode(false)
  await h.advance(1000)
  assert.ok(h.video.src)
  assert.equal(count(h.requests, '/session'), 1)
  assert.ok(h.requests.length > requests)
  h.receive('session/profile', { id: 's1', is_admin: true })
  h.privateMode(true)
  assert.ok(h.video.src, 'an admin keeps the picture in private mode')
  h.t.close()
  await flush()
  assert.equal(h.video.src, '')
  assert.equal(h.timers.size, 0)

  // a room already private when the session starts
  const p = harness()
  p.session.init.settings.private_mode = true
  await p.negotiate()
  await p.advance(1000)
  assert.equal(p.video.src, '')
  p.privateMode(false)
  await p.advance(1000)
  assert.ok(p.video.src)
})

test('native lease revocation is noticed and ends playback without a fallback', async () => {
  const h = harness()
  await h.negotiate()
  h.revoke()
  await h.advance(1000)
  assert.equal(h.status.at(-1)!.state, 'terminal')
  assert.equal(h.video.src, '')
  assert.equal(h.video.handlers.size, 0)
  assert.equal(h.timers.size, 0)
  assert.equal(h.events.filter(({ event }) => event.endsWith('/create')).length, 1)
})

test('retry starts over with a new negotiation; a reconnect (connect again) does as well', async () => {
  const h = harness()
  await h.negotiate()
  h.revoke()
  await h.advance(1000)
  assert.equal(h.status.at(-1)!.state, 'terminal')
  h.t.retry()
  assert.equal(h.status.at(-1)!.state, 'negotiating')
  assert.equal(h.events.filter(({ event }) => event.endsWith('/capabilities/request')).length, 2)
  h.t.close()
  assert.equal(h.status.at(-1)!.state, 'idle')
  assert.equal(h.transport.at(-1), 'disconnected')
  h.t.retry() // closed: nothing to retry
  assert.equal(h.events.filter(({ event }) => event.endsWith('/capabilities/request')).length, 2)
  h.start()
  assert.equal(h.events.filter(({ event }) => event.endsWith('/capabilities/request')).length, 3)
  h.t.close()
  assert.equal(h.timers.size, 0)
})

test('a private pause longer than the lease lifetime keeps renewing and resumes the same lease', async () => {
  let paused = false
  const h = harness({
    fetcher: async (url) => {
      if (url.endsWith('/session')) return lease()
      if (url.endsWith('/keepalive')) return new Response(null, { status: 204 })
      if (paused) return new Response(null, { status: 503 })
      return new Response(url.endsWith('/index.m3u8') ? child : manifest)
    },
  })
  await h.negotiate()
  paused = true
  h.privateMode(true)
  await h.advance(46000)
  assert.equal(h.status.at(-1)!.state, 'paused')
  assert.equal(h.video.src, '')
  assert.equal(count(h.requests, '/keepalive'), 3)
  assert.equal(
    h.status.some(({ state }) => state === 'terminal'),
    false,
  )
  paused = false
  h.privateMode(false)
  await h.advance(1000)
  assert.ok(h.video.src)
  assert.equal(count(h.requests, '/session'), 1)
  h.t.close()
  assert.equal(h.timers.size, 0)
})

test('a bootstrap that completes after close cannot attach a source or start polling', async () => {
  let complete!: (r: Response) => void
  const h = harness({ fetcher: () => new Promise((resolve) => (complete = resolve)) })
  await h.negotiate()
  h.t.close()
  assert.equal(h.requests[0].init.signal!.aborted, true)
  complete(lease())
  await flush()
  assert.equal(h.video.src, '')
  assert.equal(h.requests.length, 1)
  assert.equal(h.timers.size, 0)
})

test('blocked autoplay: one muted attempt, then a manual Play still works', async () => {
  for (const useMSE of [false, true]) {
    const h = harness({ mse: useMSE })
    h.video.play = () => {
      h.video.playCalls++
      return Promise.reject(new Error('NotAllowedError'))
    }
    await h.negotiate()
    assert.equal(h.video.playCalls, 2)
    assert.equal(h.video.muted, true)
    h.video.readyState = 3
    h.video.fire('canplay')
    await flush()
    assert.equal(h.video.playCalls, 2, 'readiness must not repeat blocked autoplay')
    assert.equal(h.playable.at(-1), true)
    await h.advance(35000)
    assert.notEqual(h.status.at(-1)!.state, 'terminal')
    h.video.play = () => {
      h.video.playCalls++
      h.video.paused = false
      return Promise.resolve()
    }
    assert.equal(await h.t.play(), true)
    h.t.close()
  }
})

test('muted by autoplay: reported, and unmuting clears it', async () => {
  const h = harness()
  let first = true
  h.video.play = () => {
    h.video.playCalls++
    if (first) {
      first = false
      return Promise.reject(new Error('NotAllowedError'))
    }
    h.video.paused = false
    return Promise.resolve()
  }
  const states: any[] = []
  h.t.on('state', (s) => states.push(s.video))
  await h.negotiate()
  assert.equal(states.at(-1).muted, true)
  assert.equal(states.at(-1).mutedByAutoplay, true)
  h.t.setMuted(false)
  assert.equal(states.at(-1).muted, false)
  assert.equal(states.at(-1).mutedByAutoplay, false)
  h.t.setVolume(2)
  assert.equal(h.video.volume, 1)
  h.t.close()
})

test('a pending Play rejection cannot undo a later Pause', async () => {
  const h = harness({ autoplay: false })
  await h.negotiate()
  let reject!: (e: Error) => void
  h.video.play = () => {
    h.video.playCalls++
    return new Promise((_, denied) => (reject = denied))
  }
  const playing = h.t.play()
  void h.t.setPlaying(false)
  reject(new Error('AbortError'))
  assert.equal(await playing, false)
  assert.equal(h.video.playCalls, 1)
  assert.equal(h.video.muted, false)
  assert.equal(h.video.paused, true)
  h.t.close()
})

test('initial autoplay is asked for even without a canplay event', async () => {
  for (const useMSE of [false, true]) {
    for (const readyState of [2, 0]) {
      const h = harness({
        mse: useMSE,
        mseAttach(video) {
          video.readyState = readyState
        },
      })
      h.video.load = function (this: Video) {
        this.loads++
        if (this.src) this.readyState = readyState
      }
      let complete!: () => void
      h.video.play = () => {
        h.video.playCalls++
        h.video.paused = false
        return new Promise((resolve) => (complete = resolve))
      }
      await h.negotiate()
      assert.equal(h.video.playCalls, 1, 'initial HLS autoplay waited for canplay')
      assert.equal(h.video.paused, false)
      assert.equal(h.playable.at(-1), false, 'a Play request is not proof of readiness')
      assert.equal(h.status.at(-1)!.state, 'connecting')
      // a pending early Play gets the whole readiness budget, not the shorter stall watch
      await h.advance(25000)
      assert.equal(
        h.status.some(({ state }) => state === 'terminal'),
        false,
      )
      h.video.readyState = 3
      h.video.fire('canplay')
      assert.equal(h.video.playCalls, 1)
      complete()
      await flush()
      h.video.fire('playing')
      for (let second = 1; second <= 10; second++) {
        h.video.currentTime = second
        await h.advance(1000)
      }
      assert.equal(h.status.at(-1)!.state, 'streaming')
      assert.equal(count(h.requests, '/session'), 1)
      h.t.close()
      assert.equal(h.video.src, '')
      assert.equal(h.timers.size, 0)
    }
  }
})

test('autoplay on attaching respects autoplay turned off and a Pause before readiness', async () => {
  for (const useMSE of [false, true]) {
    const manual = harness({ mse: useMSE, autoplay: false })
    await manual.negotiate()
    assert.equal(manual.video.playCalls, 0)
    manual.video.readyState = 3
    manual.video.fire('canplay')
    await flush()
    assert.equal(manual.video.playCalls, 0)
    assert.equal(await manual.t.play(), true)
    assert.equal(manual.video.playCalls, 1)
    manual.t.close()

    const h = harness({ mse: useMSE })
    let reject!: (e: Error) => void
    h.video.play = () => {
      h.video.playCalls++
      h.video.paused = false
      return new Promise((_, denied) => (reject = denied))
    }
    await h.negotiate()
    assert.equal(h.video.playCalls, 1)
    void h.t.setPlaying(false)
    reject(new Error('AbortError'))
    await flush()
    h.video.readyState = 3
    h.video.fire('canplay')
    await h.advance(35000)
    assert.equal(h.video.playCalls, 1, 'a paused startup must not retry autoplay')
    assert.equal(h.video.muted, false)
    assert.equal(h.video.paused, true)
    assert.equal(
      h.status.some(({ state }) => state === 'terminal'),
      false,
    )
    h.t.close()
    assert.equal(h.timers.size, 0)
  }
})

test('an early autoplay rejection after a private pause cannot affect the resumed player', async () => {
  for (const useMSE of [false, true]) {
    const h = harness({ mse: useMSE })
    let reject!: (e: Error) => void
    h.video.play = () => {
      h.video.playCalls++
      h.video.paused = false
      return new Promise((_, denied) => (reject = denied))
    }
    await h.negotiate()
    h.privateMode(true)
    h.video.play = () => {
      h.video.playCalls++
      h.video.paused = false
      return Promise.resolve()
    }
    h.privateMode(false)
    await h.advance(1000)
    h.video.readyState = 3
    h.video.fire('canplay')
    h.video.fire('playing')
    reject(new Error('AbortError'))
    await flush()
    assert.equal(h.video.playCalls, 2)
    assert.equal(h.video.muted, false)
    assert.equal(h.video.paused, false)
    assert.equal(h.status.at(-1)!.state, 'streaming')
    assert.equal(count(h.requests, '/session'), 1)
    h.t.close()
    assert.equal(h.timers.size, 0)
  }
})

test('once ready, buffering is never mistaken for a startup timeout', async () => {
  for (const useMSE of [false, true]) {
    for (const event of ['canplay', 'playing']) {
      const h = harness({ mse: useMSE })
      await h.negotiate()
      h.video.readyState = 3
      if (event === 'playing') h.video.paused = false
      h.video.fire(event)
      await flush()
      // progressing beyond the first deadline while a short underrun leaves only the current frame
      for (let second = 1; second <= 35; second++) {
        h.video.currentTime = second
        if (second === 29) h.video.readyState = 2
        if (second === 31) h.video.readyState = 3
        await h.advance(1000)
      }
      assert.equal(
        h.status.some(({ state }) => state === 'terminal'),
        false,
        'ready HLS playback was terminated by the startup deadline',
      )
      assert.ok(h.video.src)
      assert.equal(count(h.requests, '/keepalive'), 2)
      assert.equal(h.events.filter(({ event }) => event.endsWith('/create')).length, 1)
      h.t.close()
      assert.equal(h.timers.size, 0)
    }
  }
})

test('HLS that never becomes playable still times out and clears everything', async () => {
  for (const useMSE of [false, true]) {
    for (const pendingPlay of [false, true]) {
      const h = harness({ autoplay: false, mse: useMSE })
      await h.negotiate()
      let complete!: () => void
      let playing: Promise<boolean> | undefined
      if (pendingPlay) {
        h.video.play = () => {
          h.video.paused = false
          return new Promise((resolve) => (complete = resolve))
        }
        playing = h.t.play()
      }
      await h.advance(29999)
      assert.notEqual(h.status.at(-1)!.state, 'terminal')
      await h.advance(1)
      assert.deepEqual(h.status.at(-1), {
        state: 'terminal',
        detail: 'HLS playback did not become ready; retry manually',
      })
      assert.equal(h.video.src, '')
      assert.equal(h.video.handlers.size, 0)
      assert.equal(h.timers.size, 0)
      if (pendingPlay) {
        complete()
        assert.equal(await playing, false)
      }
    }
  }
})

test('the progress watch still bounds a stall after the first readiness', async () => {
  for (const useMSE of [false, true]) {
    const h = harness({ mse: useMSE })
    await h.negotiate()
    h.video.readyState = 3
    h.video.fire('canplay')
    await flush()
    h.video.fire('playing')
    h.video.readyState = 2
    await h.advance(20000)
    assert.equal(h.status.at(-1)!.state, 'streaming')
    await h.advance(1000)
    assert.deepEqual(h.status.at(-1), { state: 'terminal', detail: 'HLS playback stalled; retry manually' })
    assert.equal(h.video.src, '')
    assert.equal(h.timers.size, 0)
  }
})

test('live-edge seeks without current data cannot hide a stall', async () => {
  for (const useMSE of [false, true]) {
    const h = harness({ mse: useMSE })
    await h.negotiate()
    h.video.readyState = 3
    h.video.fire('playing')
    // one second of real playback, then the captured failure: six-second jumps without a current
    // frame while HTTP stays healthy
    h.video.currentTime = 1
    await h.advance(1000)
    h.video.readyState = 1
    for (let second = 1; second <= 21; second++) {
      if (second % 6 === 0) {
        h.video.currentTime += 6
        h.video.seeking = true
      }
      await h.advance(1000)
    }
    assert.deepEqual(
      h.status.at(-1),
      { state: 'terminal', detail: 'HLS playback stalled; retry manually' },
      'live-edge seeks hid the frozen HLS playback',
    )
    assert.equal(h.video.src, '')
    assert.equal(h.video.handlers.size, 0)
    assert.equal(h.timers.size, 0)
    assert.equal(
      h.requests.every(({ init }) => init.signal!.aborted),
      true,
    )
    assert.equal(h.events.filter(({ event }) => event.endsWith('/create')).length, 1)
  }
})

test('seeks completed between two probes do not renew the stall budget', async () => {
  const h = harness()
  await h.negotiate()
  h.video.readyState = 3
  h.video.fire('playing')
  for (let second = 1; second <= 21; second++) {
    if (second % 6 === 0) {
      h.video.currentTime += 6
      h.video.seeking = true
      h.video.fire('seeking')
      h.video.seeking = false
      h.video.fire('seeked')
    }
    await h.advance(1000)
  }
  assert.deepEqual(h.status.at(-1), { state: 'terminal', detail: 'HLS playback stalled; retry manually' })
  assert.equal(h.timers.size, 0)
})

test('natural playback after a completed seek still renews the stall budget', async () => {
  const h = harness()
  await h.negotiate()
  h.video.readyState = 3
  h.video.fire('playing')
  await h.advance(19000)
  h.video.currentTime = 30
  h.video.seeking = true
  h.video.fire('seeking')
  h.video.seeking = false
  h.video.fire('seeked')
  for (let second = 1; second <= 25; second++) {
    h.video.currentTime = 30 + second
    await h.advance(1000)
  }
  assert.equal(h.status.at(-1)!.state, 'streaming')
  assert.ok(h.video.src)
  h.t.close()
  assert.equal(h.timers.size, 0)
})

test('an immediate native readiness event while attaching cancels the deadline', async () => {
  const h = harness({ autoplay: false })
  h.video.load = function (this: Video) {
    this.loads++
    if (this.src) {
      this.readyState = 3
      this.fire('canplay')
    }
  }
  await h.negotiate()
  h.video.readyState = 2
  await h.advance(35000)
  assert.equal(
    h.status.some(({ state }) => state === 'terminal'),
    false,
  )
  assert.ok(h.video.src)
  assert.equal(h.video.playCalls, 0)
  h.t.close()
  assert.equal(h.timers.size, 0)
})

test('stale readiness events cannot cancel the new player’s deadline after a private resume', async () => {
  const h = harness({ autoplay: false })
  await h.negotiate()
  const staleCanplay = h.video.handlers.get('canplay')!
  const stalePlaying = h.video.handlers.get('playing')!
  h.privateMode(true)
  h.privateMode(false)
  await h.advance(1000)
  assert.ok(h.video.src)
  staleCanplay()
  stalePlaying()
  await h.advance(29999)
  assert.notEqual(h.status.at(-1)!.state, 'terminal')
  await h.advance(1)
  assert.deepEqual(h.status.at(-1), {
    state: 'terminal',
    detail: 'HLS playback did not become ready; retry manually',
  })
  assert.equal(h.timers.size, 0)
})

test('a pending first Play keeps the startup budget apart from the stall budget', async () => {
  for (const useMSE of [false, true]) {
    const h = harness({ mse: useMSE, autoplay: false })
    await h.negotiate()
    let complete!: () => void
    h.video.play = () => {
      h.video.playCalls++
      h.video.paused = false
      return new Promise((resolve) => (complete = resolve))
    }
    const playing = h.t.play()
    await h.advance(25000)
    assert.equal(
      h.status.some(({ state }) => state === 'terminal'),
      false,
    )
    h.video.readyState = 3
    h.video.fire('canplay')
    complete()
    assert.equal(await playing, true)
    h.video.fire('playing')
    await h.advance(20000)
    assert.equal(h.status.at(-1)!.state, 'streaming')
    await h.advance(1000)
    assert.deepEqual(h.status.at(-1), { state: 'terminal', detail: 'HLS playback stalled; retry manually' })
    assert.equal(h.video.src, '')
    assert.equal(h.timers.size, 0)
  }
})

test('a long deliberate Pause does not count as a stall when Play resumes', async () => {
  for (const useMSE of [false, true]) {
    const h = harness({ mse: useMSE })
    await h.negotiate()
    h.video.readyState = 3
    h.video.fire('canplay')
    await flush()
    h.video.fire('playing')
    h.video.currentTime = 1
    await h.advance(1000)
    void h.t.setPlaying(false)
    await h.advance(46000)
    let complete!: () => void
    h.video.play = () => {
      h.video.paused = false
      return new Promise((resolve) => (complete = resolve))
    }
    const resumed = h.t.play()
    // paused turns false before the playing event arrives
    await h.advance(1000)
    assert.equal(
      h.status.some(({ state }) => state === 'terminal'),
      false,
    )
    complete()
    assert.equal(await resumed, true)
    h.video.fire('playing')
    for (let second = 2; second <= 6; second++) {
      h.video.currentTime = second
      await h.advance(1000)
    }
    assert.equal(h.status.at(-1)!.state, 'streaming')
    assert.equal(count(h.requests, '/session'), 1)
    h.t.close()
    assert.equal(h.timers.size, 0)
  }
})

test('repeated Play on an unpaused frozen player cannot extend the stall deadline', async () => {
  const h = harness()
  await h.negotiate()
  h.video.readyState = 3
  h.video.fire('canplay')
  await flush()
  h.video.fire('playing')
  await h.advance(19000)
  assert.equal(await h.t.play(), true)
  await h.advance(2000)
  assert.deepEqual(h.status.at(-1), { state: 'terminal', detail: 'HLS playback stalled; retry manually' })
  assert.equal(h.timers.size, 0)
})

test('a playing event already queued cannot report playback after a deliberate Pause', async () => {
  const h = harness()
  await h.negotiate()
  h.video.readyState = 3
  h.video.fire('canplay')
  await flush()
  const queuedPlaying = h.video.handlers.get('playing')!
  void h.t.setPlaying(false)
  queuedPlaying()
  assert.equal(h.playing.at(-1), false)
  assert.notEqual(h.status.at(-1)!.state, 'streaming')
  await h.advance(35000)
  assert.equal(
    h.status.some(({ state }) => state === 'terminal'),
    false,
  )
  h.t.close()
})

test('immediate MSE readiness completes autoplay after attaching and keeps manual Play', async () => {
  for (const autoplay of [true, false]) {
    const h = harness({
      mse: true,
      autoplay,
      mseAttach(video) {
        video.readyState = 3
        video.fire('canplay')
      },
    })
    await h.negotiate()
    assert.equal(h.video.playCalls, autoplay ? 1 : 0)
    assert.equal(h.playable.at(-1), true)
    assert.equal(h.t.hls.player, 'mse')
    if (autoplay) h.video.fire('playing')
    for (let second = 1; second <= 35; second++) {
      h.video.currentTime = second
      await h.advance(1000)
    }
    assert.equal(
      h.status.some(({ state }) => state === 'terminal'),
      false,
    )
    h.t.close()
    assert.equal(h.timers.size, 0)
  }
})

test('not-ready probes and consecutive HTTP failures have budgets of their own', async () => {
  let state = 'ready'
  const h = harness({
    autoplay: false,
    fetcher: async (url) => {
      if (url.endsWith('/session')) return lease()
      if (url.endsWith('/keepalive')) return new Response(null, { status: 204 })
      if (state === 'fault') throw new Error('synthetic network failure')
      if (state === 'warming') return new Response(null, { status: 503 })
      return new Response(url.endsWith('/index.m3u8') ? child : manifest)
    },
  })
  await h.negotiate()
  state = 'warming'
  await h.advance(3000)
  state = 'fault'
  await h.advance(1000)
  assert.equal(
    h.status.some(({ state }) => state === 'terminal'),
    false,
    'one HTTP failure was added to the not-ready probes',
  )
  state = 'warming'
  await h.advance(1000)
  state = 'ready'
  await h.advance(1000)
  assert.ok(h.video.src)
  state = 'fault'
  await h.advance(2000)
  assert.equal(
    h.status.some(({ state }) => state === 'terminal'),
    false,
  )
  await h.advance(1000)
  assert.deepEqual(h.status.at(-1), { state: 'terminal', detail: 'HLS HTTP connection failed; retry manually' })
  assert.equal(h.video.src, '')
  assert.equal(h.timers.size, 0)
})

test('a private resume gets a fresh bounded readiness interval while renewing the same lease', async () => {
  const h = harness({
    autoplay: false,
    fetcher: async (url) => {
      if (url.endsWith('/session')) return lease()
      if (url.endsWith('/keepalive')) return new Response(null, { status: 204 })
      return new Response(null, { status: 503 })
    },
  })
  await h.negotiate()
  // negotiate() includes the first not-ready probe at t=0
  await h.advance(28000)
  assert.equal(h.status.at(-1)!.state, 'connecting')
  h.privateMode(true)
  await h.advance(46000)
  assert.equal(h.status.at(-1)!.state, 'paused')
  h.privateMode(false)
  await h.advance(29000)
  assert.equal(h.status.at(-1)!.state, 'connecting')
  await h.advance(1000)
  assert.deepEqual(h.status.at(-1), { state: 'terminal', detail: 'HLS media did not become ready' })
  assert.equal(count(h.requests, '/session'), 1)
  assert.ok(count(h.requests, '/keepalive') >= 6)
  assert.equal(h.timers.size, 0)
})

test('an unavailable bootstrap has a fixed message and waits for an explicit retry', async () => {
  for (const status of [503, 500]) {
    const h = harness({
      fetcher: async () => new Response('untrusted-secret https://other.example/private', { status }),
    })
    await h.negotiate()
    assert.deepEqual(h.status.at(-1), {
      state: 'terminal',
      detail:
        status === 503
          ? 'HLS media was unavailable at startup; retry manually'
          : 'HLS bootstrap failed; retry manually',
    })
    await h.advance(60000)
    assert.equal(h.requests.length, 1)
    assert.equal(h.events.filter(({ event }) => event.endsWith('/create')).length, 1)
    assert.equal(h.video.src, '')
    assert.equal(h.timers.size, 0)
    assert.doesNotMatch(h.status.at(-1)!.detail, /secret|example|private/)
  }
})
