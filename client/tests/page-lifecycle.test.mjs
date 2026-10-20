import assert from 'node:assert/strict'
import test from 'node:test'
import { bindPageLifecycle } from '../src/neko/page-lifecycle.js'

const pageEvent = (type, persisted) => Object.assign(new Event(type), { persisted })

test('a page entering or leaving the back/forward cache closes or restarts the client', () => {
  const target = new EventTarget(), calls = []
  const unbind = bindPageLifecycle(target, { leave: () => calls.push('leave'), restore: () => calls.push('restore') })
  target.dispatchEvent(pageEvent('pagehide', false))
  target.dispatchEvent(pageEvent('pageshow', false))
  assert.deepEqual(calls, [], 'an ordinary unload or first load needs no action')
  target.dispatchEvent(pageEvent('pagehide', true))
  target.dispatchEvent(pageEvent('pageshow', true))
  assert.deepEqual(calls, ['leave', 'restore'])
  unbind()
  target.dispatchEvent(pageEvent('pagehide', true))
  assert.deepEqual(calls, ['leave', 'restore'])
})
