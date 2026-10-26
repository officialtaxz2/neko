// The fork: `?debug` (e.g. `&debug=1`) shows a few lines of text over the picture for devices without
// a reachable console (iPhone, the TV): the stream's path and connection, the video element, the
// playback watchdog, input counters, the stored settings and the last page errors. It only reads;
// without the parameter nothing changes. Ported from the Vue client's video-debug.js
// (docs/IOS_RECOVERY.md). Never shows a session token or a password.
import { HLSTransport, WebRTCTransport, samplePlayback } from '@m1k1o/neko'
import type { NekoApp } from './app'
import { mediaStore } from './media'

export const debugEnabled = (search: string = location.search) => new URLSearchParams(search).has('debug')

// the last few page errors, kept from start-up on
const errors: string[] = []

export function noteError(message: unknown) {
  errors.push(String(message).slice(0, 160))
  if (errors.length > 3) errors.shift()
}

export const pageErrors = (): readonly string[] => errors

let capturing = false

export function captureErrors(target: Pick<EventTarget, 'addEventListener'> = globalThis) {
  if (capturing) return
  capturing = true
  target.addEventListener('error', (e) => noteError((e as ErrorEvent).message || (e as ErrorEvent).error))
  target.addEventListener('unhandledrejection', (e) => noteError(`unhandled: ${(e as PromiseRejectionEvent).reason}`))
}

// keys whose values are secrets: the member's session token, an old stored password
const SECRET = /session|token|password/i

// the stored page settings, without secrets, values cut short
export function storedSettings(storage: Storage | undefined = globalThis.localStorage) {
  try {
    if (!storage) return '-'
    const entries: string[] = []
    for (let i = 0; i < storage.length; i++) {
      const key = storage.key(i)
      if (key === null || SECRET.test(key)) continue
      entries.push(`${key}=${String(storage.getItem(key)).slice(0, 24)}`)
    }
    return entries.sort().join(' ') || 'empty'
  } catch {
    return 'unreadable'
  }
}

// what the overlay counts itself: touches that reached the page, when the current stream was
// put on the element and showed its first frame, and the WebRTC quality tier the server last
// selected (signal/video; with adaptive quality it changes without a new stream)
export interface DebugCounters {
  touches: number
  lastTouch: string
  streamAt: number
  firstFrameAt: number
  video?: { id: string; auto: boolean }
}

const READY = ['nothing', 'metadata', 'current', 'future', 'enough']
const round = (v: number) => (Number.isFinite(v) ? Math.round(v * 10) / 10 : v)
const seconds = (from: number, to: number) => (from && to ? `${((to - from) / 1000).toFixed(1)}s` : '-')

export function describe(neko: NekoApp, counters: DebugCounters, now = Date.now()): string {
  const { client, app } = neko
  const transport = client.transport
  const state = client.state
  const media = mediaStore(neko).getState()
  const lines = [
    `media ${media.selection.backend}${media.selection.overridden ? ' (url)' : ''} connection ${state.connection.status}`,
    // the page errors first: a short picture (a phone upright) cuts the text at its end
    ...errors.map((e) => `error ${e}`),
  ]

  if (transport instanceof WebRTCTransport) {
    const pc = transport.peer
    const { attempts, framesTrusted } = transport.playback
    lines.push(
      `peer ${pc?.connectionState ?? '-'} ice ${pc?.iceConnectionState ?? '-'} channel ${transport.input.channelState}`,
      `recovery ${attempts} frames trusted ${framesTrusted}`,
    )
  } else if (transport instanceof HLSTransport && media.hls) {
    lines.push(`hls ${media.hls.mode} ${media.hls.status}${media.hls.detail ? ` ${media.hls.detail}` : ''}`)
  }

  if (counters.video) lines.push(`quality ${counters.video.id || '-'} ${counters.video.auto ? 'auto' : 'fixed'}`)

  const video = transport.element as HTMLVideoElement | null
  if (!video) lines.push('video none')
  else {
    const sample = samplePlayback(video)
    const source = video.srcObject as MediaStream | null
    const tracks =
      source && typeof source.getTracks === 'function'
        ? source
            .getTracks()
            .map((t) => `${t.kind}:${t.readyState}${t.muted ? ':muted' : ''}${t.enabled ? '' : ':disabled'}`)
            .join(' ')
        : 'none'
    const decoded = (video as any).webkitDecodedFrameCount
    lines.push(
      `playing ${state.video.playing} paused ${video.paused} ready ${READY[video.readyState] ?? video.readyState} time ${round(video.currentTime)}`,
      `size ${video.videoWidth}x${video.videoHeight} muted ${video.muted} vol ${round(video.volume)}`,
      `frames ${sample?.frames ?? '-'} decoded ${Number.isFinite(decoded) ? decoded : '-'}`,
      `src ${source ? 'stream' : video.currentSrc ? 'url' : 'none'} tracks ${tracks}`,
    )
  }

  const input = transport instanceof WebRTCTransport ? transport.input.stats : null
  lines.push(
    `stream since ${seconds(counters.streamAt, now)} first frame after ${seconds(counters.streamAt, counters.firstFrameAt)}`,
    `input hosting ${client.controlling} locked ${state.control.locked} touch ${client.isTouchDevice} trackpad ${app.getState().settings.trackpad_mode}`,
    `touches ${counters.touches} ${counters.lastTouch || '-'} sent ${input?.sent ?? '-'} skipped ${input?.skipped ?? '-'}`,
    `visible ${typeof document === 'undefined' ? '-' : document.visibilityState}`,
    `stored ${storedSettings()}`,
  )
  return lines.join('\n')
}
