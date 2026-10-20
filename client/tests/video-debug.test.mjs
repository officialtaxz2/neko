import test from 'node:test'
import assert from 'node:assert/strict'
import { debugEnabled, describeVideo } from '../src/neko/video-debug.js'

test('debug display is opt-in by query', () => {
  assert.equal(debugEnabled('?usr=guest&debug=1'), true)
  assert.equal(debugEnabled('?debug'), true)
  assert.equal(debugEnabled('?usr=guest'), false)
  assert.equal(debugEnabled(undefined), false)
})

test('describes a paused stream element with its tracks', () => {
  const tracks = [
    { kind: 'audio', readyState: 'live', muted: false, enabled: true },
    { kind: 'video', readyState: 'live', muted: true, enabled: true },
  ]
  const video = {
    paused: true, readyState: 0, currentTime: 0, seeking: false, videoWidth: 0, videoHeight: 0,
    muted: false, volume: 1, srcObject: { getTracks: () => tracks }, currentSrc: '',
  }
  const text = describeVideo(video, { backend: 'webrtc', playing: true, lastPlay: 'pending', attempts: 2 })
  assert.match(text, /media webrtc playing true play pending/)
  assert.match(text, /paused true ready nothing time 0/)
  assert.match(text, /size 0x0 muted false/)
  assert.match(text, /recovery 2/)
  assert.match(text, /tracks audio:live video:live:muted/)
})

test('missing element is reported, not thrown', () => {
  assert.equal(describeVideo(undefined), 'video: none')
})
