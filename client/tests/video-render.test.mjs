import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'
import Vue from 'vue'
import * as progress from '../src/neko/playback-progress.js'
import * as videoDebug from '../src/neko/video-debug.js'

// Renders the real video.vue template once, before any ref exists, the way
// the first page render does. Settings come from the stored values.
const require = createRequire(import.meta.url)
const { compileTemplate } = require('vue/compiler-sfc')
const file = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')
const source = file.match(/<script lang="ts">([\s\S]*?)<\/script>/)[1]
const template = file.match(/^<template>([\s\S]*)<\/template>/m)[1]
const compiled = ts.transpileModule(source, { compilerOptions: {
  target:ts.ScriptTarget.ES2022, module:ts.ModuleKind.CommonJS,
  experimentalDecorators:true, useDefineForClassFields:false, esModuleInterop:true,
} }).outputText
const { code } = compileTemplate({ source:template, filename:'video.vue', prettify:false })
const renderFunctions = new Function(`${code.replace(/export /g, '')}; return { render, staticRenderFns }`)()

function firstRender(settings, { touch = true, search = '' } = {}) {
  const exports = {}, errors = []
  const listeners = () => ({ addEventListener() {}, removeEventListener() {} })
  runInNewContext(compiled, {
    exports, console, Object, Date,
    window:{ ...listeners(), setTimeout: () => 1, clearTimeout() {}, setInterval: () => 1, clearInterval() {},
      matchMedia: () => ({ matches:false }), requestAnimationFrame: () => 1, cancelAnimationFrame() {} },
    document:{ ...listeners(), hidden:false, body:{ style:{} } },
    location:{ search }, navigator:{ userAgent:'test', maxTouchPoints:touch ? 5 : 0, platform:'test' },
    require(path) {
      if (path === 'vue-property-decorator') return require(path)
      if (path === 'resize-observer-polyfill') return class { observe() {} disconnect() {} }
      if (path === '~/utils/guacamole-keyboard.ts') return () => ({ listenTo() {}, reset() {} })
      if (path === '~/neko/playback-progress.js') return progress
      if (path === '~/neko/video-debug.js') return videoDebug
      if (path === '~/utils/touch') return { isTouchDevice: () => touch }
      if (path === '~/utils') return { isFullscreen: () => false, lockKeyboard() {}, unlockKeyboard() {} }
      if (path.endsWith('.vue')) return {}
      throw new Error(`Unexpected component module: ${path}`)
    },
  })
  const previous = Vue.config.errorHandler
  Vue.config.errorHandler = (err) => { errors.push(String(err)) }
  try {
    const vm = new exports.default({
      ...renderFunctions,
      directives:{ tooltip:{} },
      beforeCreate() {
        this.$accessor = {
          video:{ resolution:{ w:1280, h:720 }, playing:false, playable:false, muted:false, volume:100, setPlayable() {} },
          hls:{ selected:undefined }, user:{ viewOnly:false, admin:false }, locked:{}, chat:{ emotes:[] },
          // Before the server's control state arrives, implicit hosting is assumed.
          remote:{ hosting:true, implicitHosting:true, controlling:false, hosted:true, locked:false },
          settings:{ autoplay:true, trackpad_mode:false, trackpad_cursor_hidden:false, force_touch:false, scroll:10, scroll_invert:true, ...settings },
        }
        this.$client = { on() {}, off() {}, inputStats:{ sent:0, skipped:0, last:'' }, transportState:'' }
        this.$log = { error() {}, debug() {} }
        this.$t = key => key
      },
    })
    const vnode = vm._render()
    vm.$destroy()
    return { vnode, errors }
  } finally {
    Vue.config.errorHandler = previous
  }
}

function findClass(vnode, name) {
  if (!vnode) return false
  const classes = [vnode.data?.staticClass, vnode.data?.class].flat().join(' ')
  if (classes.split(/\s+/).includes(name)) return true
  return (vnode.children || []).some(child => findClass(child, name))
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
