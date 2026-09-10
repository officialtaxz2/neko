import assert from 'node:assert/strict'
import test from 'node:test'

import { RECONNECT_DELAYS_MS, reconnectDelayForAttempt, shouldReconnect } from '../src/neko/recovery.js'

test('reconnect is limited to a previously connected real session', () => {
  const eligible = {
    previouslyConnected: true,
    credentialsAvailable: true,
    supported: true,
    demo: false,
    suppressed: false,
  }

  assert.equal(shouldReconnect(eligible), true)
  assert.equal(shouldReconnect({ ...eligible, previouslyConnected: false }), false)
  assert.equal(shouldReconnect({ ...eligible, credentialsAvailable: false }), false)
  assert.equal(shouldReconnect({ ...eligible, supported: false }), false)
  assert.equal(shouldReconnect({ ...eligible, demo: true }), false)
  assert.equal(shouldReconnect({ ...eligible, suppressed: true }), false)
})

test('reconnect schedule is ordered and bounded', () => {
  assert.deepEqual(RECONNECT_DELAYS_MS, [1000, 2000, 5000, 10000])

  RECONNECT_DELAYS_MS.forEach((delay, attempt) => {
    assert.equal(reconnectDelayForAttempt(attempt), delay)
  })

  assert.equal(reconnectDelayForAttempt(-1), null)
  assert.equal(reconnectDelayForAttempt(RECONNECT_DELAYS_MS.length), null)
  assert.equal(reconnectDelayForAttempt(0.5), null)
})
