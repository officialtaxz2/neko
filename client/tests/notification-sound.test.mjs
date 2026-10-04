import assert from 'node:assert/strict'
import test from 'node:test'
import { playNotificationSound } from '../src/utils/notification-sound.js'

test('optional notification audio tolerates legacy, unsupported and failing media', async () => {
  const original = Object.getOwnPropertyDescriptor(globalThis, 'Audio')
  const attempts = []
  try {
    delete globalThis.Audio
    assert.doesNotThrow(playNotificationSound)

    globalThis.Audio = class {
      constructor(source) { attempts.push(source) }
      play() { return undefined }
    }
    assert.doesNotThrow(playNotificationSound)
    assert.deepEqual(attempts, ['chat.mp3'])

    globalThis.Audio = class { constructor() { throw new Error('unavailable') } }
    assert.doesNotThrow(playNotificationSound)
    globalThis.Audio = class { play() { throw new Error('blocked') } }
    assert.doesNotThrow(playNotificationSound)
    globalThis.Audio = class { play() { return Promise.reject(new Error('autoplay')) } }
    assert.doesNotThrow(playNotificationSound)
    await new Promise((resolve) => setImmediate(resolve))

    globalThis.Audio = class { play() { return Promise.resolve() } }
    assert.doesNotThrow(playNotificationSound)
    await Promise.resolve()
  } finally {
    if (original) Object.defineProperty(globalThis, 'Audio', original)
    else delete globalThis.Audio
  }
})
