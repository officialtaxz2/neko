import assert from 'node:assert/strict'
import test from 'node:test'
import { samplePlayback, hasPlaybackProgress } from '../src/neko/playback-progress.js'

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
