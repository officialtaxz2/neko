import { samplePlayback } from './playback-progress.js'

// Opt-in (?debug) text describing the video element, for devices without a
// reachable console such as iPhone browsers. Reads state only.
export function debugEnabled(search) {
  return new URLSearchParams(typeof search === 'string' ? search : '').has('debug')
}

const errors = []

// Keeps the last few page errors for the debug display.
export function noteError(message) {
  errors.push(String(message).slice(0, 160))
  if (errors.length > 3) errors.shift()
}

export function captureErrors(target, vueConfig) {
  target.addEventListener('error', (event) => noteError(event.message || event.error))
  target.addEventListener('unhandledrejection', (event) => noteError(`unhandled: ${event.reason}`))
  if (vueConfig) {
    vueConfig.errorHandler = (err, vm, info) => {
      noteError(`${info}: ${err}`)
      console.error(err)
    }
  }
}

// Stored page settings, without the saved password.
function storedSettings(storage) {
  try {
    if (!storage) return '-'
    const entries = []
    for (let i = 0; i < storage.length; i++) {
      const key = storage.key(i)
      if (key === null || key === 'password') continue
      entries.push(`${key}=${String(storage.getItem(key)).slice(0, 24)}`)
    }
    return entries.sort().join(' ') || 'empty'
  } catch (_) {
    return 'unreadable'
  }
}

const READY = ['nothing', 'metadata', 'current', 'future', 'enough']

function round(value) {
  return Number.isFinite(value) ? Math.round(value * 10) / 10 : value
}

export function describeVideo(video, state = {}, storage = typeof localStorage === 'undefined' ? undefined : localStorage) {
  if (!video) return 'video: none'
  const sample = samplePlayback(video)
  const source = video.srcObject
  const tracks = source && typeof source.getTracks === 'function'
    ? source.getTracks().map((t) => `${t.kind}:${t.readyState}${t.muted ? ':muted' : ''}${t.enabled ? '' : ':disabled'}`).join(' ')
    : 'none'
  const decoded = Number.isFinite(video.webkitDecodedFrameCount) ? video.webkitDecodedFrameCount : '-'
  return [
    `media ${state.backend || '-'} playing ${!!state.playing} play ${state.lastPlay || '-'}`,
    `paused ${video.paused} ready ${READY[video.readyState] || video.readyState} time ${round(video.currentTime)}`,
    `size ${video.videoWidth}x${video.videoHeight} muted ${video.muted} vol ${round(video.volume)}`,
    `frames ${sample && sample.frames !== null ? sample.frames : '-'} decoded ${decoded} trusted ${!!state.framesTrusted} recovery ${state.attempts || 0}`,
    `src ${source ? 'stream' : (video.currentSrc ? 'url' : 'none')} tracks ${tracks}`,
    `visible ${typeof document === 'undefined' ? '-' : document.visibilityState}`,
    ...(state.extra || []),
    `stored ${storedSettings(storage)}`,
    ...errors.map((error) => `error ${error}`),
  ].join('\n')
}
