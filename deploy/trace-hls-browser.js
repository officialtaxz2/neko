// Paste into the Neko HLS page's DevTools console, then click Retry HLS once
// if it is already failed. Leave the tab open; one JSON report prints on a new
// terminal error or after five minutes. Observation only: no requests, player
// actions, fetch wrapping, cookie/storage reads or performance-buffer changes.
// Temporary observers/listeners/timers and one stop function are removed on exit.
(() => {
  const key = '__nekoHLSBrowserTraceStop'
  if (typeof window[key] === 'function') window[key]()
  if (typeof PerformanceObserver !== 'function' ||
      !PerformanceObserver.supportedEntryTypes?.includes('resource')) {
    console.log('{"diagnostic":"hls-browser-trace-v1","result":"unsupported_observer"}')
    return
  }
  const start = performance.now()
  const number = (value) => Number.isFinite(value) && value >= 0 ? Math.round(value) : null
  const elapsed = () => number(performance.now() - start)
  const http = [], slow = [], samples = [], events = [], counts = {}
  const removed = { http: 0, slow: 0, samples: 0, events: 0 }
  let video, observer, tick, deadline, terminalTimer, activeSeen = false, done = false
  let removeListeners = () => {}
  const keep = (list, row, maximum, name) => {
    if (list.length >= maximum) { list.shift(); removed[name]++ }
    list.push(row)
  }
  const ui = () => {
    const node = document.querySelector('.video .webcodecs-status')
    const label = node?.querySelector('strong, span')?.textContent?.trim()
    if (!['HLS', 'LL-HLS'].includes(label)) return { selected: false }
    const terminal = !!node.querySelector('.webcodecs-actions')
    const text = terminal ? node.querySelector('small')?.textContent : ''
    const detail = text === 'HLS HTTP connection failed; retry manually' ? 'http_connection' :
      text === 'HLS playback stalled; retry manually' ? 'playback_stalled' :
      text === 'HLS playback did not become ready; retry manually' ? 'playback_not_ready' :
      text === 'HLS media did not become ready' ? 'media_not_ready' :
      text === 'HLS bootstrap failed; retry manually' ? 'bootstrap' :
      text === 'HLS media was unavailable at startup; retry manually' ? 'startup_unavailable' :
      text === 'HLS authorization or media availability ended' ? 'authorization_or_availability' :
      text === 'HLS lease ended' ? 'lease_ended' : terminal ? 'other_fixed_error' : null
    return { selected: true, compact: node.classList.contains('compact'), terminal, detail }
  }
  const media = () => {
    if (!video) return null
    const time = video.currentTime
    const ranges = []
    try {
      for (let i = 0; i < Math.min(video.buffered.length, 4); i++)
        ranges.push([number(video.buffered.start(i) * 1000), number(video.buffered.end(i) * 1000)])
    } catch (_) {}
    let frames = null
    try { frames = number(video.getVideoPlaybackQuality?.().totalVideoFrames) } catch (_) {}
    return { time_ms: number(time * 1000), paused: video.paused, ended: video.ended,
      ready: number(video.readyState), network: number(video.networkState), muted: video.muted,
      error: [1, 2, 3, 4].includes(video.error?.code) ? video.error.code : null,
      ranges_ms: ranges, range_count: number(video.buffered.length), frames }
  }
  const bind = () => {
    const next = document.querySelector('.video .player-container video')
    if (next === video) return
    removeListeners()
    video = next
    const bound = video, listeners = []
    if (bound) {
      for (const event of ['playing', 'pause', 'waiting', 'stalled', 'seeking', 'seeked',
        'canplay', 'error', 'ended', 'emptied', 'loadedmetadata']) {
        const handler = () => { if (!done) keep(events, { t_ms: elapsed(), event, media: media() }, 80, 'events') }
        bound.addEventListener(event, handler)
        listeners.push(() => bound.removeEventListener(event, handler))
      }
    }
    removeListeners = () => listeners.forEach(remove => remove())
  }
  const receive = (entries) => {
    for (const entry of entries) {
      let url
      try { url = new URL(entry.name) } catch (_) { continue }
      if (url.origin !== location.origin) continue
      const match = /^\/api\/media\/hls\/(session|[A-Za-z0-9_-]{22}\/(.+))$/.exec(url.pathname)
      if (!match) continue
      const path = match[1] === 'session' ? 'session' : match[2]
      let kind, variant = null
      if (['session', 'master.m3u8', 'keepalive'].includes(path)) {
        kind = path === 'master.m3u8' ? 'master' : path
      } else {
        const resource = /^(audio|high|medium|low)\/(index\.m3u8|init-[1-9][0-9]*\.mp4|seg-[1-9][0-9]*\.m4s|part-[1-9][0-9]*-[0-5]\.m4s)$/.exec(path)
        if (!resource) continue
        variant = resource[1]
        kind = resource[2] === 'index.m3u8' ? 'playlist' : resource[2].split('-')[0]
        if (kind === 'seg') kind = 'segment'
      }
      const row = { kind, variant, started_ms: Math.round(entry.startTime - start),
        observed_ms: elapsed(), duration_ms: number(entry.duration),
        first_byte_ms: entry.responseStart > 0 ? number(entry.responseStart - entry.startTime) : null,
        status: Number.isInteger(entry.responseStatus) && entry.responseStatus >= 0 && entry.responseStatus <= 599 ? entry.responseStatus : null,
        body_bytes: number(entry.decodedBodySize), transfer_bytes: number(entry.transferSize) }
      const name = kind + (variant ? '_' + variant : '')
      counts[name] = (counts[name] || 0) + 1
      keep(http, row, 180, 'http')
      if (row.duration_ms >= 900) keep(slow, row, 40, 'slow')
      // Do not treat a completion from the already failed attempt as Retry.
      if (kind === 'session' && entry.startTime >= start) activeSeen = true
    }
  }
  const sample = () => {
    bind()
    const state = ui()
    keep(samples, { t_ms: elapsed(), visible: document.visibilityState === 'visible', ui: state, media: media() }, 155, 'samples')
    if (state.selected && !state.terminal) activeSeen = true
    if (activeSeen && state.terminal && terminalTimer === undefined)
      terminalTimer = window.setTimeout(() => finish('terminal'), 2000)
  }
  const finish = (reason) => {
    if (done) return
    // Drain newly queued timings before disconnect; old timeline entries are
    // deliberately not imported. Starting before Retry preserves the failure.
    receive(observer.takeRecords())
    sample()
    done = true
    observer.disconnect()
    removeListeners()
    window.clearInterval(tick)
    window.clearTimeout(deadline)
    window.clearTimeout(terminalTimer)
    if (window[key] === stop) delete window[key]
    console.log(JSON.stringify({ diagnostic: 'hls-browser-trace-v1', reason,
      duration_ms: elapsed(), request_counts: counts, discarded_rows: removed,
      media_samples: samples, media_events: events, recent_http: http, slow_http: slow,
      limits: 'New completed Resource Timing entries only, with possible browser omissions; zero/null status is not failure proof. Request start can precede installation. Body sizes are timing metadata, not inspected bodies. Media state is sampled; frames/time do not prove audible sound or timeline validity. No automatic retry, network/player changes, credentials, URLs, peer attribution or causal proof.' }))
  }
  const stop = () => finish('manual_or_replaced')
  observer = new PerformanceObserver(list => { if (!done) receive(list.getEntries()) })
  try { observer.observe({ entryTypes: ['resource'] }) } catch (_) {
    observer.disconnect()
    console.log('{"diagnostic":"hls-browser-trace-v1","result":"observer_failed"}')
    return
  }
  window[key] = stop
  sample()
  tick = window.setInterval(sample, 2000)
  deadline = window.setTimeout(() => finish('five_minute_limit'), 300000)
  console.log('HLS trace armed. If already failed, click Retry HLS once. Keep this tab open; a JSON report prints on a new failure or after five minutes.')
})()
