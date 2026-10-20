import { samplePlayback } from './playback-progress.js'

// Opt-in (?debug) text describing the video element, for devices without a
// reachable console such as iPhone browsers. Reads state only.
export function debugEnabled(search) {
  return new URLSearchParams(typeof search === 'string' ? search : '').has('debug')
}

const READY = ['nothing', 'metadata', 'current', 'future', 'enough']

function round(value) {
  return Number.isFinite(value) ? Math.round(value * 10) / 10 : value
}

export function describeVideo(video, state = {}) {
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
  ].join('\n')
}
