// Pure validation shared by the player and the dependency-free contract tests.
export const HLS_IDLE_MS = 30000
export const HLS_PLAYLIST_BYTES = 65536
export const HLS_VIDEO_TYPE = 'video/mp4; codecs="avc1.64001f"'
export const HLS_AUDIO_TYPE = 'audio/mp4; codecs="mp4a.40.2"'

export function hlsModes(payload) {
  if (!payload || payload.version !== 1 || payload.backend !== 'hls' ||
      payload.container !== 'fmp4' || payload.video_codec !== 'h264-high-3.1' ||
      payload.audio_codec !== 'aac-lc' || payload.audio_rate !== 48000 ||
      !Array.isArray(payload.modes) || !payload.modes.length || payload.modes.length > 2 ||
      new Set(payload.modes).size !== payload.modes.length ||
      payload.modes.some((mode) => mode !== 'hls' && mode !== 'll-hls') ||
      !Array.isArray(payload.variants) || payload.variants.length !== 3 ||
      !payload.limits || payload.limits.idle_expires_in_ms !== HLS_IDLE_MS ||
      payload.limits.maximum_playlist_bytes !== HLS_PLAYLIST_BYTES ||
      payload.limits.max_requests_per_lease !== 4 || payload.limits.max_blocking_per_lease !== 2) return []
  const sizes = { high: [1280, 720, 25], medium: [854, 480, 20], low: [640, 360, 15] }
  const seen = new Set()
  for (const variant of payload.variants) {
    if (!variant || !['high', 'medium', 'low'].includes(variant.id)) return []
    const size = sizes[variant.id]
    if (!size || seen.has(variant.id) || variant.container !== 'video/mp4' ||
        variant.video_codec !== 'avc1.64001f' || variant.width !== size[0] ||
        variant.height !== size[1] || variant.frame_rate !== size[2] ||
        !Number.isSafeInteger(variant.bandwidth) || variant.bandwidth <= 0) return []
    seen.add(variant.id)
  }
  return payload.modes.slice()
}

// Derive the deployment prefix from the credential-free event URL, never a login URL.
export function hlsBase(eventURL, pageURL) {
  const event = new URL(eventURL)
  const page = new URL(pageURL)
  if (event.protocol !== 'wss:' || page.protocol !== 'https:' || event.username || event.password ||
      event.search || event.hash || !event.pathname.endsWith('/ws')) throw new Error('HLS requires same-origin HTTPS')
  event.protocol = 'https:'
  if (event.origin !== page.origin) throw new Error('HLS requires same-origin HTTPS')
  event.pathname = event.pathname.slice(0, -3) + '/api/media/hls/'
  return event.href
}

export function validHLSOffer(offer, mode) {
  return !!offer && offer.version === 1 && offer.backend === 'hls' && offer.mode === mode &&
    offer.path === '/api/media/hls/session' && typeof offer.ticket === 'string' &&
    /^[A-Za-z0-9_-]{32}$/.test(offer.ticket) && Number.isInteger(offer.expires_in_ms) &&
    offer.expires_in_ms > 0 && offer.expires_in_ms <= 10000
}

export function hlsLeaseURL(response, mode, base) {
  if (!response || response.mode !== mode || response.idle_expires_in_ms !== HLS_IDLE_MS ||
      typeof response.master !== 'string' ||
      !/^\/api\/media\/hls\/[A-Za-z0-9_-]{22}\/master\.m3u8$/.test(response.master)) {
    throw new Error('Invalid HLS bootstrap response')
  }
  return new URL(response.master.slice('/api/media/hls/'.length), base).href
}

// Only the negotiated public lease path and the two LL-HLS directives are allowed.
export function scopedHLSURL(candidate, master, mode) {
  if (typeof candidate !== 'string' || /[\\\s]/.test(candidate) || /%|(?:^|\/)\.\.(?:\/|$)/.test(candidate)) return false
  try {
    const url = new URL(candidate, master)
    const root = new URL('.', master)
    if (url.protocol !== 'https:' || url.origin !== root.origin || url.username || url.password ||
        url.hash || !url.pathname.startsWith(root.pathname)) return false
    const relative = url.pathname.slice(root.pathname.length)
    if (!/^(?:master\.m3u8|(?:high|medium|low|audio)\/(?:index\.m3u8|init-[1-9][0-9]*\.mp4|seg-[1-9][0-9]*\.m4s|part-[1-9][0-9]*-[0-5]\.m4s))$/.test(relative)) return false
    const keys = [...url.searchParams.keys()]
    if (!keys.length) return !url.search
    if (mode !== 'll-hls' || !relative.endsWith('/index.m3u8') || keys.length > 2 ||
        new Set(keys).size !== keys.length || !keys.includes('_HLS_msn')) return false
    return keys.every((key) => (key === '_HLS_msn' || key === '_HLS_part') &&
      /^(?:0|[1-9][0-9]*)$/.test(url.searchParams.get(key) || '') &&
      Number.isSafeInteger(Number(url.searchParams.get(key))) &&
      (key !== '_HLS_part' || Number(url.searchParams.get(key)) <= 6))
  } catch (_) { return false }
}

export function validHLSPlaylist(text, playlistURL, master, mode) {
  if (typeof text !== 'string' || text.length > HLS_PLAYLIST_BYTES || !text.startsWith('#EXTM3U\n') ||
      /#EXT-X-(?:KEY|SESSION-KEY|CONTENT-STEERING|DEFINE|INTERSTITIAL)/.test(text)) return false
  for (const line of text.split('\n')) {
    const values = playlistURIs(line)
    for (const value of values) {
      // Rendition reports legitimately use ../low within this same lease.
      try {
        if (/[%\\\s]/.test(value) || !scopedHLSURL(new URL(value, playlistURL).href, master, mode)) return false
      } catch (_) { return false }
    }
  }
  return true
}

function playlistURIs(line) {
  if (!line.startsWith('#')) return [line.trim()].filter(Boolean)
  const values = []
  const pattern = /(?:^|[, :])(?:[A-Z-]*URI)="([^"]+)"/g
  let match
  while ((match = pattern.exec(line)) !== null) values.push(match[1])
  return values
}

export function terminalHLSStatus(status) {
  return [400, 401, 403, 404, 410, 413].includes(status)
}

export function hlsMasterChildren(text, master, mode) {
  if (!validHLSPlaylist(text, master, master, mode)) throw new Error('Invalid HLS master')
  const children = []
  for (const line of text.split('\n')) {
    const values = playlistURIs(line)
    for (const value of values) {
      const url = new URL(value, master)
      if (!/(?:high|medium|low|audio)\/index\.m3u8$/.test(url.pathname) || url.search) throw new Error('Invalid HLS master child')
      if (!children.includes(url.href)) children.push(url.href)
    }
  }
  if (children.length < 2 || children.length > 4 || !children.some((value) => value.endsWith('/audio/index.m3u8'))) throw new Error('Invalid HLS master children')
  return children
}

export function clearHLSVideo(video) {
  if (!video) return
  try { video.pause() } catch (_) {}
  video.srcObject = null
  video.removeAttribute('src')
  // URL/MSE cleanup; this helper must never be used on the WebRTC element path.
  try { video.load() } catch (_) {}
}
