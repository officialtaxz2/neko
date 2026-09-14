import assert from 'node:assert/strict'
import test from 'node:test'

import {
  hasVideoDecodeCapacity,
  MEDIA_RETRY_DELAYS_MS,
  isMediaRetryCloseCode,
  mediaRetryDelayForAttempt,
  shouldAwaitMediaCloseForEnd,
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
  MEDIA_RETRY_DELAYS_MS.forEach((delay, attempt) => assert.equal(mediaRetryDelayForAttempt(attempt), delay))
  assert.equal(mediaRetryDelayForAttempt(MEDIA_RETRY_DELAYS_MS.length), null)
})

test('backend-error END waits for the retryable private close code', () => {
  assert.equal(shouldAwaitMediaCloseForEnd('backend_error'), true)
  for (const reason of ['normal', 'revoked', 'replaced', 'shutdown']) {
    assert.equal(shouldAwaitMediaCloseForEnd(reason), false)
  }
})

test('video decode capacity includes outputs pending their callback', () => {
  assert.equal(hasVideoDecodeCapacity(0, 0, 2), true)
  assert.equal(hasVideoDecodeCapacity(0, 1, 2), true)
  assert.equal(hasVideoDecodeCapacity(1, 0, 2), true)
  assert.equal(hasVideoDecodeCapacity(0, 2, 2), false)
  assert.equal(hasVideoDecodeCapacity(1, 1, 2), false)
  assert.equal(hasVideoDecodeCapacity(2, 0, 2), false)
})
