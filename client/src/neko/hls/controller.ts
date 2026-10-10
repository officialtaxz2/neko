import { EVENT } from '../events'
import type { HLSCapabilitiesPayload, HLSMode, HLSOfferPayload } from '../messages'
import type { HLSStatus } from '~/store/hls'
import type { HLSPlayer } from './player'
import { clearHLSVideo, hlsBase, hlsLeaseURL, hlsMasterChildren, hlsModes, HLS_AUDIO_TYPE, HLS_VIDEO_TYPE,
  terminalHLSStatus, validHLSOffer, validHLSPlaylist } from './protocol.js'

interface Callbacks {
  sendEvent: (event: typeof EVENT.HLS.CAPABILITIES_REQUEST | typeof EVENT.HLS.CREATE, payload: any) => void
  eventSocketOpen: () => boolean
  autoplay: () => boolean
  setStatus: (status: HLSStatus, detail?: string) => void
  setPlayer: (player: '' | 'native' | 'mse') => void
  setPlayable: (playable: boolean) => void
  setPlaying: (playing: boolean) => void
  setMuted: (muted: boolean) => void
  resolution: (width: number, height: number) => void
}

export class HLSMediaController {
  private generation = 0
  private playerGeneration = 0
  private stopped = true
  private phase: 'off' | 'capabilities' | 'offer' | 'lease' = 'off'
  private base = ''
  private master = ''
  private masterText = ''
  private video?: HTMLVideoElement
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
  private listeners: Array<() => void> = []
  private attaching = false
  private playbackReady = false
  private lastTime = 0
  private lastProgress = 0
  private statusFailures = 0
  private readinessFailures = 0
  private statusRunning = false
  private keepAliveAt = 0

  constructor(private eventURL: string, private mode: HLSMode, private callbacks: Callbacks) {}

  start() {
    this.stop()
    this.stopped = false
    this.desiredPlaying = this.callbacks.autoplay()
    if (!this.callbacks.eventSocketOpen() || typeof fetch !== 'function' || typeof AbortController !== 'function') {
      this.fail('HLS requires an authenticated session and HTTP playback support'); return
    }
    try { this.base = hlsBase(this.eventURL, location.href) } catch (_) { this.fail('HLS requires same-origin HTTPS'); return }
    this.phase = 'capabilities'
    this.callbacks.setStatus('negotiating')
    this.callbacks.sendEvent(EVENT.HLS.CAPABILITIES_REQUEST, { version: 1, mode: this.mode })
    this.later(() => { if (this.phase === 'capabilities' || this.phase === 'offer') this.fail('Selected HLS mode was not advertised or negotiation timed out') }, 5000)
  }

  handleCapabilities(payload: HLSCapabilitiesPayload) {
    if (this.stopped || this.phase !== 'capabilities') return
    if (!hlsModes(payload).includes(this.mode)) { this.fail('Selected HLS mode is unavailable'); return }
    this.phase = 'offer'
    this.callbacks.sendEvent(EVENT.HLS.CREATE, { version: 1, backend: 'hls', mode: this.mode })
  }

  handleOffer(payload: HLSOfferPayload) {
    if (this.stopped || this.phase !== 'offer') return
    if (!validHLSOffer(payload, this.mode)) { this.fail('Invalid HLS offer'); return }
    this.phase = 'lease'
    this.callbacks.setStatus('connecting')
    const generation = this.generation
    // The one-time ticket is sent only in this POST body. Never persist or log it.
    void this.request(new URL('session', this.base).href, { method: 'POST',
      headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ticket: payload.ticket }) }, 30000)
      .then(async (response) => {
        if (!this.current(generation)) return
        if (response.status !== 201) {
          this.fail(response.status === 503 ? 'HLS media was unavailable at startup; retry manually' : 'HLS bootstrap failed; retry manually')
          return
        }
        const text = await response.text()
        if (!this.current(generation)) return
        if (text.length > 2048) throw new Error('bounded bootstrap')
        this.master = hlsLeaseURL(JSON.parse(text), this.mode, this.base)
        this.keepAliveAt = Date.now() + 15000
        this.statusFailures = 0
        this.readinessFailures = 0
        this.watchLease()
      }).catch(() => { if (this.current(generation)) this.fail('HLS bootstrap failed; retry manually') })
  }

  attach(video: HTMLVideoElement) {
    if (this.video === video) return
    this.clearPlayer()
    this.video = video
    video.muted = this.muted
    try { video.volume = this.volume } catch (_) {}
    if (this.master && !this.privatePaused && !this.stopped) void this.openPlayer()
  }

  detach(video: HTMLVideoElement) {
    if (this.video !== video) return
    this.stop()
    this.video = undefined
  }

  setPrivatePaused(paused: boolean) {
    if (this.privatePaused === paused) return
    this.privatePaused = paused
    if (this.stopped) return
    // A room-authorized pause/resume starts a new readiness interval. It must
    // not inherit either the old warm-up count or old HTTP failures.
    this.statusFailures = 0
    this.readinessFailures = 0
    if (paused) {
      this.clearPlayer()
      this.callbacks.setStatus('paused', 'Private mode')
    } else if (this.master) {
      // Reuse the live lease; the watchdog waits for fresh complete packaging.
      this.callbacks.setStatus('connecting')
    }
  }

  setMuted(muted: boolean) { this.muted = muted; if (this.video) this.video.muted = muted }
  setVolume(volume: number) {
    this.volume = Math.max(0, Math.min(1, volume))
    if (this.video) { try { this.video.volume = this.volume } catch (_) {} }
  }

  play(): Promise<boolean> {
    this.desiredPlaying = true
    if (this.playPending) return this.playPending
    const video = this.video
    if (!video || this.stopped || this.privatePaused || !this.player) return Promise.resolve(false)
    const generation = this.generation
    const playerGeneration = this.playerGeneration
    const playSequence = ++this.playSequence
    const current = () => this.current(generation) && this.playerGeneration === playerGeneration && this.video === video &&
      !this.privatePaused && this.desiredPlaying && this.playSequence === playSequence
    // Deliberate pauses are not stalled playback. Start measuring afresh when
    // resuming, but repeated Play calls on an unpaused stall cannot defer it.
    if (video.paused) this.resetProgress(video)
    // Invoke play immediately from the gesture, before any asynchronous work.
    let attempt: Promise<void>
    try { attempt = video.play() } catch (_) { attempt = Promise.reject(new Error('play blocked')) }
    const pending = (async () => {
      try { await attempt; return current() }
      catch (_) {
        if (!current()) return false
        if (!video.muted) {
          this.setMuted(true)
          this.callbacks.setMuted(true)
          try { await video.play(); return current() } catch (_) {}
        }
        if (current()) { this.desiredPlaying = false; this.callbacks.setPlaying(false) }
        return false
      }
    })()
    this.playPending = pending
    void pending.finally(() => { if (this.playPending === pending) this.playPending = undefined })
    return pending
  }

  pause() {
    this.desiredPlaying = false
    this.playSequence++
    this.playPending = undefined
    this.video?.pause()
    this.callbacks.setPlaying(false)
  }

  retry() { this.start() }

  stop() {
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

  private current(generation: number) { return !this.stopped && this.generation === generation }

  private later(callback: () => void, ms: number) {
    const generation = this.generation
    const timer = window.setTimeout(() => {
      this.timers.delete(timer)
      if (this.current(generation)) callback()
    }, ms)
    this.timers.add(timer)
    return timer
  }

  private async request(url: string, init: RequestInit = {}, timeout = 1000, playerRequest = false) {
    const controller = new AbortController()
    this.requests.add(controller)
    if (playerRequest) this.playerRequests.add(controller)
    const timer = window.setTimeout(() => controller.abort(), timeout)
    this.timers.add(timer)
    try {
      // Consume the body under the same deadline, so headers alone cannot strand cleanup.
      const response = await fetch(url, { ...init, credentials: 'same-origin', cache: 'no-store',
        redirect: 'error', referrerPolicy: 'no-referrer', signal: controller.signal })
      const maximum = init.method === 'POST' ? 2048 : 65536
      if (Number(response.headers.get('Content-Length')) > maximum) throw new Error('bounded response')
      const reader = typeof TextDecoder === 'function' ? response.body?.getReader() : undefined
      let body = ''
      if (reader && typeof TextDecoder === 'function') {
        const decoder = new TextDecoder()
        let size = 0
        for (;;) {
          const { done, value } = await reader.read()
          if (done) break
          size += value.byteLength
          if (size > maximum) { controller.abort(); throw new Error('bounded response') }
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

  private watchLease() {
    if (this.stopped || !this.master || this.statusRunning) return
    this.statusRunning = true
    const generation = this.generation
    void (async () => {
      try {
        if (!this.callbacks.eventSocketOpen()) { this.fail('HLS session ended'); return }
        if (Date.now() >= this.keepAliveAt) {
          const alive = await this.request(new URL('keepalive', new URL('.', this.master)).href, { method: 'POST' })
          if (!this.current(generation)) return
          if (terminalHLSStatus(alive.status)) { this.fail('HLS lease ended'); return }
          if (!alive.ok) throw new Error('keepalive')
          this.keepAliveAt = Date.now() + 15000
        }
        // Native playback hides its HTTP errors. This small, nonblocking master
        // probe also clears buffered content after lease revocation/expiry.
        const response = await this.request(this.master)
        if (!this.current(generation)) return
        if (terminalHLSStatus(response.status)) { this.fail('HLS authorization or media availability ended'); return }
        if (response.status === 503) {
          // A valid not-ready response is not an HTTP connection failure.
          this.statusFailures = 0
          this.clearPlayer()
          this.callbacks.setStatus(this.privatePaused ? 'paused' : 'connecting', this.privatePaused ? 'Private mode' : 'Waiting for fresh media')
          if (!this.privatePaused && ++this.readinessFailures >= 30) this.fail('HLS media did not become ready')
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
          // A live-edge seek can advance currentTime while the buffer/frame
          // remains frozen. Only forward playback with current data counts.
          if (!video.seeking && video.readyState >= 2 && video.currentTime > this.lastTime) this.lastProgress = Date.now()
          this.lastTime = video.currentTime
          if (Date.now() - this.lastProgress > 20000) this.fail('HLS playback stalled; retry manually')
        }
      } catch (_) {
        if (this.current(generation) && ++this.statusFailures >= 3) this.fail('HLS HTTP connection failed; retry manually')
      } finally {
        if (this.current(generation)) { this.statusRunning = false; this.later(() => this.watchLease(), 1000) }
      }
    })()
  }

  private async openPlayer() {
    const video = this.video
    if (!video || this.stopped || this.privatePaused || !this.master || this.attaching) return
    this.attaching = true
    const generation = this.generation
    const playerGeneration = ++this.playerGeneration
    const current = () => this.current(generation) && this.playerGeneration === playerGeneration && this.video === video && !this.privatePaused
    try {
      const native = !!video.canPlayType('application/vnd.apple.mpegurl') || !!video.canPlayType('application/x-mpegURL')
      const nativeCodec = !!video.canPlayType(HLS_VIDEO_TYPE) && !!video.canPlayType(HLS_AUDIO_TYPE)
      const apple = /iP(?:hone|ad|od)|Macintosh/.test(navigator.userAgent) && /Apple/.test(navigator.vendor)
      let kind: 'native' | 'mse' = 'native'
      let module: typeof import('./player') | undefined
      if (!apple || !native) {
        const mediaSource = window.MediaSource || (window as any).ManagedMediaSource
        const mseCodec = typeof mediaSource?.isTypeSupported === 'function' &&
          mediaSource.isTypeSupported(HLS_VIDEO_TYPE) && mediaSource.isTypeSupported(HLS_AUDIO_TYPE)
        if (mseCodec) {
          module = await import('./player')
          if (!current()) return
          if (module.mseSupported()) kind = 'mse'
        }
        if (kind !== 'mse' && (!native || !nativeCodec)) { this.fail('HLS H.264/AAC playback is unsupported in this browser'); return }
      }
      if (!current()) return
      if (kind === 'native') {
        // Native loaders cannot be intercepted. Validate the server-generated
        // master and every initially advertised child before handing it a URL.
        for (const url of hlsMasterChildren(this.masterText, this.master, this.mode)) {
          if (!current()) return
          const response = await this.request(url, {}, 5000, true)
          if (!current()) return
          if (!response.ok || !validHLSPlaylist(await response.text(), url, this.master, this.mode)) throw new Error('native playlist')
        }
      }
      if (!current()) return
      const listen = (event: string, callback: () => void) => {
        const handler = () => { if (current()) callback() }
        video.addEventListener(event, handler)
        this.listeners.push(() => video.removeEventListener(event, handler))
      }
      listen('canplay', () => {
        this.clearReadinessTimer()
        if (!this.playbackReady) this.resetProgress(video)
        this.playbackReady = true
        this.callbacks.setPlayable(true)
        this.callbacks.resolution(video.videoWidth, video.videoHeight)
        if (this.desiredPlaying && video.paused) void this.play()
      })
      listen('resize', () => this.callbacks.resolution(video.videoWidth, video.videoHeight))
      listen('playing', () => {
        if (video.paused) return // An already queued event cannot undo Pause.
        this.clearReadinessTimer()
        this.playbackReady = true
        this.resetProgress(video)
        this.callbacks.setPlayable(true)
        this.callbacks.setPlaying(true)
        this.callbacks.setStatus('streaming')
      })
      listen('pause', () => this.callbacks.setPlaying(false))
      // Also exclude a seek that completes between HTTP watchdog samples.
      listen('seeking', () => { this.lastTime = video.currentTime })
      listen('error', () => this.fail('HLS media playback failed; retry manually'))
      listen('ended', () => this.fail('HLS playback ended; retry manually'))
      video.muted = this.muted
      try { video.volume = this.volume } catch (_) {}
      this.resetProgress(video)
      this.callbacks.setPlayer(kind)
      this.callbacks.setStatus('connecting')
      // This deadline covers initial readiness only. Arm before attachment so
      // even an immediate canplay/playing event can cancel it; later buffering
      // is covered by the separate playback-progress watchdog.
      this.readinessTimer = this.later(() => {
        if (!current()) return
        this.readinessTimer = undefined
        if (video.readyState < 3) this.fail('HLS playback did not become ready; retry manually')
        else { this.playbackReady = true; this.resetProgress(video) }
      }, 30000)
      if (kind === 'mse') {
        const player = module!.createMSEPlayer(video, this.master, this.mode, (detail) => { if (current()) this.fail(detail) })
        if (!current()) { player.destroy(); return }
        this.player = player
      } else { this.player = { destroy() {} }; video.src = this.master; video.load() }
      // Request autoplay once the source/player is attached, even before
      // canplay. play() can wait for data; a first frame is not readiness.
      // Synchronous attachment events and pending Play still share one attempt.
      if (current() && this.desiredPlaying && video.paused) void this.play()
    } catch (_) { if (current()) this.fail('HLS player could not start; retry manually') }
    finally { if (this.playerGeneration === playerGeneration) this.attaching = false }
  }

  private resetProgress(video: HTMLVideoElement) {
    this.lastTime = video.currentTime
    this.lastProgress = Date.now()
  }

  private clearReadinessTimer() {
    if (this.readinessTimer !== undefined) {
      window.clearTimeout(this.readinessTimer)
      this.timers.delete(this.readinessTimer)
      this.readinessTimer = undefined
    }
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
    this.callbacks.setPlayable(false)
    this.callbacks.setPlayer('')
  }

  private fail(detail: string) {
    this.stop()
    this.callbacks.setStatus('terminal', detail)
  }
}
