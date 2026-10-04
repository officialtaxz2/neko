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

function harness({ eligible = true, autoplay = true, fetcher } = {}) {
  let now = 0, id = 0, revoked = false
  const timers = new Map(), status = [], events = [], requests = []
  const exports = {}
  const window = {
    setTimeout(callback, delay) { const key = ++id; timers.set(key, { at:now + delay, callback }); return key },
    clearTimeout(key) { timers.delete(key) },
  }
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
    navigator: { userAgent:'iPhone', vendor:'Apple Computer, Inc.' }, Date: { now: () => now },
    require(path) {
      if (path === '../events') return { EVENT: { HLS: { CAPABILITIES_REQUEST:'media/hls/capabilities/request', CREATE:'media/hls/create' } } }
      if (path === './protocol.js') return protocol
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
