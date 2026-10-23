// The MSE player (player.ts) with hls.js replaced: what its loader lets through, how many requests
// run at once, and the latency settings it hands hls.js
import { beforeEach, test, vi } from 'vitest'
import assert from 'node:assert/strict'

const mock = vi.hoisted(() => ({ instance: null as any }))
vi.mock('hls.js', () => ({
  default: class MockHls {
    static Events = { ERROR: 'error' }
    static isSupported = () => true
    config: any
    loaders: any[] = []
    liveSyncPosition: number | null = null
    constructor(config: any) {
      this.config = config
      mock.instance = this
    }
    on() {}
    attachMedia() {}
    loadSource() {}
    destroy() {
      for (const loader of this.loaders) loader.destroy()
    }
  },
}))
const { createMSEPlayer } = await import('./player.ts')

const master = 'https://neko.example/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8'
const flush = async () => {
  await new Promise(setImmediate)
  await new Promise(setImmediate)
}
type Request = { url: string; init: RequestInit; complete: (r: Response) => void }
let requests: Request[]

beforeEach(() => {
  requests = []
  vi.stubGlobal('window', globalThis)
  // a server that answers when the test says so, and ignores abort, so late completions happen
  vi.stubGlobal(
    'fetch',
    (url: string, init: RequestInit) => new Promise<Response>((complete) => requests.push({ url, init, complete })),
  )
})

function harness(mode: 'hls' | 'll-hls' = 'll-hls', workerPath?: string) {
  const terminal: string[] = []
  const success: any[] = []
  const video = Object.assign(new EventTarget(), {
    currentTime: 3,
    paused: false,
    seeking: false,
    buffered: { length: 1, end: () => 7 },
  })
  const player = createMSEPlayer(
    video as unknown as HTMLVideoElement,
    master,
    mode,
    (detail) => {
      terminal.push(detail)
      player.destroy()
    },
    workerPath,
  )
  const hls = mock.instance
  const loader = () => {
    const value = new hls.config.loader()
    hls.loaders.push(value)
    return value
  }
  const load = (value: any, path = 'high/seg-42.m4s', responseType = 'arraybuffer') =>
    value.load(
      { url: new URL(path, master).href, responseType, type: 'media-fragment' },
      { loadPolicy: { maxLoadTimeMs: 12000 } },
      { onSuccess: (response: unknown) => success.push(response), onError() {}, onTimeout() {} },
    )
  return { player, terminal, success, loader, load, video, hls, config: hls.config }
}

test('three requests at most, the lease’s fourth slot stays free; destroy drops what is queued', async () => {
  const h = harness()
  for (let i = 0; i < 4; i++) h.load(h.loader())
  assert.equal(requests.length, 3)
  h.player.destroy()
  assert.ok(requests.every(({ init }) => init.signal!.aborted))
  for (const request of requests) request.complete(new Response(new Uint8Array([1])))
  await flush()
  assert.equal(requests.length, 3)
  assert.equal(h.success.length, 0)
})

test('external and credential-bearing URLs stop the player before any request', () => {
  for (const path of ['https://other.example/seg-42.m4s', 'high/seg-42.m4s?ticket=secret']) {
    const h = harness()
    h.load(h.loader(), path)
    assert.equal(requests.length, 0)
    assert.equal(h.terminal.length, 1)
    assert.doesNotMatch(h.terminal[0], /secret|example|abcdefghijkl/)
  }
})

test('a reused loader never delivers the completion of the request it replaced', async () => {
  const h = harness()
  const loader = h.loader()
  h.load(loader, 'high/seg-42.m4s')
  h.load(loader, 'high/seg-43.m4s')
  assert.equal(requests[0].init.signal!.aborted, true)
  requests[0].complete(new Response(new Uint8Array([1])))
  requests[1].complete(new Response(new Uint8Array([2])))
  await flush()
  assert.equal(h.success.length, 1)
  assert.ok(h.success[0].url.endsWith('/seg-43.m4s'))
  h.player.destroy()
})

test('a refused request or a hostile playlist ends playback with a fixed message', async () => {
  for (const response of [
    new Response(null, { status: 403 }),
    new Response('#EXTM3U\nhttps://other.example/seg-1.m4s\n'),
  ]) {
    requests = []
    const h = harness()
    h.load(h.loader(), 'high/index.m3u8', 'text')
    requests[0].complete(response)
    await flush()
    assert.equal(h.terminal.length, 1)
    assert.equal(h.success.length, 0)
    assert.doesNotMatch(h.terminal[0], /example|abcdefghijkl/)
  }
})

test('LL-HLS keeps the server’s PART-HOLD-BACK: no segment-count live sync, no rate catch-up', () => {
  // hls.js ignores PART-HOLD-BACK whenever any liveSync* option is set
  const low = harness('ll-hls').config
  for (const key of [
    'liveSyncDuration',
    'liveSyncDurationCount',
    'liveMaxLatencyDuration',
    'liveMaxLatencyDurationCount',
  ])
    assert.equal(key in low, false, key)
  assert.equal(low.lowLatencyMode, true)
  // a rate change paused WebKit's MSE playback: the 1x default stays
  assert.equal('maxLiveSyncPlaybackRate' in low, false)
  const conventional = harness('hls').config
  assert.equal(conventional.lowLatencyMode, false)
  assert.equal(conventional.liveSyncDurationCount, 3)
  assert.equal(conventional.liveMaxLatencyDurationCount, 5)
  assert.equal('maxLiveSyncPlaybackRate' in conventional, false)
})

test('the worker only where the page serves it', () => {
  assert.equal(harness('hls').config.enableWorker, false)
  const worker = harness('hls', '/assets/hls.worker.js').config
  assert.equal(worker.enableWorker, true)
  assert.equal(worker.workerPath, '/assets/hls.worker.js')
})

test('LL-HLS skips forward to the target latency within the buffer once playback starts', () => {
  const h = harness('ll-hls')
  h.hls.liveSyncPosition = 3.3
  h.video.dispatchEvent(new Event('playing'))
  assert.equal(h.video.currentTime, 3, 'a small offset is not worth a visible skip')
  h.hls.liveSyncPosition = 4.4
  h.video.dispatchEvent(new Event('playing'))
  assert.equal(h.video.currentTime, 4.4)
  h.video.currentTime = 3
  h.hls.liveSyncPosition = 6.8
  h.video.dispatchEvent(new Event('playing'))
  assert.equal(h.video.currentTime, 3, 'never past the buffered media')
  h.hls.liveSyncPosition = 4.4
  h.player.destroy()
  h.video.dispatchEvent(new Event('playing'))
  assert.equal(h.video.currentTime, 3)
  const conventional = harness('hls')
  conventional.hls.liveSyncPosition = 4.4
  conventional.video.dispatchEvent(new Event('playing'))
  assert.equal(conventional.video.currentTime, 3)
})
