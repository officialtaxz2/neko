import { FLAG, KIND, MEDIA_BACKEND, MEDIA_PROTOCOL, RECORD, parseMediaRecord } from './protocol'
import { isMediaRetryCloseCode, shouldAwaitMediaCloseForEnd, shouldDropDecodedVideoOutput } from './recovery.js'

type MediaKindName = 'audio' | 'video'
type ResyncReason =
  | 'queue_overflow'
  | 'video_compressed_overflow'
  | 'audio_compressed_overflow'
  | 'audio_output_overflow'
  | 'audio_worklet_overflow'
  | 'decoder_error'
  | 'timestamp'
  | 'audio_underflow'
  | 'av_skew'

interface TrackState {
  name: MediaKindName
  kind: number
  requested: boolean
  generation: number
  lastSequence: number
  lastPTS: number
  awaitingKeyframe: boolean
  configured: boolean
  format: any
  compressed: any[]
  draining: boolean
  received: number
  decoded: number
  rendered: number
  drops: number
  outstandingVideo: number
  bufferedAudioMS: number
  pendingAudioMS: number
  decoder?: VideoDecoder | AudioDecoder
}

const scope = self as any
const VIDEO_COMPRESSED_CAP = 4
const AUDIO_COMPRESSED_CAP = 16
const VIDEO_DECODE_CAP = 4
const AUDIO_DECODE_CAP = 16
const VIDEO_RENDER_CAP = 2
const AUDIO_BUFFER_CAP_MS = 200
const FEEDBACK_INTERVAL_MS = 1100

const makeTrack = (name: MediaKindName, kind: number): TrackState => ({
  name,
  kind,
  requested: name === 'video',
  generation: 0,
  lastSequence: -1,
  lastPTS: -1,
  awaitingKeyframe: name === 'video',
  configured: false,
  format: undefined,
  compressed: [],
  draining: false,
  received: 0,
  decoded: 0,
  rendered: 0,
  drops: 0,
  outstandingVideo: 0,
  bufferedAudioMS: 0,
  pendingAudioMS: 0,
})

const audio = makeTrack('audio', KIND.AUDIO)
const video = makeTrack('video', KIND.VIDEO)
let socket: WebSocket | undefined
let intentionalClose = false
let readySent = false
let formatTimer: number | undefined
let feedbackTimer: number | undefined
let resyncPending = false
let avSkewMS = 0

const post = (message: any, transfer: Transferable[] = []) => scope.postMessage(message, transfer)
const currentTrack = (kind: number) => (kind === KIND.AUDIO ? audio : kind === KIND.VIDEO ? video : undefined)

function resetTrack(track: TrackState, clearGeneration = false) {
  track.compressed = []
  track.draining = false
  track.configured = false
  track.format = undefined
  track.lastSequence = -1
  track.lastPTS = -1
  track.awaitingKeyframe = track.name === 'video'
  track.outstandingVideo = 0
  track.bufferedAudioMS = 0
  track.pendingAudioMS = 0
  track.received = 0
  track.decoded = 0
  track.rendered = 0
  track.drops = 0
  if (clearGeneration) track.generation = 0
  if (track.decoder) {
    try {
      track.decoder.reset()
    } catch (_) {}
    try {
      track.decoder.close()
    } catch (_) {}
    track.decoder = undefined
  }
}

function resetRuntime(clearGeneration = true) {
  resetTrack(audio, clearGeneration)
  resetTrack(video, clearGeneration)
  readySent = false
  resyncPending = false
  avSkewMS = 0
  if (formatTimer !== undefined) scope.clearTimeout(formatTimer)
  if (feedbackTimer !== undefined) scope.clearTimeout(feedbackTimer)
  formatTimer = undefined
  feedbackTimer = undefined
}

function sendControl(record: any) {
  if (socket?.readyState !== WebSocket.OPEN) return
  const encoded = JSON.stringify(record)
  if (new TextEncoder().encode(encoded).byteLength > 4096) {
    terminal('client control record exceeded 4 KiB')
    return
  }
  socket.send(encoded)
}

function closeSocket(code = 1000, reason = 'client_stop') {
  if (!socket) return
  const current = socket
  socket = undefined
  current.onopen = null
  current.onmessage = null
  current.onerror = null
  current.onclose = null
  try {
    if (current.readyState === WebSocket.OPEN) current.close(code, reason)
    else current.close()
  } catch (_) {}
}

function terminal(detail: string) {
  post({ type: 'terminal', detail })
  intentionalClose = true
  closeSocket(1002, 'protocol_error')
  resetRuntime()
}

function trackFeedback(track: TrackState) {
  if (!track.requested) return null
  return {
    generation: String(Math.max(1, track.generation)),
    received: String(track.received),
    decoded: String(track.decoded),
    rendered: String(track.rendered),
    compressed_queue: track.compressed.length,
    decode_queue: Math.min(
      track.name === 'video' ? VIDEO_DECODE_CAP : AUDIO_DECODE_CAP,
      Number((track.decoder as any)?.decodeQueueSize || 0),
    ),
    buffered_ms: track.name === 'audio' ? Math.min(AUDIO_BUFFER_CAP_MS, Math.max(0, Math.round(track.bufferedAudioMS))) : 0,
    drops: String(track.drops),
  }
}

function sendFeedback() {
  if (!readySent || resyncPending || !video.configured || (audio.requested && !audio.configured)) return
  sendControl({
    type: 'feedback',
    audio: trackFeedback(audio),
    video: trackFeedback(video),
    av_skew_ms: Math.max(-10_000, Math.min(10_000, Math.round(avSkewMS))),
  })
}

function scheduleFeedback() {
  if (feedbackTimer !== undefined || socket?.readyState !== WebSocket.OPEN) return
  feedbackTimer = scope.setTimeout(() => {
    feedbackTimer = undefined
    sendFeedback()
    scheduleFeedback()
  }, FEEDBACK_INTERVAL_MS)
}

function requestResync(reason: ResyncReason) {
  if (resyncPending || socket?.readyState !== WebSocket.OPEN) return
  resyncPending = true
  const generation = Math.max(video.generation, 1)
  resetTrack(audio)
  resetTrack(video)
  post({ type: 'clock-reset', kind: 'all' })
  sendControl({ type: 'resync', kind: 'all', generation: String(generation), reason })
}

async function probeCapabilities(capabilities: any, allowAudio: boolean) {
  try {
    if (
      capabilities?.version !== 1 ||
      capabilities?.backend !== MEDIA_BACKEND ||
      capabilities?.protocol !== MEDIA_PROTOCOL ||
      !Array.isArray(capabilities.video) ||
      !Array.isArray(capabilities.audio)
    ) {
      throw new Error('server returned incompatible media capabilities')
    }
    if (typeof VideoDecoder === 'undefined') throw new Error('VideoDecoder is unavailable')

    let selectedVideo: any
    for (const source of capabilities.video) {
      if (source?.codec !== 'vp8' || source?.mime_type !== 'video/VP8' || typeof source.id !== 'string') continue
      const support = await VideoDecoder.isConfigSupported({
        codec: 'vp8',
        codedWidth: source.coded_width || 640,
        codedHeight: source.coded_height || 360,
        optimizeForLatency: true,
      })
      if (support.supported) {
        selectedVideo = { source_id: source.id, codec: 'vp8' }
        break
      }
    }
    if (!selectedVideo) throw new Error('no advertised VP8 source is supported by this browser')

    let selectedAudio: any = null
    if (allowAudio && typeof AudioDecoder !== 'undefined') {
      for (const source of capabilities.audio) {
        if (source?.codec !== 'opus' || source?.mime_type !== 'audio/opus' || typeof source.id !== 'string') continue
        const support = await AudioDecoder.isConfigSupported({ codec: 'opus', sampleRate: 48000, numberOfChannels: 2 })
        if (support.supported) {
          selectedAudio = { source_id: source.id, codec: 'opus' }
          break
        }
      }
    }

    audio.requested = selectedAudio !== null
    video.requested = true
    post({ type: 'capability-choice', audio: selectedAudio, video: selectedVideo })
  } catch (error) {
    post({ type: 'terminal', detail: error instanceof Error ? error.message : String(error) })
  }
}

async function configureTrack(track: TrackState, record: any) {
  const metadata = record.metadata
  resetTrack(track)
  track.generation = record.generation
  track.format = metadata
  const generation = record.generation

  if (track.name === 'video') {
    const config: VideoDecoderConfig = {
      codec: 'vp8',
      codedWidth: metadata.coded_width,
      codedHeight: metadata.coded_height,
      optimizeForLatency: true,
    }
    const support = await VideoDecoder.isConfigSupported(config)
    if (track.generation !== generation || track.format !== metadata) return
    if (!support.supported) throw new Error('exact VP8 FORMAT is unsupported')
    const decoder = new VideoDecoder({
      output: (frame) => onVideoOutput(track, frame),
      error: () => requestResync('decoder_error'),
    })
    decoder.ondequeue = () => drain(track)
    decoder.configure(config)
    track.decoder = decoder
    track.configured = true
    post({
      type: 'video-format',
      generation: track.generation,
      codedWidth: metadata.coded_width,
      codedHeight: metadata.coded_height,
      displayWidth: metadata.display_width,
      displayHeight: metadata.display_height,
      frameRateNumerator: metadata.frame_rate_numerator,
      frameRateDenominator: metadata.frame_rate_denominator,
    })
  } else {
    const config: AudioDecoderConfig = { codec: 'opus', sampleRate: 48000, numberOfChannels: 2 }
    const support = await AudioDecoder.isConfigSupported(config)
    if (track.generation !== generation || track.format !== metadata) return
    if (!support.supported) throw new Error('exact Opus FORMAT is unsupported')
    const decoder = new AudioDecoder({
      output: (data) => onAudioOutput(track, data),
      error: () => requestResync('decoder_error'),
    })
    decoder.ondequeue = () => drain(track)
    decoder.configure(config)
    track.decoder = decoder
    track.configured = true
  }

  post({ type: 'clock-reset', kind: track.name === 'video' && audio.requested ? 'video' : 'all' })
  maybeReady()
}

function allRequestedTracksConfigured() {
  return video.configured && (!audio.requested || audio.configured)
}

function maybeReady() {
  if (!allRequestedTracksConfigured()) return
  resyncPending = false
  if (!readySent) {
    readySent = true
    if (formatTimer !== undefined) scope.clearTimeout(formatTimer)
    formatTimer = undefined
    sendControl({
      type: 'ready',
      version: 1,
      audio: audio.requested
        ? { generation: String(audio.generation), codec: 'opus', sample_rate: 48000, channels: 2 }
        : null,
      video: {
        generation: String(video.generation),
        codec: 'vp8',
        coded_width: video.format.coded_width,
        coded_height: video.format.coded_height,
      },
    })
    post({ type: 'ready', audioEnabled: audio.requested })
  }
  drain(audio)
  drain(video)
}

function onVideoOutput(track: TrackState, frame: VideoFrame) {
  if (!track.configured || track.generation === 0) {
    frame.close()
    return
  }
  track.decoded++
  if (shouldDropDecodedVideoOutput(track.outstandingVideo, VIDEO_RENDER_CAP)) {
    track.drops++
    // The decoder has consumed this VP8 chunk, so discarding only its decoded
    // output preserves reference continuity. Keep at most two transferred
    // frames without turning normal renderer throttling into a server resync.
    if (track.rendered < track.decoded) track.rendered++
    frame.close()
    return
  }
  track.outstandingVideo++
  post(
    {
      type: 'video-frame',
      generation: track.generation,
      timestamp: frame.timestamp,
      duration: frame.duration || 0,
      frame,
    },
    [frame],
  )
}

function onAudioOutput(track: TrackState, data: AudioData) {
  if (!track.configured || track.generation === 0) {
    data.close()
    return
  }
  track.decoded++
  const durationUS = data.duration || 0
  const durationMS = durationUS / 1000
  track.pendingAudioMS = Math.max(0, track.pendingAudioMS - durationMS)
  if (
    data.sampleRate !== 48000 ||
    data.numberOfChannels !== 2 ||
    durationUS <= 0 ||
    durationMS > AUDIO_BUFFER_CAP_MS ||
    track.bufferedAudioMS + durationMS > AUDIO_BUFFER_CAP_MS
  ) {
    track.drops++
    data.close()
    requestResync('audio_output_overflow')
    return
  }

  try {
    const planes: ArrayBuffer[] = []
    for (let channel = 0; channel < data.numberOfChannels; channel++) {
      const size = data.allocationSize({ planeIndex: channel, format: 'f32-planar' })
      const plane = new ArrayBuffer(size)
      data.copyTo(plane, { planeIndex: channel, format: 'f32-planar' })
      planes.push(plane)
    }
    track.bufferedAudioMS += durationMS
    post(
      {
        type: 'audio-data',
        generation: track.generation,
        timestamp: data.timestamp,
        duration: durationUS,
        numberOfFrames: data.numberOfFrames,
        numberOfChannels: data.numberOfChannels,
        sampleRate: data.sampleRate,
        planes,
      },
      planes,
    )
  } catch (_) {
    track.drops++
    requestResync('decoder_error')
  } finally {
    data.close()
  }
}

function enqueueUnit(track: TrackState, record: any) {
  if (record.generation < track.generation) {
    track.drops++
    return
  }
  if (record.generation > track.generation || !track.format) {
    requestResync('timestamp')
    return
  }
  if (record.sequence !== track.lastSequence + 1) {
    requestResync('timestamp')
    return
  }
  if (track.lastPTS >= 0 && record.pts < track.lastPTS) {
    requestResync('timestamp')
    return
  }
  if (track.name === 'video' && track.awaitingKeyframe && (record.flags & FLAG.KEYFRAME) === 0) {
    track.drops++
    return
  }

  const cap = track.name === 'video' ? VIDEO_COMPRESSED_CAP : AUDIO_COMPRESSED_CAP
  if (track.compressed.length >= cap) {
    track.drops++
    requestResync(track.name === 'video' ? 'video_compressed_overflow' : 'audio_compressed_overflow')
    return
  }
  track.lastSequence = record.sequence
  track.lastPTS = record.pts
  track.received++
  if (track.name === 'video' && (record.flags & FLAG.KEYFRAME) !== 0) track.awaitingKeyframe = false
  track.compressed.push(record)
  drain(track)
}

function drain(track: TrackState) {
  if (track.draining || resyncPending || !allRequestedTracksConfigured()) return
  track.draining = true
  queueMicrotask(() => {
    try {
      while (track.compressed.length > 0 && track.configured && track.decoder) {
        const decodeCap = track.name === 'video' ? VIDEO_DECODE_CAP : AUDIO_DECODE_CAP
        if (track.decoder.decodeQueueSize >= decodeCap) return
        const record = track.compressed[0]
        if (
          track.name === 'audio' &&
          track.bufferedAudioMS + track.pendingAudioMS + record.duration / 1000 > AUDIO_BUFFER_CAP_MS
        ) {
          return
        }
        track.compressed.shift()
        if (track.name === 'video') {
          ;(track.decoder as VideoDecoder).decode(
            new EncodedVideoChunk({
              type: (record.flags & FLAG.KEYFRAME) !== 0 ? 'key' : 'delta',
              timestamp: record.pts,
              duration: record.duration,
              data: record.payload,
            }),
          )
        } else {
          track.pendingAudioMS += record.duration / 1000
          ;(track.decoder as AudioDecoder).decode(
            new EncodedAudioChunk({ type: 'key', timestamp: record.pts, duration: record.duration, data: record.payload }),
          )
        }
      }
    } catch (_) {
      requestResync('decoder_error')
    } finally {
      track.draining = false
    }
  })
}

function handleRecord(data: ArrayBuffer) {
  let record
  try {
    record = parseMediaRecord(data)
  } catch (error) {
    terminal(error instanceof Error ? error.message : String(error))
    return
  }

  if (record.type === RECORD.END) {
    if (shouldAwaitMediaCloseForEnd(record.metadata.reason)) {
      // A backend_error END deliberately precedes the authoritative private
      // close code. Keep the close handler attached so only 4413 or 4500 can
      // enter the bounded same-backend retry policy.
      resetRuntime()
      return
    }
    post({ type: 'end', reason: record.metadata.reason })
    intentionalClose = true
    closeSocket(1000, 'end')
    resetRuntime()
    return
  }

  const track = currentTrack(record.kind)
  if (!track || !track.requested) {
    terminal('server sent an unrequested media kind')
    return
  }

  if (record.type === RECORD.DISCONTINUITY) {
    if (record.generation < track.generation) return
    resetTrack(track)
    track.generation = record.generation
    resyncPending = true
    // A video-only server generation retains the normalized audio timeline;
    // clear stale canvas frames without discarding valid queued PCM. Common
    // server resyncs emit audio first, which selects the full reset path.
    post({ type: 'clock-reset', kind: track.name === 'video' && audio.requested ? 'video' : 'all' })
    return
  }

  if (record.type === RECORD.FORMAT) {
    if (record.generation < track.generation) return
    if (
      track.name === 'video' &&
      track.generation > 0 &&
      record.generation > track.generation &&
      !resyncPending &&
      audio.requested
    ) {
      track.generation = record.generation
      requestResync('timestamp')
      return
    }
    configureTrack(track, record).catch((error) => terminal(error instanceof Error ? error.message : String(error)))
    return
  }

  enqueueUnit(track, record)
}

function startMedia(url: string, ticket: string, requestedAudio: boolean) {
  intentionalClose = true
  closeSocket()
  resetRuntime()
  intentionalClose = false
  audio.requested = requestedAudio
  video.requested = true

  try {
    const candidate = new WebSocket(url, [MEDIA_PROTOCOL, `neko.media.ticket.${ticket}`])
    socket = candidate
    candidate.binaryType = 'arraybuffer'
    candidate.onopen = () => {
      if (socket !== candidate) return
      if (candidate.protocol !== MEDIA_PROTOCOL) {
        terminal('server selected an unexpected media subprotocol')
        return
      }
      post({ type: 'socket-open' })
      formatTimer = scope.setTimeout(() => terminal('media FORMAT deadline exceeded'), 5000)
      scheduleFeedback()
    }
    candidate.onmessage = (event) => {
      if (socket !== candidate) return
      if (!(event.data instanceof ArrayBuffer)) {
        terminal('server sent a non-binary media record')
        return
      }
      handleRecord(event.data)
    }
    candidate.onerror = () => {
      if (socket === candidate) post({ type: 'socket-error' })
    }
    candidate.onclose = (event) => {
      if (socket !== candidate) return
      socket = undefined
      post({ type: 'clock-reset', kind: 'all' })
      resetRuntime()
      if (!intentionalClose) {
        post({
          type: 'socket-close',
          code: event.code,
          reason: event.reason,
          recoverable: isMediaRetryCloseCode(event.code),
        })
      }
    }
  } catch (error) {
    terminal(error instanceof Error ? error.message : String(error))
  }
}

scope.onmessage = (event: MessageEvent) => {
  const message = event.data
  switch (message?.type) {
    case 'probe':
      probeCapabilities(message.capabilities, message.allowAudio)
      break
    case 'start':
      startMedia(message.url, message.ticket, message.audioEnabled)
      break
    case 'frame-release':
      if (message.generation === video.generation) {
        video.outstandingVideo = Math.max(0, video.outstandingVideo - 1)
        // `rendered` is the server's contiguous presentation-progress cursor.
        // A deliberately discarded decoded frame still advances that cursor;
        // `drops` separately records that it was not drawn.
        if (video.rendered < video.decoded) video.rendered++
        if (!message.rendered) video.drops++
        avSkewMS = Math.max(-10_000, Math.min(10_000, Number(message.skewMS) || 0))
        // The server evaluates the reported skew over consecutive feedback
        // windows. Keeping one recovery authority avoids a client request and
        // server observation racing into the bounded resync limit.
        drain(video)
      }
      break
    case 'audio-release':
      if (message.generation === audio.generation) {
        audio.bufferedAudioMS = Math.max(0, audio.bufferedAudioMS - Math.max(0, Number(message.durationMS) || 0))
        if (audio.rendered < audio.decoded) audio.rendered++
        if (!message.rendered) audio.drops++
        drain(audio)
      }
      break
    case 'resync':
      requestResync(message.reason || 'timestamp')
      break
    case 'stop':
      intentionalClose = true
      sendControl({ type: 'stop' })
      closeSocket(1000, 'client_stop')
      resetRuntime()
      break
  }
}
