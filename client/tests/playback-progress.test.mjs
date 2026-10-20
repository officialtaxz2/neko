import assert from 'node:assert/strict'
import test from 'node:test'
import { samplePlayback, hasPlaybackProgress, framesAdvanced, trackPresentedFrames } from '../src/neko/playback-progress.js'

test('metadata, paused and seeking samples cannot masquerade as playback', () => {
  const video = { currentTime:0, paused:false, seeking:false, readyState:0 }
  let before = samplePlayback(video)
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), false)
  video.readyState = 2
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), false)
  before = samplePlayback(video)
  video.currentTime = 1
  video.paused = true
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), false)
  video.paused = false
  video.seeking = true
  const seeking = samplePlayback(video)
  video.currentTime = 10
  video.seeking = false
  assert.equal(hasPlaybackProgress(seeking, samplePlayback(video)), false)
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), true)
})

test('available frame evidence rejects audio-only advance and dropped frames', () => {
  let total = 5, dropped = 0
  const video = { currentTime:1, paused:false, seeking:false, readyState:2,
    getVideoPlaybackQuality: () => ({ totalVideoFrames:total, droppedVideoFrames:dropped }) }
  const before = samplePlayback(video)
  video.currentTime = 2
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), false)
  total++; dropped++
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), false)
  total++
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), true)
})

test('old browsers retain time progress without requiring frame APIs', () => {
  const video = { currentTime:1, paused:false, seeking:false, readyState:2,
    getVideoPlaybackQuality: () => { throw new Error('unsupported') } }
  const before = samplePlayback(video)
  video.currentTime += 0.25
  assert.equal(hasPlaybackProgress(before, samplePlayback(video)), true)
})

test('an untrusted frame counter falls back to time progress', () => {
  // iPad WebKit leaves MediaStream frame counters unchanged while playing.
  const video = { currentTime:1, paused:false, seeking:false, readyState:4,
    getVideoPlaybackQuality: () => ({ totalVideoFrames:0, droppedVideoFrames:0 }) }
  const before = samplePlayback(video)
  video.currentTime = 1.5
  const after = samplePlayback(video)
  assert.equal(framesAdvanced(before, after), false)
  assert.equal(hasPlaybackProgress(before, after, false), true)
  assert.equal(hasPlaybackProgress(before, after, true), false)
})

test('presented-frame callbacks detect a frozen iPad picture whose decode counter stays at zero', () => {
  let callback = null, cancelled = null
  const video = { currentTime:1, paused:false, seeking:false, readyState:2,
    getVideoPlaybackQuality: () => ({ totalVideoFrames:0, droppedVideoFrames:0 }),
    requestVideoFrameCallback(next) { callback = next; return 7 },
    cancelVideoFrameCallback(handle) { cancelled = handle } }
  const present = () => { const next = callback; callback = null; next(0, {}) }
  const stop = trackPresentedFrames(video)
  let before = samplePlayback(video)
  present(); video.currentTime = 1.5
  let current = samplePlayback(video)
  assert.equal(framesAdvanced(before, current), true)
  assert.equal(hasPlaybackProgress(before, current, true), true)
  // Audio and time keep advancing while no frame reaches the screen.
  before = current
  video.currentTime = 2.5
  assert.equal(hasPlaybackProgress(before, samplePlayback(video), true), false)
  // Native fullscreen renders outside the page; fall back to time.
  video.webkitDisplayingFullscreen = true
  before = samplePlayback(video)
  video.currentTime = 3
  assert.equal(before.frames, null)
  assert.equal(hasPlaybackProgress(before, samplePlayback(video), true), true)
  stop()
  assert.equal(cancelled, 7)
  video.webkitDisplayingFullscreen = false
  assert.equal(samplePlayback(video).frames, 0, 'stopped tracking returns to the decode counter')
})

test('browsers without presented-frame callbacks keep their decode counters', () => {
  const video = { currentTime:1, paused:false, seeking:false, readyState:2,
    getVideoPlaybackQuality: () => ({ totalVideoFrames:5, droppedVideoFrames:1 }) }
  trackPresentedFrames(video)()
  assert.equal(samplePlayback(video).frames, 4)
})
