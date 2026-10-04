import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import { clearHLSVideo, hlsBase, hlsLeaseURL, hlsModes, scopedHLSURL, terminalHLSStatus,
  validHLSOffer, validHLSPlaylist } from '../src/neko/hls/protocol.js'

import { capabilities } from './hls-fixtures.mjs'
const base = 'https://neko.example/prefix/api/media/hls/'
const master = base + 'abcdefghijklmnopqrstuv/master.m3u8'

test('HLS advertising rejects incompatible codecs, duplicate modes and limits', () => {
  assert.deepEqual(hlsModes(capabilities), ['hls', 'll-hls'])
  for (const change of [{ version: 2 }, { modes: ['hls','hls'] }, { modes: ['auto'] },
    { audio_rate: 44100 }, { video_codec: 'hevc' }, { variants: [] }, { limits: {} }]) {
    assert.deepEqual(hlsModes({ ...capabilities, ...change }), [])
  }
})

test('HLS base and bootstrap keep prefixes without carrying credentials or redirects', () => {
  assert.equal(hlsBase('wss://neko.example/prefix/ws', 'https://neko.example/prefix/'), base)
  for (const url of ['ws://neko.example/ws', 'wss://other.example/ws', 'wss://u:p@neko.example/ws',
    'wss://neko.example/ws?password=secret', 'wss://neko.example/ws#secret']) {
    assert.throws(() => hlsBase(url, 'https://neko.example/'))
  }
  assert.equal(hlsLeaseURL({ mode: 'hls', master: '/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8', idle_expires_in_ms: 30000 }, 'hls', base), master)
  for (const path of ['https://other.example/master.m3u8', '/api/media/hls/short/master.m3u8',
    '/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8?ticket=secret']) {
    assert.throws(() => hlsLeaseURL({ mode: 'hls', master: path, idle_expires_in_ms: 30000 }, 'hls', base))
  }
  const offer = { version:1, backend:'hls', mode:'ll-hls', path:'/api/media/hls/session', ticket:'a'.repeat(32), expires_in_ms:10000 }
  assert.equal(validHLSOffer(offer, 'll-hls'), true)
  for (const change of [{ ticket:'a'.repeat(31) }, { expires_in_ms:10001 }, { path:'/session?ticket=secret' }, { mode:'hls' }]) {
    assert.equal(validHLSOffer({ ...offer, ...change }, 'll-hls'), false)
  }
})

test('HLS loader confines every playlist and object to its exact lease', () => {
  assert.equal(scopedHLSURL(new URL('high/part-45-0.m4s', master).href, master, 'll-hls'), true)
  assert.equal(scopedHLSURL(new URL('audio/index.m3u8?_HLS_msn=45&_HLS_part=2', master).href, master, 'll-hls'), true)
  assert.equal(scopedHLSURL(new URL('high/index.m3u8?_HLS_msn=45&_HLS_part=6', master).href, master, 'll-hls'), true)
  for (const path of ['https://other.example/high/index.m3u8', base + 'differentLease000000000/master.m3u8',
    master + '?ticket=secret', new URL('high/index.m3u8?_HLS_skip=YES', master).href,
    new URL('high/index.m3u8?_HLS_part=1', master).href,
    new URL('high/index.m3u8?_HLS_msn=1&_HLS_msn=2', master).href,
    new URL('high/index.m3u8?_HLS_msn=1&_HLS_part=7', master).href,
    new URL('high/%2e%2e/index.m3u8', master).href,
    new URL('high/init-1.mp4?x=1', master).href]) {
    assert.equal(scopedHLSURL(path, master, 'll-hls'), false, path)
  }
  assert.equal(scopedHLSURL(new URL('high/index.m3u8?_HLS_msn=1', master).href, master, 'hls'), false)
})

test('server golden playlists share the browser URL boundary, including rendition reports', async () => {
  for (const [name, path, mode] of [['master','master.m3u8','hls'], ['conventional','high/index.m3u8','hls'], ['low-latency','high/index.m3u8','ll-hls']]) {
    const fixture = (await readFile(new URL(`../../server/internal/mediahls/testdata/${name}.m3u8`, import.meta.url), 'utf8')).replaceAll('\r\n','\n')
    assert.equal(validHLSPlaylist(fixture, new URL(path, master).href, master, mode), true, name)
  }
  for (const line of ['https://other.example/seg-1.m4s', '#EXT-X-MAP:URI="https://other.example/init-1.mp4"',
    '#EXT-X-KEY:METHOD=AES-128,URI="key"', '#EXT-X-MEDIA:URI="//other.example/audio/index.m3u8"']) {
    assert.equal(validHLSPlaylist('#EXTM3U\n' + line + '\n', master, master, 'hls'), false)
  }
})

test('authorization failures are terminal and cleanup removes playable buffered media', () => {
  for (const status of [400,401,403,404,410,413]) assert.equal(terminalHLSStatus(status), true)
  for (const status of [0,200,429,503]) assert.equal(terminalHLSStatus(status), false)
  const calls = []
  const video = { srcObject: {}, pause: () => calls.push('pause'), removeAttribute: (name) => calls.push(name), load: () => calls.push('load') }
  clearHLSVideo(video)
  assert.equal(video.srcObject, null)
  assert.deepEqual(calls, ['pause','src','load'])
})
