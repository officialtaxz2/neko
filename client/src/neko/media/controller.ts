import { EVENT, WebSocketEvents } from '../events'
import {
  MediaCapabilitiesPayload,
  MediaCreateChoice,
  MediaCreatePayload,
  MediaOfferPayload,
  WebSocketPayloads,
} from '../messages'
import { MEDIA_BACKEND, MEDIA_PROTOCOL } from './protocol'
import { MEDIA_RETRY_DELAYS_MS, mediaRetryDelayForAttempt } from './recovery.js'

export interface ScheduledVideoFrame {
  frame: VideoFrame
  generation: number
  timestamp: number
  targetTime: number
  receivedAt: number
}

export interface WebCodecsControllerCallbacks {
  sendEvent: (event: WebSocketEvents, payload?: WebSocketPayloads) => void
  eventSocketOpen: () => boolean
  setStatus: (status: 'idle' | 'negotiating' | 'connecting' | 'streaming' | 'reconnecting' | 'terminal', detail?: string, retryAttempt?: number) => void
  setAudioEnabled: (enabled: boolean) => void
  setPlayable: (playable: boolean) => void
  onVideoFormat: (format: { width: number; height: number; rate: number }) => void
  onVideoFrame: (frame: ScheduledVideoFrame) => void
  onClockReset: () => void
}

const NEGOTIATION_TIMEOUT_MS = 5000

export class WebCodecsMediaController {
  private worker?: Worker
  private audioContext?: AudioContext
  private audioNode?: AudioWorkletNode
  private gainNode?: GainNode
  private audioAllowed = false
  private audioEnabled = false
  private audioAnchorPTS?: number
  private audioAnchorTime?: number
  private videoAnchorPTS?: number
  private videoAnchorTime?: number
  private audioChunkID = 0
  private choice?: { audio: MediaCreateChoice | null; video: MediaCreateChoice }
  private negotiationTimer?: number
  private retryTimer?: number
  private retryAttempt = 0
  private awaitingOffer = false
  private recovering = false
  private stopped = false
  private muted = false
  private volume = 1

  constructor(private eventURL: string, private callbacks: WebCodecsControllerCallbacks) {}

  public async start() {
    this.stopped = false
    this.clearTimers()
    this.retryAttempt = 0
    this.recovering = false
    this.choice = undefined
    this.audioAllowed = false
    this.audioEnabled = false
    this.callbacks.setPlayable(false)
    this.callbacks.setAudioEnabled(false)
    this.callbacks.setStatus('negotiating', 'Checking browser and server capabilities')
    this.createWorker()
    this.audioAllowed = await this.prepareAudio()
    if (this.stopped || !this.callbacks.eventSocketOpen()) return
    this.callbacks.sendEvent(EVENT.MEDIA.CAPABILITIES_REQUEST, { version: 1 })
    this.armNegotiationTimeout('The server did not advertise the opt-in backend')
  }

  public stop() {
    this.stopped = true
    this.clearTimers()
    this.awaitingOffer = false
    this.audioAllowed = false
    this.audioEnabled = false
    if (this.worker) {
      this.worker.postMessage({ type: 'stop' })
      this.worker.terminate()
      this.worker = undefined
    }
    this.resetClock()
    this.releaseAudioGraph()
    this.callbacks.setPlayable(false)
    this.callbacks.setAudioEnabled(false)
  }

  public handleCapabilities(payload: MediaCapabilitiesPayload) {
    if (this.stopped || !this.worker) return
    this.clearNegotiationTimer()
    this.armNegotiationTimeout('Browser codec capability probing timed out')
    this.worker.postMessage({ type: 'probe', capabilities: payload, allowAudio: this.audioAllowed })
  }

  public handleOffer(payload: MediaOfferPayload) {
    if (this.stopped || !this.worker || !this.awaitingOffer || !this.choice) return
    this.awaitingOffer = false
    this.clearNegotiationTimer()
    if (
      payload.version !== 1 ||
      payload.backend !== MEDIA_BACKEND ||
      payload.protocol !== MEDIA_PROTOCOL ||
      payload.path !== '/api/media/ws' ||
      !/^[A-Za-z0-9_-]{32}$/.test(payload.ticket) ||
      !Number.isSafeInteger(payload.expires_in_ms) ||
      payload.expires_in_ms <= 0 ||
      payload.expires_in_ms > 10_000
    ) {
      this.fail('The server returned an invalid media offer')
      return
    }

    this.callbacks.setStatus(this.recovering ? 'reconnecting' : 'connecting', 'Opening dedicated media socket', this.retryAttempt)
    this.worker.postMessage({
      type: 'start',
      url: this.mediaURL(payload.path),
      ticket: payload.ticket,
      audioEnabled: this.choice.audio !== null,
    })
  }

  public retry() {
    if (this.stopped || !this.callbacks.eventSocketOpen()) return
    this.stop()
    this.start().catch((error) => this.fail(error instanceof Error ? error.message : String(error)))
  }

  public async play() {
    if (!this.audioEnabled || !this.audioContext) return true
    const wasRunning = this.audioContext.state === 'running'
    try {
      await this.audioContext.resume()
    } catch (_) {}
    const running = this.audioContext.state === 'running'
    if (running && !wasRunning) {
      this.resetClock()
      this.worker?.postMessage({ type: 'resync', reason: 'timestamp' })
    }
    return running
  }

  public setMuted(muted: boolean) {
    this.muted = muted
    this.applyGain()
  }

  public setVolume(volume: number) {
    this.volume = Math.max(0, Math.min(1, volume))
    this.applyGain()
  }

  public releaseVideoFrame(frame: ScheduledVideoFrame, rendered: boolean, skewMS: number) {
    try {
      frame.frame.close()
    } catch (_) {}
    this.worker?.postMessage({ type: 'frame-release', generation: frame.generation, rendered, skewMS })
  }

  public requestResync(reason: 'queue_overflow' | 'decoder_error' | 'timestamp' | 'audio_underflow' | 'av_skew') {
    this.worker?.postMessage({ type: 'resync', reason })
  }

  private createWorker() {
    if (typeof Worker === 'undefined') throw new Error('Worker is unavailable')
    const worker = new Worker(new URL('./worker.ts', import.meta.url), { type: 'module', name: 'neko-media-v1' })
    this.worker = worker
    worker.onmessage = ({ data }) => {
      if (this.worker !== worker || this.stopped) {
        if (data?.type === 'video-frame' && data.frame) data.frame.close()
        return
      }
      this.onWorkerMessage(data)
    }
    worker.onerror = () => {
      if (this.worker === worker && !this.stopped) this.fail('The dedicated media worker failed')
    }
    worker.onmessageerror = () => {
      if (this.worker === worker && !this.stopped) this.fail('The dedicated media worker returned an invalid message')
    }
  }

  private onWorkerMessage(message: any) {
    switch (message?.type) {
      case 'capability-choice':
        this.clearNegotiationTimer()
        this.choice = { audio: message.audio, video: message.video }
        this.audioEnabled = message.audio !== null
        if (!this.audioEnabled) this.releaseAudioGraph()
        this.callbacks.setAudioEnabled(this.audioEnabled)
        this.callbacks.setPlayable(true)
        this.requestOffer()
        break
      case 'socket-open':
        this.callbacks.setStatus(this.recovering ? 'reconnecting' : 'connecting', 'Waiting for exact media FORMAT', this.retryAttempt)
        break
      case 'ready':
        this.recovering = false
        this.retryAttempt = 0
        this.callbacks.setStatus('streaming', this.audioEnabled ? 'VP8 video and Opus audio' : 'VP8 video; audio disabled')
        break
      case 'video-format':
        this.callbacks.onVideoFormat({
          width: message.displayWidth || message.codedWidth,
          height: message.displayHeight || message.codedHeight,
          rate: Math.max(1, Math.round(message.frameRateNumerator / message.frameRateDenominator)),
        })
        break
      case 'video-frame':
        this.onVideoFrame(message)
        break
      case 'audio-data':
        this.onAudioData(message)
        break
      case 'clock-reset':
        if (message.kind === 'video') this.resetVideoClock()
        else this.resetClock()
        break
      case 'socket-close':
        if (message.recoverable && this.callbacks.eventSocketOpen()) this.scheduleRetry(message.reason || `close ${message.code}`)
        else this.fail(`Media socket closed (${message.code}${message.reason ? `: ${message.reason}` : ''})`)
        break
      case 'end':
        this.fail(`Media delivery ended: ${message.reason}`)
        break
      case 'terminal':
        this.fail(message.detail || 'Media protocol failure')
        break
    }
  }

  private requestOffer() {
    if (this.stopped || !this.choice || !this.callbacks.eventSocketOpen()) {
      this.fail('The authenticated event session is no longer available')
      return
    }
    this.awaitingOffer = true
    const payload: MediaCreatePayload = {
      version: 1,
      backend: MEDIA_BACKEND,
      audio: this.choice.audio,
      video: this.choice.video,
    }
    this.callbacks.setStatus(this.recovering ? 'reconnecting' : 'negotiating', 'Requesting a single-use media ticket', this.retryAttempt)
    this.callbacks.sendEvent(EVENT.MEDIA.CREATE, payload)
    this.armNegotiationTimeout('The server did not issue a media ticket')
  }

  private scheduleRetry(detail: string) {
    this.clearNegotiationTimer()
    this.awaitingOffer = false
    if (this.retryTimer !== undefined) return
    const delay = mediaRetryDelayForAttempt(this.retryAttempt)
    if (delay === null) {
      this.fail(`WebCodecs retry limit reached: ${detail}`)
      return
    }
    this.retryAttempt++
    this.recovering = true
    this.callbacks.setStatus(
      'reconnecting',
      `Media-only retry ${this.retryAttempt}/${MEDIA_RETRY_DELAYS_MS.length} in ${delay / 1000}s`,
      this.retryAttempt,
    )
    this.retryTimer = window.setTimeout(() => {
      this.retryTimer = undefined
      if (!this.callbacks.eventSocketOpen()) {
        this.fail('The authenticated event session closed during media recovery')
        return
      }
      this.requestOffer()
    }, delay)
  }

  private async prepareAudio() {
    const AudioContextConstructor = window.AudioContext
    if (!AudioContextConstructor || typeof AudioWorkletNode === 'undefined') return false
    let context: AudioContext | undefined
    try {
      context = new AudioContextConstructor({ latencyHint: 'interactive', sampleRate: 48000 })
      if (!context.audioWorklet || context.sampleRate !== 48000) {
        await context.close()
        return false
      }
      await context.audioWorklet.addModule(new URL('./audio-worklet.js', import.meta.url))
      if (this.stopped) {
        await context.close()
        return false
      }
      const node = new AudioWorkletNode(context, 'neko-media-audio', { numberOfInputs: 0, numberOfOutputs: 1, outputChannelCount: [2] })
      const gain = context.createGain()
      node.connect(gain).connect(context.destination)
      node.port.onmessage = ({ data }) => {
        if (this.audioNode !== node || this.stopped) return
        if (data?.type === 'consumed') {
          this.worker?.postMessage({
            type: 'audio-release',
            generation: data.generation,
            durationMS: data.durationMS,
            rendered: true,
          })
        } else if (data?.type === 'underflow' || data?.type === 'overflow') {
          this.worker?.postMessage({ type: data.type === 'underflow' ? 'audio-underflow' : 'resync', reason: 'queue_overflow' })
        }
      }
      this.audioContext = context
      this.audioNode = node
      this.gainNode = gain
      this.applyGain()
      return true
    } catch (_) {
      if (context) await context.close().catch(() => {})
      return false
    }
  }

  private onAudioData(message: any) {
    const context = this.audioContext
    const node = this.audioNode
    const durationMS = Number(message.duration) / 1000
    if (!context || !node || context.state !== 'running') {
      this.worker?.postMessage({
        type: 'audio-release',
        generation: message.generation,
        durationMS,
        rendered: false,
      })
      return
    }

    if (this.audioAnchorPTS === undefined || this.audioAnchorTime === undefined) {
      this.audioAnchorPTS = message.timestamp
      this.audioAnchorTime = context.currentTime + 0.08
      this.videoAnchorPTS = undefined
      this.videoAnchorTime = undefined
      this.callbacks.onClockReset()
      node.port.postMessage({ type: 'reset' })
    }
    const startTime = this.audioAnchorTime + (message.timestamp - this.audioAnchorPTS) / 1_000_000
    const id = ++this.audioChunkID
    node.port.postMessage(
      {
        type: 'chunk',
        id,
        generation: message.generation,
        durationMS,
        startTime,
        numberOfFrames: message.numberOfFrames,
        planes: message.planes,
      },
      message.planes,
    )
  }

  private onVideoFrame(message: any) {
    const now = performance.now()
    let targetTime: number
    if (
      this.audioEnabled &&
      this.audioContext?.state === 'running' &&
      this.audioAnchorPTS !== undefined &&
      this.audioAnchorTime !== undefined
    ) {
      targetTime = now + (this.audioAnchorTime + (message.timestamp - this.audioAnchorPTS) / 1_000_000 - this.audioContext.currentTime) * 1000
    } else {
      if (this.videoAnchorPTS === undefined || this.videoAnchorTime === undefined) {
        this.videoAnchorPTS = message.timestamp
        this.videoAnchorTime = now + 80
      }
      targetTime = this.videoAnchorTime + (message.timestamp - this.videoAnchorPTS) / 1000
    }
    this.callbacks.onVideoFrame({
      frame: message.frame,
      generation: message.generation,
      timestamp: message.timestamp,
      targetTime,
      receivedAt: now,
    })
  }

  private resetClock() {
    this.audioAnchorPTS = undefined
    this.audioAnchorTime = undefined
    this.audioNode?.port.postMessage({ type: 'reset' })
    this.resetVideoClock()
  }

  private resetVideoClock() {
    this.videoAnchorPTS = undefined
    this.videoAnchorTime = undefined
    this.callbacks.onClockReset()
  }

  private applyGain() {
    if (this.gainNode) this.gainNode.gain.value = this.muted ? 0 : this.volume
  }

  private releaseAudioGraph() {
    if (this.audioNode) this.audioNode.disconnect()
    if (this.gainNode) this.gainNode.disconnect()
    this.audioNode = undefined
    this.gainNode = undefined
    if (this.audioContext) this.audioContext.close().catch(() => {})
    this.audioContext = undefined
  }

  private mediaURL(path: string) {
    const url = new URL(this.eventURL)
    const prefix = url.pathname.replace(/\/ws$/, '')
    url.pathname = `${prefix}${path}`.replace(/\/{2,}/g, '/')
    url.search = ''
    url.hash = ''
    return url.toString()
  }

  private armNegotiationTimeout(detail: string) {
    this.clearNegotiationTimer()
    this.negotiationTimer = window.setTimeout(() => {
      this.negotiationTimer = undefined
      if (this.recovering) this.scheduleRetry(detail)
      else this.fail(detail)
    }, NEGOTIATION_TIMEOUT_MS)
  }

  private clearNegotiationTimer() {
    if (this.negotiationTimer !== undefined) window.clearTimeout(this.negotiationTimer)
    this.negotiationTimer = undefined
  }

  private clearTimers() {
    this.clearNegotiationTimer()
    if (this.retryTimer !== undefined) window.clearTimeout(this.retryTimer)
    this.retryTimer = undefined
  }

  private fail(detail: string) {
    if (this.stopped) return
    this.clearTimers()
    this.awaitingOffer = false
    this.recovering = false
    this.audioAllowed = false
    this.audioEnabled = false
    if (this.worker) {
      this.worker.postMessage({ type: 'stop' })
      this.worker.terminate()
      this.worker = undefined
    }
    this.resetClock()
    this.releaseAudioGraph()
    this.callbacks.setPlayable(false)
    this.callbacks.setAudioEnabled(false)
    this.callbacks.setStatus('terminal', detail)
  }
}
