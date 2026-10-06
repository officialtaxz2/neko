// Paste into the existing Neko page's DevTools console. Read-only: no requests,
// player actions, timers, fetch wrapping, storage access or raw URL/header output.
(() => {
  const number = (value) => Number.isFinite(value) && value >= 0 ? Math.round(value) : null
  const all = performance.getEntriesByType('resource')
  const rows = []
  for (const entry of all.slice(-2000)) {
    let url
    try { url = new URL(entry.name) } catch (_) { continue }
    if (url.origin !== location.origin) continue
    const match = /\/api\/media\/hls\/(session|[A-Za-z0-9_-]{22}\/(.+))$/.exec(url.pathname)
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
    rows.push({ kind, variant, started_ms: number(entry.startTime), duration_ms: number(entry.duration),
      first_byte_ms: entry.responseStart > 0 ? number(entry.responseStart - entry.startTime) : null,
      status: Number.isInteger(entry.responseStatus) && entry.responseStatus >= 0 && entry.responseStatus <= 599 ? entry.responseStatus : null,
      body_bytes: number(entry.decodedBodySize), transfer_bytes: number(entry.transferSize) })
  }
  rows.sort((left, right) => left.started_ms - right.started_ms)
  const video = document.querySelector('.video .player-container video')
  const result = { diagnostic: 'hls-browser-timing-v1', clock_ms: number(performance.now()),
    total_timing_entries: all.length, hls_entries_in_last_2000: rows.length,
    current_media: video ? { paused: video.paused, ended: video.ended, ready_state: number(video.readyState),
      network_state: number(video.networkState), current_time_seconds: number(video.currentTime),
      error_code: video.error && [1, 2, 3, 4].includes(video.error.code) ? video.error.code : null } : null,
    recent_http: rows.slice(-40),
    limits: 'Retained browser timing entries only; buffer may omit recent requests. Zero/null status or sizes do not prove a timeout. Current media may already be cleared by terminal cleanup; no original frozen-player state, response body, peer attribution or causal proof.' }
  console.log(JSON.stringify(result, null, 2))
})()
