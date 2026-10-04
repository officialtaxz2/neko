// Lazy-loaded full player: alternate AAC audio support is required (not the light build).
import Hls, { Loader, LoaderCallbacks, LoaderConfiguration, LoaderContext, LoaderStats } from 'hls.js'
import workerURL from 'hls.js/dist/hls.worker.js?url'
import { HLS_PLAYLIST_BYTES, scopedHLSURL, terminalHLSStatus, validHLSPlaylist } from './protocol.js'

export interface HLSPlayer { destroy(): void }

export function mseSupported() {
  return Hls.isSupported()
}

export function createMSEPlayer(video: HTMLVideoElement, master: string, mode: 'hls' | 'll-hls', terminal: (detail: string) => void): HLSPlayer {
  let alive = true
  let active = 0
  const queue: Array<() => void> = []
  // Three media requests leave the fourth server slot for a watchdog/keepalive.
  const pump = () => { while (alive && active < 3 && queue.length) queue.shift()!() }
  const freshStats = (): LoaderStats => ({ aborted: false, loaded: 0, total: 0, retry: 0, chunkCount: 0, bwEstimate: 0,
    loading: { start: 0, first: 0, end: 0 }, parsing: { start: 0, end: 0 }, buffering: { start: 0, first: 0, end: 0 } })

  class ScopedLoader implements Loader<LoaderContext> {
    context: LoaderContext | null = null
    stats: LoaderStats = freshStats()
    private serial = 0
    private controller?: AbortController
    private timer?: number
    private pending?: () => void
    private release?: () => void

    load(context: LoaderContext, config: LoaderConfiguration, callbacks: LoaderCallbacks<LoaderContext>) {
      this.abort()
      const serial = this.serial
      const stats = this.stats = freshStats()
      const current = () => alive && this.serial === serial && !stats.aborted
      this.context = context
      if (!alive || !scopedHLSURL(context.url, master, mode) || queue.length >= 12) {
        terminal('HLS request boundary rejected')
        return
      }
      const begin = () => {
        this.pending = undefined
        if (!current()) return
        active++
        let released = false
        let timer: number
        const release = () => {
          if (released) return
          released = true
          active--
          window.clearTimeout(timer)
          if (this.timer === timer) this.timer = undefined
          if (this.release === release) this.release = undefined
          pump()
        }
        this.release = release
        this.controller = new AbortController()
        const controller = this.controller
        const playlist = context.responseType !== 'arraybuffer'
        const pathname = new URL(context.url).pathname
        const maximum = playlist ? HLS_PLAYLIST_BYTES : pathname.endsWith('.mp4') ? 2 * 1024 * 1024 :
          pathname.includes('/part-') ? 1024 * 1024 : 8 * 1024 * 1024
        stats.loading.start = performance.now()
        let timedOut = false
        timer = window.setTimeout(() => { timedOut = true; controller.abort() }, Math.min(config.loadPolicy.maxLoadTimeMs, 12000))
        this.timer = timer
        void (async () => {
          try {
            const response = await fetch(context.url, { credentials: 'same-origin', cache: 'no-store', redirect: 'error',
              referrerPolicy: 'no-referrer', signal: controller.signal })
            stats.loading.first = performance.now()
            if (!current()) return
            if (!response.ok) {
              if (terminalHLSStatus(response.status)) terminal('HLS authorization or media availability ended')
              else callbacks.onError({ code: response.status, text: 'HLS request failed' }, context, null, stats)
              return
            }
            if (!response.body || Number(response.headers.get('Content-Length')) > maximum) throw new Error('bounded body')
            const reader = response.body.getReader()
            const chunks: Uint8Array[] = []
            let size = 0
            for (;;) {
              const { done, value } = await reader.read()
              if (!current()) return
              if (done) break
              size += value.byteLength
              if (size > maximum) { controller.abort(); terminal('HLS response exceeded its bound'); return }
              chunks.push(value)
            }
            if (!current()) return
            const data = new Uint8Array(size)
            let offset = 0
            for (const chunk of chunks) { data.set(chunk, offset); offset += chunk.byteLength }
            stats.loaded = stats.total = size
            stats.loading.end = performance.now()
            stats.chunkCount = 1
            stats.bwEstimate = size * 8000 / Math.max(1, stats.loading.end - stats.loading.start)
            const result = playlist ? new TextDecoder().decode(data) : data.buffer
            if (playlist && !validHLSPlaylist(result, context.url, master, mode)) { terminal('Invalid HLS playlist boundary'); return }
            callbacks.onSuccess({ url: context.url, data: result }, stats, context, null)
          } catch (_) {
            if (!current()) return
            if (timedOut) callbacks.onTimeout(stats, context, null)
            else callbacks.onError({ code: 0, text: 'HLS request failed' }, context, null, stats)
          } finally { controller.abort(); release() }
        })()
      }
      this.pending = begin
      queue.push(begin)
      pump()
    }

    abort() {
      this.serial++
      this.stats.aborted = true
      if (this.pending) {
        const index = queue.indexOf(this.pending)
        if (index >= 0) queue.splice(index, 1)
        this.pending = undefined
      }
      this.controller?.abort()
      this.release?.()
    }
    destroy() { this.abort(); this.context = null }
  }

  const retry = { maxNumRetry: 2, retryDelayMs: 1000, maxRetryDelayMs: 2000 }
  const policy = { default: { maxTimeToFirstByteMs: 10000, maxLoadTimeMs: 12000, timeoutRetry: retry, errorRetry: retry } }
  const hls = new Hls({ debug: false, enableWorker: true, workerPath: workerURL, loader: ScopedLoader,
    lowLatencyMode: mode === 'll-hls', progressive: false, startLevel: -1,
    maxBufferLength: mode === 'll-hls' ? 6 : 18, maxMaxBufferLength: 24, backBufferLength: 6,
    maxBufferSize: 16 * 1024 * 1024, liveSyncDurationCount: 3, liveMaxLatencyDurationCount: 5,
    manifestLoadPolicy: policy, playlistLoadPolicy: policy, fragLoadPolicy: policy })
  hls.on(Hls.Events.ERROR, (_event, data) => {
    // Error objects include URLs; only fixed messages may reach state/UI/logs.
    if (alive && (data.fatal || terminalHLSStatus(data.response?.code))) terminal('HLS player stopped; retry manually')
  })
  try {
    hls.attachMedia(video)
    hls.loadSource(master)
  } catch (_) {
    alive = false
    queue.length = 0
    hls.destroy()
    throw new Error('HLS player initialization failed')
  }
  return { destroy() { alive = false; queue.length = 0; hls.destroy() } }
}
