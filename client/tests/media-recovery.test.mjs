import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import {
  MEDIA_RETRY_DELAYS_MS,
  MEDIA_RETRY_STABILITY_MS,
  deliverOrReleaseVideoFrame,
  isMediaRetryCloseCode,
  mediaRetryDelayForAttempt,
  shouldAwaitMediaCloseForEnd,
  shouldDropDecodedVideoOutput,
} from '../src/neko/media/recovery.js'
import { elementRequestFullscreen } from '../src/utils/fullscreen.js'

test('media-only retry is limited to backpressure and backend failure closes', () => {
  assert.equal(isMediaRetryCloseCode(4413), true)
  assert.equal(isMediaRetryCloseCode(4500), true)
  for (const code of [1000, 1001, 1002, 4400, 4401, 4403, 4408, 4409, 4429]) {
    assert.equal(isMediaRetryCloseCode(code), false)
  }
})

test('media-only retry uses four serialized bounded delays', () => {
  assert.deepEqual(MEDIA_RETRY_DELAYS_MS, [1000, 2000, 5000, 10000])
  assert.equal(MEDIA_RETRY_STABILITY_MS, 30_000)
  MEDIA_RETRY_DELAYS_MS.forEach((delay, attempt) => assert.equal(mediaRetryDelayForAttempt(attempt), delay))
  assert.equal(mediaRetryDelayForAttempt(MEDIA_RETRY_DELAYS_MS.length), null)
})

test('backend-error END waits for the retryable private close code', () => {
  assert.equal(shouldAwaitMediaCloseForEnd('backend_error'), true)
  for (const reason of ['normal', 'revoked', 'replaced', 'shutdown']) {
    assert.equal(shouldAwaitMediaCloseForEnd(reason), false)
  }
})

test('decoded video saturation drops output at the fixed renderer bound', () => {
  assert.equal(shouldDropDecodedVideoOutput(0, 2), false)
  assert.equal(shouldDropDecodedVideoOutput(1, 2), false)
  assert.equal(shouldDropDecodedVideoOutput(2, 2), true)
  assert.equal(shouldDropDecodedVideoOutput(3, 2), true)
})

test('video frames without a mounted renderer listener are released', () => {
  let releases = 0
  assert.equal(deliverOrReleaseVideoFrame(() => false, () => releases++), false)
  assert.equal(releases, 1)

  assert.equal(deliverOrReleaseVideoFrame(() => true, () => releases++), true)
  assert.equal(releases, 1)

  assert.throws(
    () => deliverOrReleaseVideoFrame(() => { throw new Error('listener failed') }, () => releases++),
    /listener failed/,
  )
  assert.equal(releases, 2)
})

test('Vue WebCodecs callbacks keep live component state', async () => {
  const source = await readFile(new URL('../src/components/video.vue', import.meta.url), 'utf8')

  for (const method of ['onWebCodecsFrame', 'onWebCodecsClockReset', 'renderWebCodecsFrame']) {
    assert.match(source, new RegExp(`private ${method}\\([^)]*\\) \\{`))
    assert.doesNotMatch(source, new RegExp(`private ${method}\\s*=\\s*\\(`))
  }
})

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
  assert.match(source, /if \(await elementRequestFullscreen\(surface\)\)/)
  assert.match(source, /this\.enterFallbackFullscreen\(\)/)
  assert.match(source, /'fallback-fullscreen': fallbackFullscreen/)
})

test('media feedback is self-scheduled and leaves skew recovery to the server', async () => {
  const source = await readFile(new URL('../src/neko/media/worker.ts', import.meta.url), 'utf8')

  assert.match(source, /feedbackTimer = scope\.setTimeout\(\(\) => \{/)
  assert.doesNotMatch(source, /setInterval\(sendFeedback/)
  assert.doesNotMatch(source, /requestResync\('av_skew'\)/)
})

test('AudioWorklet reports only a sustained 100 ms underflow', async () => {
  const originalProcessor = globalThis.AudioWorkletProcessor
  const originalRegister = globalThis.registerProcessor
  const originalSampleRate = globalThis.sampleRate
  const originalCurrentTime = globalThis.currentTime
  let Processor
  const messages = []

  globalThis.AudioWorkletProcessor = class {
    constructor() {
      this.port = {
        onmessage: null,
        postMessage: (message) => messages.push(message),
      }
    }
  }
  globalThis.registerProcessor = (name, constructor) => {
    assert.equal(name, 'neko-media-audio')
    Processor = constructor
  }
  globalThis.sampleRate = 48_000
  globalThis.currentTime = 0

  try {
    const workletURL = new URL('../src/neko/media/audio-worklet.js', import.meta.url)
    workletURL.searchParams.set('test', String(Date.now()))
    await import(workletURL)
    assert.equal(typeof Processor, 'function')

    const outputs = () => [[new Float32Array(128), new Float32Array(128)]]
    const sustained = new Processor()
    sustained.active = true
    sustained.process([], outputs())
    globalThis.currentTime = 0.05
    sustained.process([], outputs())
    assert.equal(messages.some((message) => message.type === 'underflow'), false)
    globalThis.currentTime = 0.101
    sustained.process([], outputs())
    assert.equal(messages.filter((message) => message.type === 'underflow').length, 1)

    const recovered = new Processor()
    globalThis.currentTime = 1
    recovered.active = true
    recovered.process([], outputs())
    globalThis.currentTime = 1.05
    recovered.port.onmessage({
      data: {
        type: 'chunk',
        id: 1,
        generation: 1,
        durationMS: 128 / 48,
        startTime: 1.05,
        numberOfFrames: 128,
        planes: [new Float32Array(128).buffer, new Float32Array(128).buffer],
      },
    })
    recovered.process([], outputs())
    assert.equal(messages.filter((message) => message.type === 'underflow').length, 1)

    const rebuffered = new Processor()
    rebuffered.port.onmessage({
      data: {
        type: 'chunk',
        id: 2,
        generation: 3,
        durationMS: 20,
        startTime: 2,
        numberOfFrames: 960,
        planes: [new Float32Array(960).buffer, new Float32Array(960).buffer],
      },
    })
    rebuffered.port.onmessage({ data: { type: 'rebuffer' } })
    assert.deepEqual(messages.at(-1), {
      type: 'discarded',
      id: 2,
      generation: 3,
      durationMS: 20,
    })
    assert.equal(rebuffered.queue.length, 0)
    assert.equal(rebuffered.bufferedFrames, 0)
    assert.equal(rebuffered.active, false)
  } finally {
    if (originalProcessor === undefined) delete globalThis.AudioWorkletProcessor
    else globalThis.AudioWorkletProcessor = originalProcessor
    if (originalRegister === undefined) delete globalThis.registerProcessor
    else globalThis.registerProcessor = originalRegister
    if (originalSampleRate === undefined) delete globalThis.sampleRate
    else globalThis.sampleRate = originalSampleRate
    if (originalCurrentTime === undefined) delete globalThis.currentTime
    else globalThis.currentTime = originalCurrentTime
  }
})

test('AudioWorklet underflow locally rebuffers instead of requesting a common server resync', async () => {
  const controller = await readFile(new URL('../src/neko/media/controller.ts', import.meta.url), 'utf8')
  const worker = await readFile(new URL('../src/neko/media/worker.ts', import.meta.url), 'utf8')

  assert.match(controller, /const AUDIO_REBUFFER_LEAD_MS = 160/)
  assert.match(controller, /else if \(data\?\.type === 'underflow'\) \{\s*this\.rebufferAudio\(node\)/)
  assert.match(controller, /rendered: data\.type === 'consumed'/)
  assert.match(controller, /node\.port\.postMessage\(\{ type: 'rebuffer' \}\)/)
  assert.match(controller, /this\.audioLeadMS = Math\.max\(this\.audioLeadMS, AUDIO_REBUFFER_LEAD_MS\)/)
  assert.match(controller, /if \(!localRebuffer\) \{\s*this\.callbacks\.onClockReset\(\)/)
  assert.doesNotMatch(controller, /type: data\.type === 'underflow' \? 'audio-underflow'/)
  assert.doesNotMatch(worker, /case 'audio-underflow'/)
})
