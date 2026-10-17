// Read only media state, with no decoder-specific requirement for old browsers.
export function samplePlayback(video) {
  if (!video) return null
  let frames = null
  try {
    const quality = video.getVideoPlaybackQuality?.()
    if (Number.isFinite(quality?.totalVideoFrames)) {
      frames = quality.totalVideoFrames - (Number.isFinite(quality.droppedVideoFrames) ? quality.droppedVideoFrames : 0)
    }
    else if (Number.isFinite(video.webkitDecodedFrameCount)) frames = video.webkitDecodedFrameCount
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
  // WebKit on iPad does not count MediaStream frames; callers trust the
  // counter only after it has advanced once for the current stream.
  if (framesTrusted && current.frames !== null && previous.frames !== null) return current.frames > previous.frames
  return Number.isFinite(current.time) && Number.isFinite(previous.time) && current.time > previous.time + 0.001
}

// True once a frame counter has visibly advanced between two samples.
export function framesAdvanced(previous, current) {
  return !!previous && !!current && previous.frames !== null && current.frames !== null && current.frames > previous.frames
}
