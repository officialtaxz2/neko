// Playback progress of a <video>, and a watchdog that notices a stream which stopped moving while it
// should play: no new frames (or, without frame counters, no time progress) for STALL_MS in the
// foreground. It then re-attaches the stream (srcObject) up to MAX_ATTEMPTS times, never with
// video.load(); after that playback is stopped and an explicit Play starts it again. Background tabs,
// a deliberate pause and a running re-attachment are not stalls.

export const STALL_MS = 8_000
export const POLL_MS = 500
export const MAX_ATTEMPTS = 3

export interface PlaybackSample {
  time: number
  frames: number | null
  paused: boolean
  seeking: boolean
  ready: boolean
}

const presented = new WeakMap<HTMLVideoElement, number>()

// requestVideoFrameCallback counts frames the page actually presented, also on iPad WebKit, whose
// MediaStream decode counters stay at zero
export function trackPresentedFrames(video: HTMLVideoElement): () => void {
  if (typeof video?.requestVideoFrameCallback !== 'function') return () => {}
  let active = true
  let handle = 0
  presented.set(video, 0)
  const tick = () => {
    if (!active) return
    presented.set(video, (presented.get(video) ?? 0) + 1)
    handle = video.requestVideoFrameCallback(tick)
  }
  handle = video.requestVideoFrameCallback(tick)
  return () => {
    active = false
    presented.delete(video)
    try {
      video.cancelVideoFrameCallback?.(handle)
    } catch {}
  }
}

// native fullscreen and Picture-in-Picture can render outside the page, where neither frame
// callbacks nor counters are reliable; playback time remains
function renderedOutsidePage(video: any) {
  return (
    !!video.webkitDisplayingFullscreen ||
    (typeof video.webkitPresentationMode === 'string' && video.webkitPresentationMode !== 'inline') ||
    (typeof document !== 'undefined' && (document as any).pictureInPictureElement === video)
  )
}

export function samplePlayback(video: HTMLVideoElement | null): PlaybackSample | null {
  if (!video) return null
  let frames: number | null = null
  try {
    if (renderedOutsidePage(video)) frames = null
    else if (presented.has(video)) frames = presented.get(video)!
    else {
      const quality = video.getVideoPlaybackQuality?.()
      if (quality && Number.isFinite(quality.totalVideoFrames))
        frames =
          quality.totalVideoFrames - (Number.isFinite(quality.droppedVideoFrames) ? quality.droppedVideoFrames : 0)
      else if (Number.isFinite((video as any).webkitDecodedFrameCount)) frames = (video as any).webkitDecodedFrameCount
    }
  } catch {} // time progress remains the fallback
  return {
    time: video.currentTime,
    frames,
    paused: video.paused,
    seeking: video.seeking,
    ready: video.readyState >= 2,
  }
}

// A seek, metadata or a repeated playing/timeupdate event is no progress. Where frame counters work,
// time (audio) advancing alone is not enough; a counter is trusted only once it advanced.
export function hasPlaybackProgress(
  previous: PlaybackSample | null,
  current: PlaybackSample | null,
  framesTrusted = true,
) {
  if (!previous || !current || current.paused || current.seeking || previous.seeking || !current.ready) return false
  if (framesTrusted && current.frames !== null && previous.frames !== null) return current.frames > previous.frames
  return Number.isFinite(current.time) && Number.isFinite(previous.time) && current.time > previous.time + 0.001
}

export function framesAdvanced(previous: PlaybackSample | null, current: PlaybackSample | null) {
  return (
    !!previous && !!current && previous.frames !== null && current.frames !== null && current.frames > previous.frames
  )
}

export interface WatchdogHooks {
  // put the stream on the element again and start playing it
  reattach(): void
  // give up: stop playback so that an explicit Play is needed
  giveUp(): void
}

export class PlaybackWatchdog {
  private video: HTMLVideoElement | null = null
  private timer = 0
  private sample: PlaybackSample | null = null
  private lastProgress = 0
  private framesTrusted = false
  private hidden = false
  private stopFrames: () => void = () => {}
  attempts = 0
  private readonly hooks: WatchdogHooks

  constructor(hooks: WatchdogHooks) {
    this.hooks = hooks
  }

  // watch this element while it should play; a new stream starts the attempts over
  start(video: HTMLVideoElement) {
    if (this.video !== video) {
      this.stop()
      this.video = video
      this.stopFrames = trackPresentedFrames(video)
    }
    if (this.timer) return
    this.lastProgress = Date.now()
    this.sample = samplePlayback(video)
    this.timer = window.setTimeout(this.poll, POLL_MS)
  }

  stop() {
    clearTimeout(this.timer)
    this.timer = 0
    this.stopFrames()
    this.stopFrames = () => {}
    this.video = null
    this.sample = null
    this.framesTrusted = false
  }

  // a new stream on the element, or the user started playback: earlier attempts do not count
  resetAttempts() {
    this.attempts = 0
    this.lastProgress = Date.now()
  }

  get running() {
    return this.timer !== 0
  }

  // the frame counter is trusted (it advanced for this stream); else playback time decides
  get trusted() {
    return this.framesTrusted
  }

  private observe() {
    const current = samplePlayback(this.video)
    if (framesAdvanced(this.sample, current)) this.framesTrusted = true
    if (hasPlaybackProgress(this.sample, current, this.framesTrusted)) {
      this.lastProgress = Date.now()
      this.attempts = 0
    }
    this.sample = current
  }

  private poll = () => {
    this.timer = 0
    if (!this.video) return
    // background throttling is no evidence of a broken foreground decoder
    if (document.hidden || this.hidden) {
      this.hidden = !!document.hidden
      this.lastProgress = Date.now()
      this.sample = samplePlayback(this.video)
    } else {
      this.observe()
      if (Date.now() - this.lastProgress >= STALL_MS) {
        this.lastProgress = Date.now()
        if (this.attempts >= MAX_ATTEMPTS) {
          console.error(`[neko] stream recovery failed after ${MAX_ATTEMPTS} attempts`)
          this.stop()
          this.hooks.giveUp()
          return
        }
        this.attempts++
        console.warn(`[neko] no playback progress, re-attaching the stream (#${this.attempts})`)
        this.hooks.reattach()
        this.sample = samplePlayback(this.video)
      }
    }
    if (this.video && !this.timer) this.timer = window.setTimeout(this.poll, POLL_MS)
  }
}
