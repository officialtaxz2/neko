// HLS / LL-HLS: passive playback for browsers and devices where WebRTC is no option (Smart TVs,
// strict networks). The server negotiates over the main socket (media/hls/*), hands out a one-time
// ticket, and the bootstrap POST turns it into a cookie-scoped lease; the picture then comes over
// plain HTTPS, natively (Safari, iOS, TVs) or through hls.js over MSE (./hls/player.ts, loaded
// lazily). The transport keeps the lease alive, watches it, follows the room's private mode, and on
// a failure stops with a fixed message (no automatic fallback): the room stays usable, retry()
// starts over. The contract: docs/HLS_LL_HLS.md; ported from the fork's Vue client.
import { Emitter } from '../emitter.ts'
import type { SessionInfo, StreamTransport, TransportEvents, TransportState, TransportStatus } from '../transport.ts'
import type { HLSPlayer } from './hls/player.ts'
import {
  clearHLSVideo,
  hlsBase,
  hlsLeaseURL,
  hlsMasterChildren,
  hlsModes,
  HLS_AUDIO_TYPE,
  HLS_VIDEO_TYPE,
  terminalHLSStatus,
  validHLSOffer,
  validHLSPlaylist,
  type HLSMode,
} from './hls/protocol.ts'

export type { HLSMode }
export { hlsModes }

export type HLSStatus = 'idle' | 'negotiating' | 'connecting' | 'streaming' | 'paused' | 'terminal'

export interface HLSState {
  mode: HLSMode
  status: HLSStatus
  // a fixed English message for paused and terminal (never a URL, ticket or cookie)
  detail: string
  player: '' | 'native' | 'mse'
}

export interface HLSTransportEvents extends TransportEvents {
  hls: (state: HLSState) => void
}

export interface HLSTransportOptions {
  mode: HLSMode
  // whether to play as soon as a source is attached: native HLS in Safari loads nothing before play()
  autoplay?: () => boolean
  // where the page serves hls.js's worker (hls.js/dist/hls.worker.js), for the MSE player
  workerPath?: string
}

const NEGOTIATION_MS = 5000
const BOOTSTRAP_MS = 30000
const KEEPALIVE_MS = 15000
const WATCH_MS = 1000
const READY_MS = 30000
const STALL_MS = 20000
const NOT_READY_MAX = 30 // one per WATCH_MS
const HTTP_FAILURES_MAX = 3

const TRANSPORT_STATUS: Record<HLSStatus, TransportStatus> = {
  idle: 'disconnected',
  negotiating: 'connecting',
  // the server granted the stream: the room is up, the picture follows (or the room is private)
  connecting: 'connected',
  streaming: 'connected',
  paused: 'connected',
  terminal: 'unavailable',
}

export class HLSTransport implements StreamTransport {
  readonly kind = 'hls' as const
  readonly mode: HLSMode

  static supported() {
    return typeof fetch === 'function' && typeof AbortController === 'function'
  }

  private readonly opts: HLSTransportOptions
  private readonly events = new Emitter<HLSTransportEvents>()
  private readonly state: TransportState = {
    status: 'disconnected',
    size: { width: 0, height: 0 },
    video: { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false },
  }
  private readonly info: HLSState
  private session: SessionInfo | null = null
  private unsubs: (() => void)[] = []
  private generation = 0
  private playerGeneration = 0
  private stopped = true
  private phase: 'off' | 'capabilities' | 'offer' | 'lease' = 'off'
  private base = ''
  private master = ''
  private masterText = ''
  private video: HTMLVideoElement | null = null
  private player?: HLSPlayer
  private privatePaused = false
  private desiredPlaying = false
  private muted = false
  private volume = 1
  private playPending?: Promise<boolean>
  private playSequence = 0
  private timers = new Set<number>()
  private requests = new Set<AbortController>()
  private playerRequests = new Set<AbortController>()
  private readinessTimer?: number
  private listeners: (() => void)[] = []
  private attaching = false
  private playbackReady = false
  private lastTime = 0
  private lastProgress = 0
  private statusFailures = 0
  private readinessFailures = 0
  private statusRunning = false
  private keepAliveAt = 0

  constructor(opts: HLSTransportOptions) {
    this.opts = opts
    this.mode = opts.mode
    this.info = { mode: opts.mode, status: 'idle', detail: '', player: '' }
  }

  get element() {
    return this.video
  }

  get hls(): HLSState {
    return { ...this.info }
  }

  on<K extends keyof HLSTransportEvents>(event: K, cb: HLSTransportEvents[K]) {
    return this.events.on(event, cb)
  }

  /////////////////////////////
  // session
  /////////////////////////////

  // on every (re)connect: a fresh negotiation and lease; private mode as the room says
  async connect(session: SessionInfo) {
    this.unsubscribe()
    this.session = session
    const on = (event: string, cb: (p: any) => void) => this.unsubs.push(session.on(event, cb))
    const me = session.init.session_id
    let privateMode = !!session.init.settings?.private_mode
    let admin = !!session.init.sessions?.[me]?.profile?.is_admin
    // the server pauses the lease itself (PrivateModeEnabled: private mode, not an admin); the
    // transport follows so the buffered picture goes too
    const follow = () => this.setPrivatePaused(privateMode && !admin)
    on('media/hls/capabilities', (p) => this.onCapabilities(p))
    on('media/hls/offer', (p) => this.onOffer(p))
    on('system/settings', (p) => {
      privateMode = !!p?.private_mode
      follow()
    })
    on('session/profile', (p) => {
      if (p?.id !== me) return
      admin = !!p.is_admin
      follow()
    })
    this.privatePaused = privateMode && !admin
    this.negotiate()
  }

  // stop for good: the lease is left to expire, the picture goes
  close() {
    this.unsubscribe()
    this.session = null
    this.stop()
    this.desiredPlaying = false
    this.setHLS('idle')
  }

  // after a failure: start over with the same session
  retry() {
    if (this.session) this.negotiate()
  }

  private unsubscribe() {
    this.unsubs.splice(0).forEach((off) => off())
  }

  private negotiate() {
    this.stop()
    this.stopped = false
    this.desiredPlaying = !!this.opts.autoplay?.()
    if (!this.session || !HLSTransport.supported()) {
      this.fail('HLS requires an authenticated session and HTTP playback support')
      return
    }
    try {
      this.base = hlsBase(this.session.url, location.href)
    } catch {
      this.fail('HLS requires same-origin HTTPS')
      return
    }
    this.phase = 'capabilities'
    this.setHLS('negotiating')
    this.session.send('media/hls/capabilities/request', { version: 1, mode: this.mode })
    this.later(() => {
      if (this.phase === 'capabilities' || this.phase === 'offer')
        this.fail('Selected HLS mode was not advertised or negotiation timed out')
    }, NEGOTIATION_MS)
  }

  private onCapabilities(payload: unknown) {
    if (this.stopped || this.phase !== 'capabilities') return
    if (!hlsModes(payload).includes(this.mode)) {
      this.fail('Selected HLS mode is unavailable')
      return
    }
    this.phase = 'offer'
    this.session?.send('media/hls/create', { version: 1, backend: 'hls', mode: this.mode })
  }

  private onOffer(payload: any) {
    if (this.stopped || this.phase !== 'offer') return
    if (!validHLSOffer(payload, this.mode)) {
      this.fail('Invalid HLS offer')
      return
    }
    this.phase = 'lease'
    this.setHLS('connecting')
    const generation = this.generation
    // the one-time ticket travels only in this POST body: never stored, never logged
    void this.request(
      new URL('session', this.base).href,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ticket: payload.ticket }),
      },
      BOOTSTRAP_MS,
    )
      .then(async (response) => {
        if (!this.current(generation)) return
        if (response.status !== 201) {
          this.fail(
            response.status === 503
              ? 'HLS media was unavailable at startup; retry manually'
              : 'HLS bootstrap failed; retry manually',
          )
          return
        }
        const text = await response.text()
        if (!this.current(generation)) return
        if (text.length > 2048) throw new Error('bounded bootstrap')
        this.master = hlsLeaseURL(JSON.parse(text), this.mode, this.base)
        this.keepAliveAt = Date.now() + KEEPALIVE_MS
        this.statusFailures = 0
        this.readinessFailures = 0
        this.watchLease()
      })
      .catch(() => {
        if (this.current(generation)) this.fail('HLS bootstrap failed; retry manually')
      })
  }

  private setPrivatePaused(paused: boolean) {
    if (this.privatePaused === paused) return
    this.privatePaused = paused
    if (this.stopped) return
    // a pause or resume the room decided starts a new readiness interval, without the old warm-up
    // count or the old HTTP failures
    this.statusFailures = 0
    this.readinessFailures = 0
    if (paused) {
      this.clearPlayer()
      this.setHLS('paused', 'Private mode')
    } else if (this.master) {
      // the lease is still valid; the watch waits for freshly packaged media
      this.setHLS('connecting')
    }
  }

  /////////////////////////////
  // the element
  /////////////////////////////

  attach(container: HTMLElement) {
    if (this.video) {
      this.clearPlayer()
      this.video.remove() // a second attach replaces the element
    }
    const video = (this.video = document.createElement('video'))
    video.playsInline = true
    video.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;background:transparent'
    video.muted = this.muted
    try {
      video.volume = this.volume
    } catch {}
    const v = this.state.video
    video.addEventListener('volumechange', () => {
      v.muted = this.muted = video.muted
      v.volume = this.volume = video.volume
      this.emitState()
    })
    v.muted = video.muted
    v.volume = video.volume
    container.append(video)
    this.emitState()
    if (this.master && !this.privatePaused && !this.stopped) void this.openPlayer()
    return () => {
      video.remove()
      if (this.video !== video) return
      this.clearPlayer()
      this.video = null
    }
  }

  setPlaying(on: boolean) {
    if (!on) {
      this.pause()
      return Promise.resolve()
    }
    return this.play().then(() => {})
  }

  setVolume(volume: number) {
    this.volume = Math.max(0, Math.min(1, volume))
    if (this.video) {
      try {
        this.video.volume = this.volume
      } catch {}
    }
  }

  setMuted(on: boolean) {
    this.muted = on
    if (this.video) this.video.muted = on
    if (!on && this.state.video.mutedByAutoplay) {
      this.state.video.mutedByAutoplay = false
      this.emitState()
    }
  }

  // resolves whether playback runs; one muted retry when the browser refuses sound
  play(): Promise<boolean> {
    this.desiredPlaying = true
    if (this.playPending) return this.playPending
    const video = this.video
    if (!video || this.stopped || this.privatePaused || !this.player) return Promise.resolve(false)
    const generation = this.generation
    const playerGeneration = this.playerGeneration
    const playSequence = ++this.playSequence
    const current = () =>
      this.current(generation) &&
      this.playerGeneration === playerGeneration &&
      this.video === video &&
      !this.privatePaused &&
      this.desiredPlaying &&
      this.playSequence === playSequence
    // a deliberate pause is no stall: measure afresh when resuming; repeated play() on an unpaused
    // stall cannot defer the watchdog
    if (video.paused) this.resetProgress(video)
    // play() right away, inside the user's gesture, before anything asynchronous
    let attempt: Promise<void>
    try {
      attempt = video.play()
    } catch {
      attempt = Promise.reject(new Error('play blocked'))
    }
    const pending = (async () => {
      try {
        await attempt
        return current()
      } catch {
        if (!current()) return false
        if (!video.muted) {
          this.setMuted(true)
          this.state.video.mutedByAutoplay = true
          this.emitState()
          try {
            await video.play()
            if (typeof document !== 'undefined')
              document.addEventListener('click', () => this.setMuted(false), { once: true })
            return current()
          } catch {}
        }
        if (current()) {
          this.desiredPlaying = false
          this.setPlayingState(false)
        }
        return false
      }
    })()
    this.playPending = pending
    void pending.finally(() => {
      if (this.playPending === pending) this.playPending = undefined
    })
    return pending
  }

  pause() {
    this.desiredPlaying = false
    this.playSequence++
    this.playPending = undefined
    this.video?.pause()
    this.setPlayingState(false)
  }

  /////////////////////////////
  // the lease
  /////////////////////////////

  private stop() {
    this.stopped = true
    this.generation++
    this.phase = 'off'
    this.master = ''
    this.masterText = ''
    for (const timer of this.timers) window.clearTimeout(timer)
    this.timers.clear()
    for (const request of this.requests) request.abort()
    this.requests.clear()
    this.statusRunning = false
    this.clearPlayer()
  }

  private current(generation: number) {
    return !this.stopped && this.generation === generation
  }

  private later(callback: () => void, ms: number) {
    const generation = this.generation
    const timer = window.setTimeout(() => {
      this.timers.delete(timer)
      if (this.current(generation)) callback()
    }, ms)
    this.timers.add(timer)
    return timer
  }

  // one bounded request: the body is read under the same deadline, so headers alone cannot hold
  // the cleanup
  private async request(url: string, init: RequestInit = {}, timeout = 1000, playerRequest = false) {
    const controller = new AbortController()
    this.requests.add(controller)
    if (playerRequest) this.playerRequests.add(controller)
    const timer = window.setTimeout(() => controller.abort(), timeout)
    this.timers.add(timer)
    try {
      const response = await fetch(url, {
        ...init,
        credentials: 'same-origin',
        cache: 'no-store',
        redirect: 'error',
        referrerPolicy: 'no-referrer',
        signal: controller.signal,
      })
      const maximum = init.method === 'POST' ? 2048 : 65536
      if (Number(response.headers.get('Content-Length')) > maximum) throw new Error('bounded response')
      const reader = typeof TextDecoder === 'function' ? response.body?.getReader() : undefined
      let body = ''
      if (reader) {
        const decoder = new TextDecoder()
        let size = 0
        for (;;) {
          const { done, value } = await reader.read()
          if (done) break
          size += value.byteLength
          if (size > maximum) {
            controller.abort()
            throw new Error('bounded response')
          }
          body += decoder.decode(value, { stream: true })
        }
        body += decoder.decode()
      } else {
        body = await response.text()
        if (body.length > maximum) throw new Error('bounded response')
      }
      return { status: response.status, ok: response.ok, text: async () => body }
    } finally {
      controller.abort()
      window.clearTimeout(timer)
      this.timers.delete(timer)
      this.requests.delete(controller)
      this.playerRequests.delete(controller)
    }
  }

  // once a second: the keepalive when due, and a small probe of the master, which native playback
  // needs (it hides its HTTP errors) and which notices a revoked or expired lease
  private watchLease() {
    if (this.stopped || !this.master || this.statusRunning) return
    this.statusRunning = true
    const generation = this.generation
    void (async () => {
      try {
        if (!this.session) {
          this.fail('HLS session ended')
          return
        }
        if (Date.now() >= this.keepAliveAt) {
          const alive = await this.request(new URL('keepalive', new URL('.', this.master)).href, { method: 'POST' })
          if (!this.current(generation)) return
          if (terminalHLSStatus(alive.status)) {
            this.fail('HLS lease ended')
            return
          }
          if (!alive.ok) throw new Error('keepalive')
          this.keepAliveAt = Date.now() + KEEPALIVE_MS
        }
        const response = await this.request(this.master)
        if (!this.current(generation)) return
        if (terminalHLSStatus(response.status)) {
          this.fail('HLS authorization or media availability ended')
          return
        }
        if (response.status === 503) {
          // not ready yet is no HTTP failure
          this.statusFailures = 0
          this.clearPlayer()
          if (this.privatePaused) this.setHLS('paused', 'Private mode')
          else this.setHLS('connecting', 'Waiting for fresh media')
          if (!this.privatePaused && ++this.readinessFailures >= NOT_READY_MAX)
            this.fail('HLS media did not become ready')
          return
        }
        const text = await response.text()
        if (!response.ok || !validHLSPlaylist(text, this.master, this.master, this.mode)) throw new Error('playlist')
        this.masterText = text
        this.statusFailures = 0
        this.readinessFailures = 0
        if (!this.privatePaused && !this.player && !this.attaching) await this.openPlayer()
        if (!this.current(generation)) return
        const video = this.video
        if (video && this.player && this.playbackReady && this.desiredPlaying && !video.paused) {
          // a live-edge seek can move currentTime while the picture stays frozen: only forward
          // playback with current data counts
          if (!video.seeking && video.readyState >= 2 && video.currentTime > this.lastTime)
            this.lastProgress = Date.now()
          this.lastTime = video.currentTime
          if (Date.now() - this.lastProgress > STALL_MS) this.fail('HLS playback stalled; retry manually')
        }
      } catch {
        if (this.current(generation) && ++this.statusFailures >= HTTP_FAILURES_MAX)
          this.fail('HLS HTTP connection failed; retry manually')
      } finally {
        if (this.current(generation)) {
          this.statusRunning = false
          this.later(() => this.watchLease(), WATCH_MS)
        }
      }
    })()
  }

  /////////////////////////////
  // the player
  /////////////////////////////

  private async openPlayer() {
    const video = this.video
    if (!video || this.stopped || this.privatePaused || !this.master || this.attaching) return
    this.attaching = true
    const generation = this.generation
    const playerGeneration = ++this.playerGeneration
    const current = () =>
      this.current(generation) &&
      this.playerGeneration === playerGeneration &&
      this.video === video &&
      !this.privatePaused
    try {
      const native =
        !!video.canPlayType('application/vnd.apple.mpegurl') || !!video.canPlayType('application/x-mpegURL')
      const nativeCodec = !!video.canPlayType(HLS_VIDEO_TYPE) && !!video.canPlayType(HLS_AUDIO_TYPE)
      // Apple's own player keeps native fullscreen and PiP; elsewhere canPlayType alone is no proof
      // (navigator.vendor is deprecated, and still the one way to tell Safari from other WebKit shells)
      const vendor = (navigator as { vendor?: string }).vendor ?? ''
      const apple = /iP(?:hone|ad|od)|Macintosh/.test(navigator.userAgent) && /Apple/.test(vendor)
      let kind: 'native' | 'mse' = 'native'
      let module: typeof import('./hls/player.ts') | undefined
      if (!apple || !native) {
        const mediaSource = (window as any).MediaSource || (window as any).ManagedMediaSource
        const mseCodec =
          typeof mediaSource?.isTypeSupported === 'function' &&
          mediaSource.isTypeSupported(HLS_VIDEO_TYPE) &&
          mediaSource.isTypeSupported(HLS_AUDIO_TYPE)
        if (mseCodec) {
          module = await import('./hls/player.ts')
          if (!current()) return
          if (module.mseSupported()) kind = 'mse'
        }
        if (kind !== 'mse' && (!native || !nativeCodec)) {
          this.fail('HLS H.264/AAC playback is unsupported in this browser')
          return
        }
      }
      if (!current()) return
      if (kind === 'native') {
        // native loaders cannot be intercepted: check the master and every media playlist it names
        // before handing the element a URL
        for (const url of hlsMasterChildren(this.masterText, this.master, this.mode)) {
          if (!current()) return
          const response = await this.request(url, {}, 5000, true)
          if (!current()) return
          if (!response.ok || !validHLSPlaylist(await response.text(), url, this.master, this.mode))
            throw new Error('native playlist')
        }
      }
      if (!current()) return
      const listen = (event: string, callback: () => void) => {
        const handler = () => {
          if (current()) callback()
        }
        video.addEventListener(event, handler)
        this.listeners.push(() => video.removeEventListener(event, handler))
      }
      const resolution = () => {
        if (!video.videoWidth || !video.videoHeight) return
        this.state.size = { width: video.videoWidth, height: video.videoHeight }
        this.emitState()
      }
      listen('canplay', () => {
        this.clearReadinessTimer()
        if (!this.playbackReady) this.resetProgress(video)
        this.playbackReady = true
        this.setPlayable(true)
        resolution()
        if (this.desiredPlaying && video.paused) void this.play()
      })
      listen('resize', resolution)
      listen('playing', () => {
        if (video.paused) return // an event already queued cannot undo a pause
        this.clearReadinessTimer()
        this.playbackReady = true
        this.resetProgress(video)
        this.setPlayable(true)
        this.setPlayingState(true)
        this.setHLS('streaming')
      })
      listen('pause', () => this.setPlayingState(false))
      // a seek that completes between two probes does not count as progress either
      listen('seeking', () => (this.lastTime = video.currentTime))
      listen('error', () => this.fail('HLS media playback failed; retry manually'))
      listen('ended', () => this.fail('HLS playback ended; retry manually'))
      video.muted = this.muted
      try {
        video.volume = this.volume
      } catch {}
      this.resetProgress(video)
      this.info.player = kind
      this.setHLS('connecting')
      // this deadline covers the first readiness only (later buffering is the stall watch's);
      // armed before attaching, so even an immediate canplay or playing event cancels it
      this.readinessTimer = this.later(() => {
        if (!current()) return
        this.readinessTimer = undefined
        if (video.readyState < 3) this.fail('HLS playback did not become ready; retry manually')
        else {
          this.playbackReady = true
          this.resetProgress(video)
        }
      }, READY_MS)
      if (kind === 'mse') {
        const player = module!.createMSEPlayer(
          video,
          this.master,
          this.mode,
          (detail) => {
            if (current()) this.fail(detail)
          },
          this.opts.workerPath,
        )
        if (!current()) {
          player.destroy()
          return
        }
        this.player = player
      } else {
        this.player = { destroy() {} }
        video.src = this.master
        video.load()
      }
      // ask for playback once the source is attached, before canplay: play() can wait for data, and
      // native HLS may load nothing until it is called
      if (current() && this.desiredPlaying && video.paused) void this.play()
    } catch {
      if (current()) this.fail('HLS player could not start; retry manually')
    } finally {
      if (this.playerGeneration === playerGeneration) this.attaching = false
    }
  }

  private resetProgress(video: HTMLVideoElement) {
    this.lastTime = video.currentTime
    this.lastProgress = Date.now()
  }

  private clearReadinessTimer() {
    if (this.readinessTimer === undefined) return
    window.clearTimeout(this.readinessTimer)
    this.timers.delete(this.readinessTimer)
    this.readinessTimer = undefined
  }

  private clearPlayer() {
    this.playerGeneration++
    this.playbackReady = false
    this.clearReadinessTimer()
    for (const request of this.playerRequests) request.abort()
    this.playerRequests.clear()
    this.attaching = false
    this.playPending = undefined
    for (const remove of this.listeners) remove()
    this.listeners = []
    this.player?.destroy()
    this.player = undefined
    clearHLSVideo(this.video)
    this.info.player = ''
    this.state.video.playing = false
    this.setPlayable(false)
  }

  private fail(detail: string) {
    this.stop()
    this.setHLS('terminal', detail)
  }

  /////////////////////////////
  // state
  /////////////////////////////

  private setPlayable(on: boolean) {
    this.state.video.playable = on
    this.emitState()
  }

  private setPlayingState(on: boolean) {
    this.state.video.playing = on
    this.emitState()
  }

  private setHLS(status: HLSStatus, detail = '') {
    this.info.status = status
    this.info.detail = detail
    this.state.status = TRANSPORT_STATUS[status]
    this.events.emit('hls', this.hls)
    this.emitState()
  }

  private emitState() {
    const { status, size, video } = this.state
    this.events.emit('state', { status, size: { ...size }, video: { ...video } })
  }
}
