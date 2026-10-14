// This module deliberately uses JavaScript-compatible TypeScript so the
// Node 18 fixture test can execute the exact browser parser source.

export const MEDIA_PROTOCOL = 'neko.media.v1'
export const MEDIA_BACKEND = 'webcodecs-ws'
export const MEDIA_HEADER_LENGTH = 64
export const MAX_SAFE_MEDIA_INTEGER = Number.MAX_SAFE_INTEGER

export const RECORD = Object.freeze({
  FORMAT: 1,
  UNIT: 2,
  DISCONTINUITY: 3,
  END: 4,
})

export const KIND = Object.freeze({
  NONE: 0,
  AUDIO: 1,
  VIDEO: 2,
})

export const FLAG = Object.freeze({
  KEYFRAME: 1 << 0,
  PTS_VALID: 1 << 1,
  DTS_VALID: 1 << 2,
  CONFIG_PRESENT: 1 << 3,
})

const KNOWN_FLAGS = FLAG.KEYFRAME | FLAG.PTS_VALID | FLAG.DTS_VALID | FLAG.CONFIG_PRESENT
const MAX_METADATA = 4 * 1024
const MAX_AUDIO_PAYLOAD = 64 * 1024
const MAX_VIDEO_PAYLOAD = 8 * 1024 * 1024
const MAX_DURATION_US = 10_000_000
const MAX_SOURCE_ID_BYTES = 256
const MAX_VIDEO_DIMENSION = 16_383
const FORMAT_FIELDS = [
  'schema',
  'source_id',
  'source_generation',
  'codec',
  'mime_type',
  'clock_rate',
  'channels',
  'coded_width',
  'coded_height',
  'display_width',
  'display_height',
  'frame_rate_numerator',
  'frame_rate_denominator',
  'nominal_bitrate',
]

const DISCONTINUITY_REASONS = new Set([
  'source_switch',
  'source_restart',
  'format_change',
  'timestamp_reset',
  'server_overflow',
  'browser_resync',
  'resumed',
])
const END_REASONS = new Set(['normal', 'revoked', 'replaced', 'backend_error', 'shutdown'])
const textDecoder = new TextDecoder('utf-8', { fatal: true })
const textEncoder = new TextEncoder()

function invalid(reason) {
  return new Error(`invalid neko.media.v1 record: ${reason}`)
}

function safeUint64(view, offset, name) {
  const value = view.getBigUint64(offset, false)
  if (value > BigInt(MAX_SAFE_MEDIA_INTEGER)) throw invalid(`${name} exceeds JavaScript safe integer range`)
  return Number(value)
}

function safeInt64(view, offset, name) {
  const value = view.getBigInt64(offset, false)
  if (value < 0n || value > BigInt(MAX_SAFE_MEDIA_INTEGER)) throw invalid(`${name} is outside the accepted range`)
  return Number(value)
}

function assertSafeUnsigned(value, name, { positive = false, max = MAX_SAFE_MEDIA_INTEGER } = {}) {
  if (!Number.isSafeInteger(value) || value < (positive ? 1 : 0) || value > max) {
    throw invalid(`${name} is outside the accepted range`)
  }
}

function assertExactKeys(value, expected, name) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw invalid(`${name} must be an object`)
  const actual = Object.keys(value).sort()
  const wanted = [...expected].sort()
  if (actual.length !== wanted.length || actual.some((key, index) => key !== wanted[index])) {
    throw invalid(`${name} fields`)
  }
}

// JSON.parse accepts duplicate object keys. This small grammar walk rejects
// them (including in nested objects) before JSON.parse constructs a value.
function rejectDuplicateJSONKeys(source) {
  let index = 0

  const whitespace = () => {
    while (index < source.length && /[\t\n\r ]/.test(source[index])) index++
  }

  const stringToken = () => {
    if (source[index] !== '"') throw invalid('JSON object key')
    const start = index++
    while (index < source.length) {
      const code = source.charCodeAt(index)
      if (code === 0x22) {
        index++
        try {
          return JSON.parse(source.slice(start, index))
        } catch (_) {
          throw invalid('JSON string')
        }
      }
      if (code < 0x20) throw invalid('JSON control character')
      if (code === 0x5c) {
        index++
        if (index >= source.length) throw invalid('JSON escape')
        if (source[index] === 'u') {
          const unicode = source.slice(index + 1, index + 5)
          if (!/^[0-9a-fA-F]{4}$/.test(unicode)) throw invalid('JSON unicode escape')
          index += 5
          continue
        }
        if (!/["\\/bfnrt]/.test(source[index])) throw invalid('JSON escape')
      }
      index++
    }
    throw invalid('unterminated JSON string')
  }

  const literal = () => {
    const rest = source.slice(index)
    const match = /^(?:true|false|null|-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?)/.exec(rest)
    if (!match) throw invalid('JSON value')
    index += match[0].length
  }

  const value = () => {
    whitespace()
    if (source[index] === '{') return object()
    if (source[index] === '[') return array()
    if (source[index] === '"') {
      stringToken()
      return
    }
    literal()
  }

  const object = () => {
    index++
    whitespace()
    const keys = new Set()
    if (source[index] === '}') {
      index++
      return
    }
    for (;;) {
      whitespace()
      const key = stringToken()
      if (keys.has(key)) throw invalid('duplicate JSON key')
      keys.add(key)
      whitespace()
      if (source[index++] !== ':') throw invalid('JSON object separator')
      value()
      whitespace()
      const delimiter = source[index++]
      if (delimiter === '}') return
      if (delimiter !== ',') throw invalid('JSON object delimiter')
    }
  }

  const array = () => {
    index++
    whitespace()
    if (source[index] === ']') {
      index++
      return
    }
    for (;;) {
      value()
      whitespace()
      const delimiter = source[index++]
      if (delimiter === ']') return
      if (delimiter !== ',') throw invalid('JSON array delimiter')
    }
  }

  value()
  whitespace()
  if (index !== source.length) throw invalid('trailing JSON data')
}

function strictJSON(bytes, expected, name) {
  if (bytes.byteLength === 0 || bytes.byteLength > MAX_METADATA) throw invalid(`${name} length`)
  let source
  try {
    source = textDecoder.decode(bytes)
  } catch (_) {
    throw invalid(`${name} UTF-8`)
  }
  rejectDuplicateJSONKeys(source)
  let value
  try {
    value = JSON.parse(source)
  } catch (_) {
    throw invalid(`${name} JSON`)
  }
  assertExactKeys(value, expected, name)
  return value
}

function validateTrack(kind, trackID) {
  if ((kind === KIND.AUDIO && trackID === 1) || (kind === KIND.VIDEO && trackID === 2)) return
  throw invalid('kind/track mismatch')
}

function validateFormat(kind, metadata) {
  if (
    metadata.schema !== 'neko.media.format/1' ||
    typeof metadata.source_id !== 'string' ||
    metadata.source_id.length === 0 ||
    textEncoder.encode(metadata.source_id).byteLength > MAX_SOURCE_ID_BYTES ||
    typeof metadata.codec !== 'string' ||
    typeof metadata.mime_type !== 'string'
  ) {
    throw invalid('format metadata')
  }
  assertSafeUnsigned(metadata.source_generation, 'source_generation', { positive: true })
  assertSafeUnsigned(metadata.clock_rate, 'clock_rate')
  assertSafeUnsigned(metadata.channels, 'channels')
  assertSafeUnsigned(metadata.coded_width, 'coded_width', { max: MAX_VIDEO_DIMENSION })
  assertSafeUnsigned(metadata.coded_height, 'coded_height', { max: MAX_VIDEO_DIMENSION })
  assertSafeUnsigned(metadata.display_width, 'display_width', { max: MAX_VIDEO_DIMENSION })
  assertSafeUnsigned(metadata.display_height, 'display_height', { max: MAX_VIDEO_DIMENSION })
  assertSafeUnsigned(metadata.frame_rate_numerator, 'frame_rate_numerator')
  assertSafeUnsigned(metadata.frame_rate_denominator, 'frame_rate_denominator')
  assertSafeUnsigned(metadata.nominal_bitrate, 'nominal_bitrate')

  if (kind === KIND.VIDEO) {
    if (
      metadata.codec !== 'vp8' ||
      metadata.mime_type !== 'video/VP8' ||
      metadata.clock_rate !== 90000 ||
      metadata.channels !== 0 ||
      metadata.coded_width === 0 ||
      metadata.coded_height === 0 ||
      metadata.display_width === 0 ||
      metadata.display_height === 0 ||
      metadata.frame_rate_numerator === 0 ||
      metadata.frame_rate_denominator === 0
    ) {
      throw invalid('video format metadata')
    }
    return
  }

  if (
    metadata.codec !== 'opus' ||
    metadata.mime_type !== 'audio/opus' ||
    metadata.clock_rate !== 48000 ||
    metadata.channels !== 2 ||
    metadata.coded_width !== 0 ||
    metadata.coded_height !== 0 ||
    metadata.display_width !== 0 ||
    metadata.display_height !== 0 ||
    metadata.frame_rate_numerator !== 0 ||
    metadata.frame_rate_denominator !== 0
  ) {
    throw invalid('audio format metadata')
  }
}

export function parseMediaRecord(data) {
  if (!(data instanceof ArrayBuffer)) throw invalid('message is not an ArrayBuffer')
  if (data.byteLength < MEDIA_HEADER_LENGTH) throw invalid('truncated header')

  const bytes = new Uint8Array(data)
  const view = new DataView(data)
  if (bytes[0] !== 0x4e || bytes[1] !== 0x45 || bytes[2] !== 0x4b || bytes[3] !== 0x4f) throw invalid('magic')
  if (bytes[4] !== 1) throw invalid('version')
  const type = bytes[5]
  const kind = bytes[6]
  const flags = bytes[7]
  if (view.getUint16(8, false) !== MEDIA_HEADER_LENGTH) throw invalid('header length')
  if (view.getUint16(10, false) !== 0) throw invalid('reserved')
  if ((flags & ~KNOWN_FLAGS) !== 0) throw invalid('unknown flags')

  const metadataLength = view.getUint32(12, false)
  const payloadLength = view.getUint32(16, false)
  const total = MEDIA_HEADER_LENGTH + metadataLength + payloadLength
  if (!Number.isSafeInteger(total) || total !== data.byteLength) throw invalid('record length')
  if (metadataLength > MAX_METADATA) throw invalid('metadata too large')

  if (type === RECORD.FORMAT) {
    if ((kind !== KIND.AUDIO && kind !== KIND.VIDEO) || metadataLength === 0 || payloadLength !== 0) {
      throw invalid('format lengths')
    }
  } else if (type === RECORD.UNIT) {
    if (metadataLength !== 0 || payloadLength === 0) throw invalid('unit lengths')
    if (kind === KIND.AUDIO && payloadLength > MAX_AUDIO_PAYLOAD) throw invalid('audio payload too large')
    if (kind === KIND.VIDEO && payloadLength > MAX_VIDEO_PAYLOAD) throw invalid('video payload too large')
    if (kind !== KIND.AUDIO && kind !== KIND.VIDEO) throw invalid('unit kind')
  } else if (type === RECORD.DISCONTINUITY || type === RECORD.END) {
    if (metadataLength === 0 || payloadLength !== 0) throw invalid('lifecycle lengths')
  } else {
    throw invalid('record type')
  }

  const trackID = view.getUint32(20, false)
  const generation = safeUint64(view, 24, 'generation')
  const sequence = safeUint64(view, 32, 'sequence')
  const pts = safeInt64(view, 40, 'PTS')
  const dts = safeInt64(view, 48, 'DTS')
  const duration = safeUint64(view, 56, 'duration')
  const metadataBytes = bytes.subarray(MEDIA_HEADER_LENGTH, MEDIA_HEADER_LENGTH + metadataLength)
  const payload = bytes.subarray(MEDIA_HEADER_LENGTH + metadataLength)
  let metadata = null

  if (type === RECORD.FORMAT) {
    validateTrack(kind, trackID)
    if (flags !== 0 || generation === 0 || sequence !== 0 || pts !== 0 || dts !== 0 || duration !== 0) {
      throw invalid('format lifecycle fields')
    }
    metadata = strictJSON(metadataBytes, FORMAT_FIELDS, 'format metadata')
    validateFormat(kind, metadata)
  } else if (type === RECORD.UNIT) {
    validateTrack(kind, trackID)
    if (generation === 0 || (flags & FLAG.CONFIG_PRESENT) !== 0) throw invalid('unit fields')
    if (kind === KIND.AUDIO && (flags & FLAG.KEYFRAME) !== 0) throw invalid('audio keyframe flag')
    if ((flags & FLAG.DTS_VALID) === 0 && dts !== 0) throw invalid('invalid DTS must be zero')
    if (duration === 0 || duration > MAX_DURATION_US) throw invalid('duration')
    if (kind === KIND.VIDEO) {
      const payloadSaysKeyframe = (payload[0] & 0x01) === 0
      const envelopeSaysKeyframe = (flags & FLAG.KEYFRAME) !== 0
      if (payloadSaysKeyframe !== envelopeSaysKeyframe) throw invalid('VP8 keyframe marker mismatch')
    }
  } else if (type === RECORD.DISCONTINUITY) {
    validateTrack(kind, trackID)
    if (flags !== 0 || generation === 0 || sequence !== 0 || pts !== 0 || dts !== 0 || duration !== 0) {
      throw invalid('discontinuity fields')
    }
    metadata = strictJSON(metadataBytes, ['schema', 'reason'], 'discontinuity metadata')
    if (metadata.schema !== 'neko.media.discontinuity/1' || !DISCONTINUITY_REASONS.has(metadata.reason)) {
      throw invalid('discontinuity metadata')
    }
  } else {
    if (kind !== KIND.NONE || trackID !== 0 || flags !== 0 || generation !== 0 || sequence !== 0 || pts !== 0 || dts !== 0 || duration !== 0) {
      throw invalid('end fields')
    }
    metadata = strictJSON(metadataBytes, ['schema', 'reason'], 'end metadata')
    if (metadata.schema !== 'neko.media.end/1' || !END_REASONS.has(metadata.reason)) throw invalid('end metadata')
  }

  return { type, kind, flags, trackID, generation, sequence, pts, dts, duration, metadata, payload }
}
