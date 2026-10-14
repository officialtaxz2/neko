import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'
import * as protocol from '../src/neko/hls/protocol.js'

const source = await readFile(new URL('../src/neko/hls/player.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { target:ts.ScriptTarget.ES2022, module:ts.ModuleKind.CommonJS } }).outputText
const master = 'https://neko.example/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8'
const flush = async () => { await new Promise(setImmediate); await new Promise(setImmediate) }

function harness() {
  const exports = {}, requests = [], terminal = [], success = []
  let instance, player
  class MockHls {
    static Events = { ERROR:'error' }
    constructor(config) { this.config = config; this.loaders = []; instance = this }
    on() {}
    attachMedia() {}
    loadSource() {}
    destroy() { for (const loader of this.loaders) loader.destroy() }
  }
  runInNewContext(compiled, {
    exports, window:{ setTimeout, clearTimeout }, performance, URL, AbortController, TextDecoder, Uint8Array,
    require(path) {
      if (path === 'hls.js') return { default:MockHls }
      if (path === 'hls.js/dist/hls.worker.js?url') return { default:'/assets/hls.worker.js' }
      if (path === './protocol.js') return protocol
      throw new Error('Unexpected player module')
    },
    // Deliberately ignore abort in this double to exercise stale completions.
    fetch: (url, init) => new Promise((resolve) => requests.push({ url, init, complete:resolve })),
  })
  player = exports.createMSEPlayer({}, master, 'll-hls', (detail) => { terminal.push(detail); player.destroy() })
  const loader = () => {
    const value = new instance.config.loader()
    instance.loaders.push(value)
    return value
  }
  const load = (value, path = 'high/seg-42.m4s', responseType = 'arraybuffer') => value.load(
    { url:new URL(path, master).href, responseType, type:'media-fragment' },
    { loadPolicy:{ maxLoadTimeMs:12000 } },
    { onSuccess: (response) => success.push(response), onError() {}, onTimeout() {} },
  )
  return { player, requests, terminal, success, loader, load }
}

test('MSE loader leaves the fourth lease slot free and discards queued work on destroy', async () => {
  const h = harness()
  for (let i = 0; i < 4; i++) h.load(h.loader())
  assert.equal(h.requests.length, 3)
  h.player.destroy()
  assert.ok(h.requests.every(({init}) => init.signal.aborted))
  for (const request of h.requests) request.complete(new Response(new Uint8Array([1])))
  await flush()
  assert.equal(h.requests.length, 3)
  assert.equal(h.success.length, 0)
})

test('MSE loader rejects external and credential-bearing URLs before HTTP', () => {
  for (const path of ['https://other.example/seg-42.m4s', 'high/seg-42.m4s?ticket=secret']) {
    const h = harness()
    h.load(h.loader(), path)
    assert.equal(h.requests.length, 0)
    assert.equal(h.terminal.length, 1)
    assert.doesNotMatch(h.terminal[0], /secret|example|abcdefghijkl/)
  }
})

test('a reused loader never delivers the completion of its replaced request', async () => {
  const h = harness(), loader = h.loader()
  h.load(loader, 'high/seg-42.m4s')
  h.load(loader, 'high/seg-43.m4s')
  assert.equal(h.requests[0].init.signal.aborted, true)
  h.requests[0].complete(new Response(new Uint8Array([1])))
  h.requests[1].complete(new Response(new Uint8Array([2])))
  await flush()
  assert.equal(h.success.length, 1)
  assert.ok(h.success[0].url.endsWith('/seg-43.m4s'))
  h.player.destroy()
})

test('HTTP denial and hostile playlists cause a fixed terminal error', async () => {
  for (const response of [new Response(null, { status:403 }), new Response('#EXTM3U\nhttps://other.example/seg-1.m4s\n')]) {
    const h = harness()
    h.load(h.loader(), 'high/index.m3u8', 'text')
    h.requests[0].complete(response)
    await flush()
    assert.equal(h.terminal.length, 1)
    assert.equal(h.success.length, 0)
    assert.doesNotMatch(h.terminal[0], /example|abcdefghijkl/)
  }
})
