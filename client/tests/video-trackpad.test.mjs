import assert from 'node:assert/strict'
import test from 'node:test'
import Vue from 'vue'
import { createVideo } from './video-fixture.mjs'

const touch = (type, identifier) => ({ type, changedTouches:[{ identifier, clientX:100, clientY:100 }] })

function trackpad() {
  const video = createVideo({ trackpad_mode:true }, { remote:{ hosting:true, implicitHosting:false, controlling:true } })
  // Taps size the cursor from the video element; the instance is not mounted.
  video.vm.$refs.video = { getBoundingClientRect: () => ({ width:1280, height:720, left:0, top:0 }), videoWidth:1280, videoHeight:720 }
  return video
}

function tap(vm, identifier) {
  vm.onTrackpadTouch(touch('touchstart', identifier))
  vm.onTrackpadTouch(touch('touchend', identifier))
}

test('a trackpad tap clicks while controlling', () => {
  const { vm, sent, errors, destroy } = trackpad()
  tap(vm, 1)
  assert.ok(sent.includes('mousedown'))
  assert.deepEqual(errors, [])
  destroy()
})

for (const [name, lose, regain] of [
  ['losing control', (remote) => { remote.hosting = false; remote.controlling = false }, (remote) => { remote.hosting = true; remote.controlling = true }],
  ['pausing input', (remote) => { remote.locked = true }, (remote) => { remote.locked = false }],
]) {
  test(`${name} during a trackpad touch does not block later touches`, async () => {
    const { vm, accessor, sent, errors, destroy } = trackpad()
    vm.onTrackpadTouch(touch('touchstart', 1))
    lose(accessor.remote)
    await Vue.nextTick()
    // Without control the end of the touch is ignored.
    vm.onTrackpadTouch(touch('touchend', 1))
    regain(accessor.remote)
    await Vue.nextTick()
    sent.length = 0
    tap(vm, 2)
    assert.equal(vm.trackpadTouchId, null)
    assert.ok(sent.includes('mousedown'), `sent: ${sent}`)
    assert.deepEqual(errors, [])
    destroy()
  })
}
