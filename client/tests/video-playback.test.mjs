import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'
import Vue from 'vue'
import * as progress from '../src/neko/playback-progress.js'

// Run on the target. This mounts the actual decorated Vue class with its real
// data/method binding and lifecycle, replacing DOM/media/network with doubles.
// It checks callback ownership, not a browser decoder or TV compatibility.
const require = createRequire(import.meta.url)
const file = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')
const source = file.match(/<script lang="ts">([\s\S]*?)<\/script>/)[1]
const compiled = ts.transpileModule(source, { compilerOptions: {
  target:ts.ScriptTarget.ES2022, module:ts.ModuleKind.CommonJS,
  experimentalDecorators:true, useDefineForClassFields:false, esModuleInterop:true,
} }).outputText
const flush = async () => { await Vue.nextTick(); await new Promise(setImmediate) }

function eventTarget() {
  const handlers = new Map()
  return { handlers,
    addEventListener(event, callback) { if (!handlers.has(event)) handlers.set(event, new Set()); handlers.get(event).add(callback) },
    removeEventListener(event, callback) { handlers.get(event)?.delete(callback) },
    fire(event) { for (const callback of handlers.get(event) || []) callback({ type:event }) },
  }
}

function harness({ autoplay = true, play } = {}) {
  let now = 0, key = 0
  const timers = new Map(), exports = {}
  const window = { ...eventTarget(), innerWidth:1920,
    setTimeout(callback, delay) { timers.set(++key, { at:now + delay, callback }); return key },
    clearTimeout(id) { timers.delete(id) },
    cancelAnimationFrame() {}, matchMedia: () => ({ matches:false }),
    requestAnimationFrame: () => ++key,
    URL:{ createObjectURL: () => 'blob:legacy-stream' },
  }
  const document = { ...eventTarget(), hidden:false, body:{ style:{} }, hasFocus: () => true }
  const video = { ...eventTarget(), srcObject:null, paused:true, seeking:false,
    currentTime:0, readyState:0, videoWidth:640, videoHeight:360, muted:false, volume:1, playCalls:0,
    play() { this.playCalls++; if (play) return play(this); this.paused = false; return undefined },
    pause() { this.paused = true; this.fire('pause') },
  }
  const stream = { ...eventTarget(), getVideoTracks: () => [{}] }
  const state = Vue.observable({ stream, track:undefined, playing:false, playable:false, muted:false, volume:100,
    width:640, height:360, resolution:{ w:640, h:360 } })
  Object.assign(state, { play: () => { state.playing = true }, pause: () => { state.playing = false },
    setPlayable: value => { state.playable = value }, setMuted: value => { state.muted = value }, setVolume: value => { state.volume = value } })
  runInNewContext(compiled, {
    exports, window, document, navigator:{ userAgent:'test', maxTouchPoints:0, platform:'test' },
    Date:{ now: () => now }, console, Object,
    require(path) {
      if (path === 'vue-property-decorator') return require(path)
      if (path === 'resize-observer-polyfill') return class { observe() {} disconnect() {} }
      if (path === '~/utils/guacamole-keyboard.ts') return () => ({ listenTo() {}, reset() {} })
      if (path === '~/neko/playback-progress.js') return progress
      if (path === '~/neko/media-selection.js') return { useCompactMediaStatus: () => true }
      if (path === '~/utils') return { isFullscreen: () => false, lockKeyboard() {}, unlockKeyboard() {} }
      if (path.endsWith('.vue')) return {}
      throw new Error(`Unexpected component module: ${path}`)
    },
  })
  const vm = new exports.default({
    beforeCreate() {
      this.$accessor = { video:state, hls:Vue.observable({ selected:undefined }), media:Vue.observable({ selected:false }), user:{}, remote:{}, locked:{}, chat:{ emotes:[] }, settings:{ autoplay } }
      this.$client = { on() {}, off() {}, playWebCodecs: () => Promise.resolve(true) }
      this.$log = { error() {}, debug() {} }
    },
    render: h => h('div'),
  })
  Object.assign(vm.$refs, { video, container:eventTarget(), component:{}, overlay:eventTarget() })
  vm.onResize = () => {}
  vm.$mount()
  async function advance(milliseconds) {
    const end = now + milliseconds
    while (true) {
      const next = [...timers].filter(([, timer]) => timer.at <= end).sort((a, b) => a[1].at - b[1].at)[0]
      if (!next) break
      now = next[1].at; timers.delete(next[0]); next[1].callback(); await flush()
    }
    now = end; await flush()
  }
  return { vm, video, state, timers, document, advance, stream }
}

test('mounted Vue callbacks update live health and cannot hide seek-only stalls', async t => {
  const h = harness(); t.after(() => h.vm.$destroy()); await flush()
  h.video.readyState = 2
  h.video.currentTime = 0.25
  h.video.fire('timeupdate')
  assert.equal(h.vm.playbackHealth.sample.time, 0.25)
  h.video.fire('seeking'); h.video.currentTime = 60; h.video.fire('seeked'); h.video.fire('timeupdate')
  await h.advance(8000)
  assert.equal(h.vm.playbackHealth.attempts, 1)
  h.video.currentTime += 0.25; h.video.fire('timeupdate')
  assert.equal(h.vm.playbackHealth.attempts, 0)
})

test('startup and recovery are bounded without requiring canplay or resetting on unmute', async t => {
  const h = harness(); t.after(() => h.vm.$destroy()); await flush()
  assert.equal(h.video.playCalls, 1)
  await h.advance(32000)
  assert.equal(h.state.playing, false)
  assert.equal(h.timers.size, 0)
  assert.equal(h.video.playCalls, 1) // Healthy play() state is not an extra repair mechanism.
})

test('gesture and native Pause stop recovery; Play works with an undefined return', async t => {
  const h = harness({ autoplay:false }); t.after(() => h.vm.$destroy()); await flush()
  await h.vm.play(); await flush()
  assert.equal(h.state.playing, true)
  assert.equal(h.video.playCalls, 1)
  h.video.fire('playing')
  h.video.pause(); await flush()
  assert.equal(h.state.playing, false)
  await h.advance(10000)
  assert.equal(h.timers.size, 0)
  await h.vm.play(); h.vm.pause(); await flush()
  assert.equal(h.state.playing, false)
  assert.equal(h.video.paused, true)
})

test('blocked autoplay ends at Play and a stale rejection cannot mute a replacement', async t => {
  const blocked = harness({ play: () => Promise.reject(Object.assign(new Error('blocked'), { name:'NotAllowedError' })) })
  t.after(() => blocked.vm.$destroy()); await flush()
  assert.equal(blocked.state.playing, false)
  assert.equal(blocked.timers.size, 0)
  let reject
  const h = harness({ play: () => new Promise((_, failure) => { reject = failure }) })
  t.after(() => h.vm.$destroy()); await flush()
  const oldReject = reject
  h.state.stream = { ...eventTarget(), getVideoTracks: () => [{}] }; await flush()
  oldReject(Object.assign(new Error('stale'), { name:'NotAllowedError' })); await flush()
  assert.equal(h.state.muted, false)
  assert.equal(h.state.playing, true)
  h.vm.$destroy(); reject(Object.assign(new Error('disposed'), { name:'NotAllowedError' })); await flush()
  assert.equal(h.state.muted, false)
  assert.equal(h.timers.size, 0)
})

test('background throttling and repeated metadata/playing events do not prove recovery', async t => {
  const h = harness(); t.after(() => h.vm.$destroy()); await flush()
  h.document.hidden = true; await h.advance(20000)
  assert.equal(h.vm.playbackHealth.attempts, 0)
  h.document.hidden = false; await h.advance(500)
  h.video.fire('loadedmetadata'); h.video.fire('playing')
  await h.advance(8000)
  assert.equal(h.vm.playbackHealth.attempts, 1)
})

test('the existing object-URL fallback still uses the common playback owner', async t => {
  const h = harness({ autoplay:false }); t.after(() => h.vm.$destroy()); await flush()
  delete h.video.srcObject
  h.state.stream = { ...eventTarget(), getVideoTracks: () => [{}] }; await flush()
  await h.vm.play(); await flush()
  assert.equal(h.video.src, 'blob:legacy-stream')
  assert.equal(h.video.playCalls, 1)
  assert.equal(h.state.playing, true)
})

test('WebCodecs toggle retains its own store/controller path', async t => {
  const h = harness({ autoplay:false }); t.after(() => h.vm.$destroy()); await flush()
  h.vm.$accessor.media.selected = true
  h.vm.toggle(); await flush()
  assert.equal(h.state.playing, true)
  assert.equal(h.video.playCalls, 0)
  h.vm.toggle(); await flush()
  assert.equal(h.state.playing, false)
})

test('late WebCodecs audio resume cannot mutate state after Pause, selection or destruction', async t => {
  for (const change of ['pause', 'selection', 'destroy', 'unmute']) {
    const h = harness({ autoplay:false }); t.after(() => h.vm.$destroy()); await flush()
    h.vm.$accessor.media.selected = true
    h.vm.$accessor.media.audioEnabled = true
    const overlay = h.vm.mutedOverlay
    let complete
    h.vm.$client.playWebCodecs = () => new Promise(resolve => { complete = resolve })
    h.state.play(); await flush()
    if (change === 'pause') { h.state.pause(); await flush() }
    if (change === 'selection') h.vm.$accessor.media.selected = false
    if (change === 'destroy') h.vm.$destroy()
    if (change === 'unmute') {
      const oldComplete = complete
      h.vm.unmute()
      h.state.pause(); await flush()
      oldComplete(false)
    }
    complete(false); await flush()
    assert.equal(h.state.muted, false, change)
    assert.equal(h.vm.mutedOverlay, overlay, change)
  }
})
