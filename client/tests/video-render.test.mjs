import assert from 'node:assert/strict'
import test from 'node:test'
import { createVideo, findClass } from './video-fixture.mjs'

// Renders the real video.vue template once, before any ref exists, the way
// the first page render does. Settings come from the stored values.
function firstRender(settings, options) {
  const { vm, errors, destroy } = createVideo(settings, options)
  try {
    return { vnode: vm._render(), errors }
  } finally {
    destroy()
  }
}

test('a stored trackpad mode renders the player on a touch device', () => {
  const { vnode, errors } = firstRender({ trackpad_mode:true })
  assert.deepEqual(errors, [])
  assert.equal(vnode.isComment, false)
  assert.equal(vnode.tag, 'div')
  assert.ok(findClass(vnode, 'trackpad-cursor'))
})

test('the debug overlay is part of the first render', () => {
  const { vnode, errors } = firstRender({ trackpad_mode:true }, { search:'?debug=1' })
  assert.deepEqual(errors, [])
  assert.ok(findClass(vnode, 'video-debug'))
})
