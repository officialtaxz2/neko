// Browser harness for server/internal/mediahls/browser_integration_test.go: the production MSE player
// (core/src/transport/hls/player.ts) plus the HLS transport's master/keepalive watch, reporting
// hls.js errors, terminal failures and playback progress to /log.
import Hls from 'hls.js'
import { createMSEPlayer } from '../../core/src/transport/hls/player.ts'

const log = (...parts) => {
  const message = parts.map((part) => (typeof part === 'string' ? part : JSON.stringify(part))).join(' ')
  fetch('/log', { method: 'POST', body: message }).catch(() => {})
}
let hls
const trigger = Hls.prototype.trigger
Hls.prototype.trigger = function (event, data) {
  hls = this
  if (event === Hls.Events.ERROR) {
    log('ERROR', {
      details: data.details,
      fatal: data.fatal,
      reason: data.reason,
      code: data.response?.code,
      parent: data.parent,
    })
  }
  return trigger.call(this, event, data)
}

const master = new URL(location.hash.slice(1), location.href).href
const mode = new URLSearchParams(location.search).get('mode')
const video = document.querySelector('video')
video.muted = true
createMSEPlayer(video, master, mode, (detail) => log('TERMINAL', detail), '/hls.worker.js')
video.play().catch(() => {})
for (const name of ['pause', 'playing', 'error', 'ended'])
  video.addEventListener(name, () => log('EVENT', name, video.currentTime.toFixed(2)))
setInterval(() => {
  const end = video.buffered.length ? video.buffered.end(video.buffered.length - 1) : 0
  log(
    'PROGRESS',
    video.currentTime.toFixed(2),
    (end - video.currentTime).toFixed(2),
    String(video.paused),
    'latency',
    Number(hls?.latency || 0).toFixed(2),
    'target',
    Number(hls?.targetLatency || 0).toFixed(2),
    'rate',
    video.playbackRate,
    'sync',
    String(hls?.liveSyncPosition),
  )
}, 2000)

let keepAliveAt = 0
setInterval(async () => {
  try {
    if (Date.now() >= keepAliveAt) {
      const alive = await fetch(new URL('keepalive', new URL('.', master)), {
        method: 'POST',
        credentials: 'same-origin',
      })
      if (!alive.ok) log('KEEPALIVE', alive.status)
      keepAliveAt = Date.now() + 15000
    }
    const response = await fetch(master, { credentials: 'same-origin', cache: 'no-store' })
    if (!response.ok) log('MASTER', response.status)
    await response.text()
  } catch {
    log('WATCHDOG failed')
  }
}, 1000)
