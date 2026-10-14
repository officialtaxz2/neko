import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import { elementRequestFullscreen } from '../src/utils/fullscreen.js'

test('fullscreen rejection remains available to the mobile fallback', async () => {
  let attempts = 0
  assert.equal(await elementRequestFullscreen({
    requestFullscreen: () => {
      attempts++
      return Promise.reject(new Error('unsupported target'))
    },
  }), false)
  assert.equal(attempts, 1)

  assert.equal(await elementRequestFullscreen({
    webkitEnterFullscreen: () => {
      attempts++
    },
  }), true)
  assert.equal(attempts, 2)
  assert.equal(await elementRequestFullscreen({}), false)
})

test('mobile fullscreen fallback mutates the live Vue component', async () => {
  const source = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')

  assert.match(source, /private onFullscreenChangeHandler\(\) \{/)
  assert.doesNotMatch(source, /private onFullscreenChangeHandler\s*=\s*\(/)
  assert.match(source, /if \(await elementRequestFullscreen\(this\._player\)\)/)
  assert.match(source, /if \(await elementRequestFullscreen\(videoSurface\)\)/)
  assert.match(source, /this\.enterFallbackFullscreen\(\)/)
  assert.match(source, /'fallback-fullscreen': fallbackFullscreen/)
})
