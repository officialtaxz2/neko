// The HLS contract's checks (protocol.ts), against the server's golden playlists and hostile input
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import {
  clearHLSVideo,
  hlsBase,
  hlsLeaseURL,
  hlsMasterChildren,
  hlsModes,
  scopedHLSURL,
  terminalHLSStatus,
  validHLSOffer,
  validHLSPlaylist,
} from './protocol.ts'
import { capabilities } from '../../test/hls.ts'

const base = 'https://neko.example/prefix/api/media/hls/'
const master = base + 'abcdefghijklmnopqrstuv/master.m3u8'
const golden = async (name: string) =>
  (
    await readFile(new URL(`../../../../../server/internal/mediahls/testdata/${name}.m3u8`, import.meta.url), 'utf8')
  ).replaceAll('\r\n', '\n')

test('advertising: incompatible codecs, duplicate modes and other limits advertise nothing', () => {
  assert.deepEqual(hlsModes(capabilities), ['hls', 'll-hls'])
  for (const change of [
    { version: 2 },
    { modes: ['hls', 'hls'] },
    { modes: ['auto'] },
    { audio_rate: 44100 },
    { video_codec: 'hevc' },
    { variants: [] },
    { limits: {} },
  ])
    assert.deepEqual(hlsModes({ ...capabilities, ...change }), [], JSON.stringify(change))
  assert.deepEqual(hlsModes(null), [])
})

test('the lease routes keep the path prefix and carry no credentials; plain HTTP or another origin is refused', () => {
  assert.equal(hlsBase('https://neko.example/prefix', 'https://neko.example/prefix/'), base)
  assert.equal(hlsBase('https://neko.example/prefix/', 'https://neko.example/prefix/#/ShareToken-12345'), base)
  assert.equal(hlsBase('https://neko.example', 'https://neko.example/'), 'https://neko.example/api/media/hls/')
  for (const [server, page] of [
    ['http://neko.example', 'http://neko.example/'],
    ['https://neko.example', 'http://neko.example/'],
    ['https://other.example', 'https://neko.example/'],
    ['https://u:p@neko.example', 'https://neko.example/'],
    ['https://neko.example/?password=secret', 'https://neko.example/'],
    ['https://neko.example/#secret', 'https://neko.example/'],
  ])
    assert.throws(() => hlsBase(server, page), /same-origin HTTPS/, server)
})

test('the offer and the bootstrap response: exact shapes only', () => {
  const lease = { mode: 'hls', master: '/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8', idle_expires_in_ms: 30000 }
  assert.equal(hlsLeaseURL(lease, 'hls', base), master)
  for (const path of [
    'https://other.example/master.m3u8',
    '/api/media/hls/short/master.m3u8',
    '/api/media/hls/abcdefghijklmnopqrstuv/master.m3u8?ticket=secret',
  ])
    assert.throws(() => hlsLeaseURL({ ...lease, master: path }, 'hls', base))
  assert.throws(() => hlsLeaseURL({ ...lease, mode: 'll-hls' }, 'hls', base))
  const offer = {
    version: 1,
    backend: 'hls',
    mode: 'll-hls',
    path: '/api/media/hls/session',
    ticket: 'a'.repeat(32),
    expires_in_ms: 10000,
  }
  assert.equal(validHLSOffer(offer, 'll-hls'), true)
  for (const change of [
    { ticket: 'a'.repeat(31) },
    { expires_in_ms: 10001 },
    { path: '/session?ticket=secret' },
    { mode: 'hls' },
  ])
    assert.equal(validHLSOffer({ ...offer, ...change }, 'll-hls'), false, JSON.stringify(change))
})

test('the loader is confined to its exact lease', () => {
  assert.equal(scopedHLSURL(new URL('high/part-45-0.m4s', master).href, master, 'll-hls'), true)
  assert.equal(scopedHLSURL(new URL('audio/index.m3u8?_HLS_msn=45&_HLS_part=2', master).href, master, 'll-hls'), true)
  assert.equal(scopedHLSURL(new URL('high/index.m3u8?_HLS_msn=45&_HLS_part=6', master).href, master, 'll-hls'), true)
  for (const path of [
    'https://other.example/high/index.m3u8',
    base + 'differentLease000000000/master.m3u8',
    master + '?ticket=secret',
    new URL('high/index.m3u8?_HLS_skip=YES', master).href,
    new URL('high/index.m3u8?_HLS_part=1', master).href,
    new URL('high/index.m3u8?_HLS_msn=1&_HLS_msn=2', master).href,
    new URL('high/index.m3u8?_HLS_msn=1&_HLS_part=7', master).href,
    new URL('high/%2e%2e/index.m3u8', master).href,
    new URL('high/init-1.mp4?x=1', master).href,
  ])
    assert.equal(scopedHLSURL(path, master, 'll-hls'), false, path)
  assert.equal(scopedHLSURL(new URL('high/index.m3u8?_HLS_msn=1', master).href, master, 'hls'), false)
})

test('the server’s golden playlists pass the same boundary, rendition reports included', async () => {
  for (const [name, path, mode] of [
    ['master', 'master.m3u8', 'hls'],
    ['conventional', 'high/index.m3u8', 'hls'],
    ['low-latency', 'high/index.m3u8', 'll-hls'],
  ] as const)
    assert.equal(validHLSPlaylist(await golden(name), new URL(path, master).href, master, mode), true, name)
  const children = hlsMasterChildren(await golden('master'), master, 'hls')
  assert.ok(children.some((c) => c.endsWith('/audio/index.m3u8')))
  for (const line of [
    'https://other.example/seg-1.m4s',
    '#EXT-X-MAP:URI="https://other.example/init-1.mp4"',
    '#EXT-X-KEY:METHOD=AES-128,URI="key"',
    '#EXT-X-MEDIA:URI="//other.example/audio/index.m3u8"',
  ])
    assert.equal(validHLSPlaylist('#EXTM3U\n' + line + '\n', master, master, 'hls'), false, line)
})

test('authorization failures are terminal; cleanup drops the source and the buffered media', () => {
  for (const status of [400, 401, 403, 404, 410, 413]) assert.equal(terminalHLSStatus(status), true)
  for (const status of [0, 200, 429, 503]) assert.equal(terminalHLSStatus(status), false)
  const calls: string[] = []
  const video = {
    srcObject: {},
    pause: () => calls.push('pause'),
    removeAttribute: (name: string) => calls.push(name),
    load: () => calls.push('load'),
  }
  clearHLSVideo(video as unknown as HTMLVideoElement)
  assert.equal(video.srcObject, null)
  assert.deepEqual(calls, ['pause', 'src', 'load'])
})
