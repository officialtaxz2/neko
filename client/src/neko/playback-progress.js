const presented = new WeakMap()

// requestVideoFrameCallback reports frames the page actually presented, also
// on iPad WebKit, whose MediaStream decode counters stay at zero.
export function trackPresentedFrames(video) {
  if (typeof video?.requestVideoFrameCallback !== 'function') return () => {}
  let active = true
  let handle
  presented.set(video, 0)
  const tick = () => {
    if (!active) return
    presented.set(video, presented.get(video) + 1)
    handle = video.requestVideoFrameCallback(tick)
  }
  handle = video.requestVideoFrameCallback(tick)
  return () => {
    active = false
    presented.delete(video)
    try { video.cancelVideoFrameCallback?.(handle) } catch (_) { /* already gone */ }
  }
}

// Native fullscreen and Picture-in-Picture can render outside the page, where
// neither frame callbacks nor counters are reliable; playback time remains.
function renderedOutsidePage(video) {
  return !!video.webkitDisplayingFullscreen ||
    (typeof video.webkitPresentationMode === 'string' && video.webkitPresentationMode !== 'inline') ||
    (typeof document !== 'undefined' && document.pictureInPictureElement === video)
}

// Read only media state, with no decoder-specific requirement for old browsers.
export function samplePlayback(video) {
  if (!video) return null
  let frames = null
  try {
    if (renderedOutsidePage(video)) frames = null
    else if (presented.has(video)) frames = presented.get(video)
    else {
      const quality = video.getVideoPlaybackQuality?.()
      if (Number.isFinite(quality?.totalVideoFrames)) {
        frames = quality.totalVideoFrames - (Number.isFinite(quality.droppedVideoFrames) ? quality.droppedVideoFrames : 0)
      }
      else if (Number.isFinite(video.webkitDecodedFrameCount)) frames = video.webkitDecodedFrameCount
    }
  } catch (_) { /* Time progress remains the compatibility fallback. */ }
  return {
    time: video.currentTime, frames, paused: video.paused,
    seeking: video.seeking, ready: video.readyState >= 2,
  }
}

export function hasPlaybackProgress(previous, current, framesTrusted = true) {
  if (!previous || !current || current.paused || current.seeking || previous.seeking || !current.ready) return false
  // A seek, metadata or repeated playing/timeupdate event cannot reset recovery.
  // Where frame counters work, audio/time advancement alone is insufficient.
  // WebKit on iPad does not count decoded MediaStream frames (presented-frame
  // callbacks do); callers trust a counter only after it advanced once.
  if (framesTrusted && current.frames !== null && previous.frames !== null) return current.frames > previous.frames
  return Number.isFinite(current.time) && Number.isFinite(previous.time) && current.time > previous.time + 0.001
}

// True once a frame counter has visibly advanced between two samples.
export function framesAdvanced(previous, current) {
  return !!previous && !!current && previous.frames !== null && current.frames !== null && current.frames > previous.frames
}
