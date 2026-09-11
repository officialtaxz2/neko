import assert from 'node:assert/strict'
import test from 'node:test'

import { shouldCreateClientOffer } from '../src/neko/negotiation.js'

test('client offer requires the established current peer and an open socket', () => {
  const ready = {
    currentPeer: true,
    socketOpen: true,
    hasRemoteDescription: true,
    signalingState: 'stable',
    makingOffer: false,
  }

  assert.equal(shouldCreateClientOffer(ready), true)
  assert.equal(shouldCreateClientOffer({ ...ready, currentPeer: false }), false)
  assert.equal(shouldCreateClientOffer({ ...ready, socketOpen: false }), false)
  assert.equal(shouldCreateClientOffer({ ...ready, hasRemoteDescription: false }), false)
})

test('client offer is serialized and waits for stable signaling', () => {
  const ready = {
    currentPeer: true,
    socketOpen: true,
    hasRemoteDescription: true,
    signalingState: 'stable',
    makingOffer: false,
  }

  assert.equal(shouldCreateClientOffer({ ...ready, signalingState: 'have-local-offer' }), false)
  assert.equal(shouldCreateClientOffer({ ...ready, signalingState: 'have-remote-offer' }), false)
  assert.equal(shouldCreateClientOffer({ ...ready, makingOffer: true }), false)
})
