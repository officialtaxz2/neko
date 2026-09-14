import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import {
  MEDIA_RETRY_DELAYS_MS,
  MEDIA_RETRY_STABILITY_MS,
  deliverOrReleaseVideoFrame,
  isMediaRetryCloseCode,
  mediaRetryDelayForAttempt,
  shouldAwaitMediaCloseForEnd,
  shouldDropDecodedVideoOutput,
} from '../src/neko/media/recovery.js'

test('media-only retry is limited to backpressure and backend failure closes', () => {
  assert.equal(isMediaRetryCloseCode(4413), true)
  assert.equal(isMediaRetryCloseCode(4500), true)
  for (const code of [1000, 1001, 1002, 4400, 4401, 4403, 4408, 4409, 4429]) {
    assert.equal(isMediaRetryCloseCode(code), false)
  }
})

test('media-only retry uses four serialized bounded delays', () => {
  assert.deepEqual(MEDIA_RETRY_DELAYS_MS, [1000, 2000, 5000, 10000])
  assert.equal(MEDIA_RETRY_STABILITY_MS, 30_000)
  MEDIA_RETRY_DELAYS_MS.forEach((delay, attempt) => assert.equal(mediaRetryDelayForAttempt(attempt), delay))
  assert.equal(mediaRetryDelayForAttempt(MEDIA_RETRY_DELAYS_MS.length), null)
})

test('backend-error END waits for the retryable private close code', () => {
  assert.equal(shouldAwaitMediaCloseForEnd('backend_error'), true)
  for (const reason of ['normal', 'revoked', 'replaced', 'shutdown']) {
    assert.equal(shouldAwaitMediaCloseForEnd(reason), false)
  }
})

test('decoded video saturation drops output at the fixed renderer bound', () => {
  assert.equal(shouldDropDecodedVideoOutput(0, 2), false)
  assert.equal(shouldDropDecodedVideoOutput(1, 2), false)
  assert.equal(shouldDropDecodedVideoOutput(2, 2), true)
  assert.equal(shouldDropDecodedVideoOutput(3, 2), true)
})

test('video frames without a mounted renderer listener are released', () => {
  let releases = 0
  assert.equal(deliverOrReleaseVideoFrame(() => false, () => releases++), false)
  assert.equal(releases, 1)

  assert.equal(deliverOrReleaseVideoFrame(() => true, () => releases++), true)
  assert.equal(releases, 1)

  assert.throws(
    () => deliverOrReleaseVideoFrame(() => { throw new Error('listener failed') }, () => releases++),
    /listener failed/,
  )
  assert.equal(releases, 2)
})

test('Vue WebCodecs callbacks keep live component state', async () => {
  const source = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')

  for (const method of ['onWebCodecsFrame', 'onWebCodecsClockReset', 'renderWebCodecsFrame']) {
    assert.match(source, new RegExp(`private ${method}\\([^)]*\\) \\{`))
    assert.doesNotMatch(source, new RegExp(`private ${method}\\s*=\\s*\\(`))
  }
})
