import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const parserURL = new URL('../src/neko/media/protocol.ts', import.meta.url)
const parserSource = await readFile(parserURL, 'utf8')
const parser = await import(`data:text/javascript;base64,${Buffer.from(parserSource).toString('base64')}`)
const fixture = JSON.parse(
  await readFile(new URL('../../server/internal/mediaws/testdata/neko_media_v1_golden.json', import.meta.url), 'utf8'),
)

const fromHex = (hex) => Uint8Array.from(Buffer.from(hex, 'hex')).buffer

test('browser parser accepts every shared version-1 golden record', () => {
  assert.equal(fixture.schema, 'neko.media.fixture/1')
  const parsed = new Map(fixture.records.map(({ name, hex }) => [name, parser.parseMediaRecord(fromHex(hex))]))

  assert.equal(parsed.get('format_video').metadata.codec, 'vp8')
  assert.equal(parsed.get('format_audio').metadata.codec, 'opus')
  assert.equal(parsed.get('unit_video_key').flags & parser.FLAG.KEYFRAME, parser.FLAG.KEYFRAME)
  assert.equal(parsed.get('unit_audio').duration, 20_000)
  assert.equal(parsed.get('discontinuity_video').metadata.reason, 'source_restart')
  assert.equal(parsed.get('end').metadata.reason, 'normal')
})

test('browser parser rejects framing, lifecycle and VP8 marker mismatches', () => {
  const unitHex = fixture.records.find(({ name }) => name === 'unit_video_key').hex
  const badMagic = new Uint8Array(fromHex(unitHex))
  badMagic[0] = 0
  assert.throws(() => parser.parseMediaRecord(badMagic.buffer), /magic/)

  const trailing = new Uint8Array(badMagic.byteLength + 1)
  trailing.set(new Uint8Array(fromHex(unitHex)))
  assert.throws(() => parser.parseMediaRecord(trailing.buffer), /record length/)

  const wrongTrack = new Uint8Array(fromHex(unitHex))
  new DataView(wrongTrack.buffer).setUint32(20, 1, false)
  assert.throws(() => parser.parseMediaRecord(wrongTrack.buffer), /kind\/track/)

  const wrongMarker = new Uint8Array(fromHex(unitHex))
  wrongMarker[parser.MEDIA_HEADER_LENGTH] |= 1
  assert.throws(() => parser.parseMediaRecord(wrongMarker.buffer), /VP8 keyframe marker/)
})

test('browser parser rejects duplicate and unknown metadata fields', () => {
  const formatHex = fixture.records.find(({ name }) => name === 'format_video').hex
  const original = new Uint8Array(fromHex(formatHex))
  const metadata = new TextDecoder().decode(original.subarray(parser.MEDIA_HEADER_LENGTH))

  const buildFormat = (replacement) => {
    const encoded = new TextEncoder().encode(replacement)
    const output = new Uint8Array(parser.MEDIA_HEADER_LENGTH + encoded.byteLength)
    output.set(original.subarray(0, parser.MEDIA_HEADER_LENGTH))
    new DataView(output.buffer).setUint32(12, encoded.byteLength, false)
    output.set(encoded, parser.MEDIA_HEADER_LENGTH)
    return output.buffer
  }

  assert.throws(
    () => parser.parseMediaRecord(buildFormat(metadata.replace('{', '{"schema":"duplicate",'))),
    /duplicate JSON key/,
  )
  assert.throws(
    () => parser.parseMediaRecord(buildFormat(metadata.replace(/}$/, ',"unexpected":true}'))),
    /format metadata fields/,
  )
})
