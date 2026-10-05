import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'
import * as protocol from '../src/neko/hls/protocol.js'
import { capabilities, offer } from './hls-fixtures.mjs'

// Target-server tests only: transpilation lets the real controller run with
// deterministic HTTP/video/timer doubles; no browser or transport is started.
const source = await readFile(new URL('../src/neko/hls/controller.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
const masterPath = '/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8'
const manifest = '#EXTM3U\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",URI="audio/index.m3u8"\n#EXT-X-STREAM-INF:BANDWIDTH=650000,CODECS="avc1.64001f,mp4a.40.2",AUDIO="audio"\nlow/index.m3u8\n'
const child = '#EXTM3U\n#EXT-X-MAP:URI="init-1.mp4"\n#EXTINF:6,\nseg-1.m4s\n'
const flush = async () => { await new Promise(setImmediate); await new Promise(setImmediate) }

function harness({ eligible = true, autoplay = true, fetcher, mse = false } = {}) {
  let now = 0, id = 0, revoked = false
  const timers = new Map(), status = [], events = [], requests = []
  const exports = {}
  const window = {
    setTimeout(callback, delay) { const key = ++id; timers.set(key, { at:now + delay, callback }); return key },
    clearTimeout(key) { timers.delete(key) },
  }
  if (mse) window.MediaSource = { isTypeSupported: () => true }
  const video = {
    src: '', srcObject: null, paused: true, muted: false, volume: 1, currentTime: 0, readyState: 0,
    videoWidth: 640, videoHeight: 360, handlers: new Map(), playCalls: 0, loads: 0,
    canPlayType: () => 'maybe',
    addEventListener(event, callback) { this.handlers.set(event, callback) },
    removeEventListener(event, callback) { if (this.handlers.get(event) === callback) this.handlers.delete(event) },
    removeAttribute(name) { if (name === 'src') this.src = '' },
    load() { this.loads++ },
    play() { this.playCalls++; this.paused = false; return Promise.resolve() },
    pause() { this.paused = true },
    fire(event) { this.handlers.get(event)?.() },
  }
  runInNewContext(compiled, {
    exports, window, URL, AbortController, TextDecoder, location: { href:'https://neko.example/prefix/' },
    navigator: mse ? { userAgent:'Chrome', vendor:'Google Inc.' } : { userAgent:'iPhone', vendor:'Apple Computer, Inc.' },
    Date: { now: () => now },
    require(path) {
      if (path === '../events') return { EVENT: { HLS: { CAPABILITIES_REQUEST:'media/hls/capabilities/request', CREATE:'media/hls/create' } } }
      if (path === './protocol.js') return protocol
      if (path === './player' && mse) return {
        mseSupported: () => true,
        createMSEPlayer(video) { video.src = 'blob:synthetic-hls'; return { destroy() {} } },
      }
      throw new Error(`Unexpected module: ${path}`)
    },
    fetch: async (url, init) => {
      requests.push({ url, init })
      if (fetcher) return fetcher(url, init)
      if (url.endsWith('/session')) return new Response(JSON.stringify({ mode:'hls', master:masterPath, idle_expires_in_ms:30000 }), { status:201 })
      if (revoked) return new Response(null, { status:404 })
      if (url.endsWith('/index.m3u8')) return new Response(child)
      return url.endsWith('/keepalive') ? new Response(null, { status:204 }) : new Response(manifest)
    },
  })
  const controller = new exports.HLSMediaController('wss://neko.example/prefix/ws', 'hls', {
    sendEvent: (event, payload) => events.push({ event, payload }), eligible: () => eligible,
    eventSocketOpen: () => true, autoplay: () => autoplay, setStatus: (state, detail) => status.push({ state, detail }),
    setPlayer() {}, setPlayable() {}, setPlaying() {}, setMuted() {}, resolution() {},
  })
  controller.attach(video)
  const advance = async (ms) => {
    const end = now + ms
    for (;;) {
      const entries = [...timers].filter(([, timer]) => timer.at <= end).sort((a,b) => a[1].at - b[1].at)
      if (!entries.length) break
      const [key, timer] = entries[0]
      timers.delete(key); now = timer.at; timer.callback(); await flush()
    }
    now = end
    await flush()
  }
  const negotiate = async () => {
    controller.start(); controller.handleCapabilities(capabilities); controller.handleOffer(offer); await flush()
  }
  return { controller, video, timers, status, events, requests, advance, negotiate, revoke: () => { revoked = true } }
}

test('unadvertised and ineligible HLS modes terminate without a bootstrap or another backend', async () => {
  const disabled = harness()
  disabled.controller.start()
  await disabled.advance(5000)
  assert.equal(disabled.status.at(-1).state, 'terminal')
  assert.equal(disabled.requests.length, 0)
  assert.equal(disabled.events.length, 1)
  const ordinary = harness({ eligible:false })
  ordinary.controller.start()
  assert.equal(ordinary.status.at(-1).state, 'terminal')
  assert.equal(ordinary.events.length, 0)
})

test('private mode clears native buffers immediately and resumes the same lease', async () => {
  const h = harness()
  await h.negotiate()
  assert.match(h.video.src, /\/prefix\/api\/media\/hls\/[A-Za-z0-9_-]{22}\/master\.m3u8$/)
  const requests = h.requests.length
  h.controller.setPrivatePaused(true)
  assert.equal(h.video.src, '')
  assert.equal(h.video.srcObject, null)
  assert.equal(h.video.handlers.size, 0)
  assert.equal(h.status.at(-1).state, 'paused')
  h.controller.setPrivatePaused(false)
  await h.advance(1000)
  assert.ok(h.video.src)
  assert.equal(h.requests.filter(({url}) => url.endsWith('/session')).length, 1)
  assert.ok(h.requests.length > requests)
  h.controller.stop()
  await flush()
  assert.equal(h.video.src, '')
  assert.equal(h.timers.size, 0)
})

test('native lease revocation is detected and terminates playback without a fallback', async () => {
  const h = harness()
  await h.negotiate()
  h.revoke()
  await h.advance(1000)
  assert.equal(h.status.at(-1).state, 'terminal')
  assert.equal(h.video.src, '')
  assert.equal(h.video.handlers.size, 0)
  assert.equal(h.timers.size, 0)
  assert.equal(h.events.filter(({event}) => event.endsWith('/create')).length, 1)
})

test('a private pause longer than the lease lifetime keeps renewing and resumes the same lease', async () => {
  let paused = false
  const h = harness({ fetcher: async (url) => {
    if (url.endsWith('/session')) return new Response(JSON.stringify({ mode:'hls', master:masterPath, idle_expires_in_ms:30000 }), { status:201 })
    if (url.endsWith('/keepalive')) return new Response(null, { status:204 })
    if (paused) return new Response(null, { status:503 })
    return new Response(url.endsWith('/index.m3u8') ? child : manifest)
  } })
  await h.negotiate()
  paused = true
  h.controller.setPrivatePaused(true)
  await h.advance(46000)
  assert.equal(h.status.at(-1).state, 'paused')
  assert.equal(h.video.src, '')
  assert.equal(h.requests.filter(({url}) => url.endsWith('/keepalive')).length, 3)
  assert.equal(h.status.some(({state}) => state === 'terminal'), false)
  paused = false
  h.controller.setPrivatePaused(false)
  await h.advance(1000)
  assert.ok(h.video.src)
  assert.equal(h.requests.filter(({url}) => url.endsWith('/session')).length, 1)
  h.controller.stop()
  assert.equal(h.timers.size, 0)
})

test('late bootstrap completion after stop cannot reattach a source or start HTTP polling', async () => {
  let complete
  const h = harness({ fetcher: () => new Promise((resolve) => { complete = resolve }) })
  h.controller.start(); h.controller.handleCapabilities(capabilities); h.controller.handleOffer(offer)
  await flush()
  h.controller.stop()
  assert.equal(h.requests[0].init.signal.aborted, true)
  complete(new Response(JSON.stringify({ mode:'hls', master:masterPath, idle_expires_in_ms:30000 }), { status:201 }))
  await flush()
  assert.equal(h.video.src, '')
  assert.equal(h.requests.length, 1)
  assert.equal(h.timers.size, 0)
})

test('blocked autoplay retains a manual Play action after one muted attempt', async () => {
  const h = harness()
  await h.negotiate()
  h.video.play = () => { h.video.playCalls++; return Promise.reject(new Error('NotAllowedError')) }
  h.video.fire('canplay')
  await flush()
  assert.equal(h.video.playCalls, 2)
  assert.equal(h.video.muted, true)
  await h.advance(35000)
  assert.notEqual(h.status.at(-1).state, 'terminal')
  h.video.play = () => { h.video.playCalls++; h.video.paused = false; return Promise.resolve() }
  assert.equal(await h.controller.play(), true)
  h.controller.stop()
})

test('a pending Play rejection cannot undo a later user Pause', async () => {
  const h = harness()
  await h.negotiate()
  let reject
  h.video.play = () => { h.video.playCalls++; return new Promise((_, denied) => { reject = denied }) }
  const playing = h.controller.play()
  h.controller.pause()
  reject(new Error('AbortError'))
  assert.equal(await playing, false)
  assert.equal(h.video.playCalls, 1)
  assert.equal(h.video.muted, false)
  assert.equal(h.video.paused, true)
  h.controller.stop()
})

test('initial HLS readiness cannot later turn buffering into a startup timeout', async () => {
  for (const mse of [false, true]) {
    for (const event of ['canplay', 'playing']) {
      const h = harness({ mse })
      await h.negotiate()
      h.video.readyState = 3
      if (event === 'playing') h.video.paused = false
      h.video.fire(event)
      await flush()
      // Continue progressing beyond the original deadline while a brief
      // buffer underrun leaves only the current frame available (readyState 2).
      for (let second = 1; second <= 35; second++) {
        h.video.currentTime = second
        if (second === 29) h.video.readyState = 2
        if (second === 31) h.video.readyState = 3
        await h.advance(1000)
      }
      assert.equal(h.status.some(({state}) => state === 'terminal'), false,
        'ready HLS playback was terminated by the startup deadline')
      assert.ok(h.video.src)
      assert.equal(h.requests.filter(({url}) => url.endsWith('/keepalive')).length, 2)
      assert.equal(h.events.filter(({event}) => event.endsWith('/create')).length, 1)
      h.controller.stop()
      assert.equal(h.timers.size, 0)
    }
  }
})

test('HLS that never becomes playable still times out and clears resources', async () => {
  for (const mse of [false, true]) {
    const h = harness({ autoplay:false, mse })
    await h.negotiate()
    await h.advance(29999)
    assert.notEqual(h.status.at(-1).state, 'terminal')
    await h.advance(1)
    assert.deepEqual(h.status.at(-1), { state:'terminal', detail:'HLS playback did not become ready; retry manually' })
    assert.equal(h.video.src, '')
    assert.equal(h.video.handlers.size, 0)
    assert.equal(h.timers.size, 0)
  }
})

test('HLS playback progress watchdog still bounds a stall after initial readiness', async () => {
  for (const mse of [false, true]) {
    const h = harness({ mse })
    await h.negotiate()
    h.video.readyState = 3
    h.video.fire('canplay')
    await flush()
    h.video.fire('playing')
    h.video.readyState = 2
    await h.advance(20000)
    assert.equal(h.status.at(-1).state, 'streaming')
    await h.advance(1000)
    assert.deepEqual(h.status.at(-1), { state:'terminal', detail:'HLS playback stalled; retry manually' })
    assert.equal(h.video.src, '')
    assert.equal(h.timers.size, 0)
  }
})

test('an immediate native readiness event during attachment cancels the deadline', async () => {
  const h = harness({ autoplay:false })
  h.video.load = function () {
    this.loads++
    if (this.src) { this.readyState = 3; this.fire('canplay') }
  }
  await h.negotiate()
  h.video.readyState = 2
  await h.advance(35000)
  assert.equal(h.status.some(({state}) => state === 'terminal'), false)
  assert.ok(h.video.src)
  assert.equal(h.video.playCalls, 0)
  h.controller.stop()
  assert.equal(h.timers.size, 0)
})

test('stale readiness events cannot cancel the new player deadline after private resume', async () => {
  const h = harness({ autoplay:false })
  await h.negotiate()
  const staleCanplay = h.video.handlers.get('canplay')
  const stalePlaying = h.video.handlers.get('playing')
  h.controller.setPrivatePaused(true)
  h.controller.setPrivatePaused(false)
  await h.advance(1000)
  assert.ok(h.video.src)
  staleCanplay()
  stalePlaying()
  await h.advance(29999)
  assert.notEqual(h.status.at(-1).state, 'terminal')
  await h.advance(1)
  assert.deepEqual(h.status.at(-1), { state:'terminal', detail:'HLS playback did not become ready; retry manually' })
  assert.equal(h.timers.size, 0)
})
