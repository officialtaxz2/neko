// The HLS contract (docs/HLS_LL_HLS.md), checked without side effects: what the server advertises
// and offers, where a lease lives, and which URLs and playlists the player may touch. Shared by the
// transport (../hls.ts), the MSE player (player.ts) and their tests.

export type HLSMode = 'hls' | 'll-hls'

export const HLS_IDLE_MS = 30000
export const HLS_PLAYLIST_BYTES = 65536
export const HLS_VIDEO_TYPE = 'video/mp4; codecs="avc1.64001f"'
export const HLS_AUDIO_TYPE = 'audio/mp4; codecs="mp4a.40.2"'

const SIZES: Record<string, [number, number, number]> = {
  high: [1280, 720, 25],
  medium: [854, 480, 20],
  low: [640, 360, 15],
}

// the modes a media/hls/capabilities payload advertises, or none when anything in it is not the
// version-1 contract (codecs, container, the three variants, the limits)
export function hlsModes(payload: any): HLSMode[] {
  if (
    !payload ||
    payload.version !== 1 ||
    payload.backend !== 'hls' ||
    payload.container !== 'fmp4' ||
    payload.video_codec !== 'h264-high-3.1' ||
    payload.audio_codec !== 'aac-lc' ||
    payload.audio_rate !== 48000 ||
    !Array.isArray(payload.modes) ||
    !payload.modes.length ||
    payload.modes.length > 2 ||
    new Set(payload.modes).size !== payload.modes.length ||
    payload.modes.some((mode: unknown) => mode !== 'hls' && mode !== 'll-hls') ||
    !Array.isArray(payload.variants) ||
    payload.variants.length !== 3 ||
    !payload.limits ||
    payload.limits.idle_expires_in_ms !== HLS_IDLE_MS ||
    payload.limits.maximum_playlist_bytes !== HLS_PLAYLIST_BYTES ||
    // the player needs at least four requests (two blocking) per lease
    !(payload.limits.max_requests_per_lease >= 4) ||
    !(payload.limits.max_blocking_per_lease >= 2)
  )
    return []
  const seen = new Set<string>()
  for (const variant of payload.variants) {
    const size = variant && SIZES[variant.id]
    if (
      !size ||
      seen.has(variant.id) ||
      variant.container !== 'video/mp4' ||
      variant.video_codec !== 'avc1.64001f' ||
      variant.width !== size[0] ||
      variant.height !== size[1] ||
      variant.frame_rate !== size[2] ||
      !Number.isSafeInteger(variant.bandwidth) ||
      variant.bandwidth <= 0
    )
      return []
    seen.add(variant.id)
  }
  return payload.modes.slice()
}

// the lease routes under the server's URL (its path prefix kept): HTTPS on the page's own origin,
// never a URL with credentials, a query or a fragment
export function hlsBase(serverURL: string, pageURL: string) {
  const server = new URL(serverURL)
  const page = new URL(pageURL)
  if (
    server.protocol !== 'https:' ||
    page.protocol !== 'https:' ||
    server.origin !== page.origin ||
    server.username ||
    server.password ||
    server.search ||
    server.hash
  )
    throw new Error('HLS requires same-origin HTTPS')
  server.pathname = server.pathname.replace(/\/*$/, '') + '/api/media/hls/'
  return server.href
}

export function validHLSOffer(offer: any, mode: HLSMode) {
  return (
    !!offer &&
    offer.version === 1 &&
    offer.backend === 'hls' &&
    offer.mode === mode &&
    offer.path === '/api/media/hls/session' &&
    typeof offer.ticket === 'string' &&
    /^[A-Za-z0-9_-]{32}$/.test(offer.ticket) &&
    Number.isInteger(offer.expires_in_ms) &&
    offer.expires_in_ms > 0 &&
    offer.expires_in_ms <= 10000
  )
}

// the master playlist of the lease the bootstrap POST returned
export function hlsLeaseURL(response: any, mode: HLSMode, base: string) {
  if (
    !response ||
    response.mode !== mode ||
    response.idle_expires_in_ms !== HLS_IDLE_MS ||
    typeof response.master !== 'string' ||
    !/^\/api\/media\/hls\/[A-Za-z0-9_-]{22}\/master\.m3u8$/.test(response.master)
  )
    throw new Error('Invalid HLS bootstrap response')
  return new URL(response.master.slice('/api/media/hls/'.length), base).href
}

// only the negotiated lease path, and on LL-HLS media playlists the two blocking-reload directives
export function scopedHLSURL(candidate: unknown, master: string, mode: HLSMode) {
  if (typeof candidate !== 'string' || /[\\\s]/.test(candidate) || /%|(?:^|\/)\.\.(?:\/|$)/.test(candidate))
    return false
  try {
    const url = new URL(candidate, master)
    const root = new URL('.', master)
    if (
      url.protocol !== 'https:' ||
      url.origin !== root.origin ||
      url.username ||
      url.password ||
      url.hash ||
      !url.pathname.startsWith(root.pathname)
    )
      return false
    const relative = url.pathname.slice(root.pathname.length)
    if (
      !/^(?:master\.m3u8|(?:high|medium|low|audio)\/(?:index\.m3u8|init-[1-9][0-9]*\.mp4|seg-[1-9][0-9]*\.m4s|part-[1-9][0-9]*-[0-5]\.m4s))$/.test(
        relative,
      )
    )
      return false
    const keys = [...url.searchParams.keys()]
    if (!keys.length) return !url.search
    if (
      mode !== 'll-hls' ||
      !relative.endsWith('/index.m3u8') ||
      keys.length > 2 ||
      new Set(keys).size !== keys.length ||
      !keys.includes('_HLS_msn')
    )
      return false
    return keys.every(
      (key) =>
        (key === '_HLS_msn' || key === '_HLS_part') &&
        /^(?:0|[1-9][0-9]*)$/.test(url.searchParams.get(key) || '') &&
        Number.isSafeInteger(Number(url.searchParams.get(key))) &&
        (key !== '_HLS_part' || Number(url.searchParams.get(key)) <= 6),
    )
  } catch {
    return false
  }
}

export function validHLSPlaylist(text: unknown, playlistURL: string, master: string, mode: HLSMode) {
  if (
    typeof text !== 'string' ||
    text.length > HLS_PLAYLIST_BYTES ||
    !text.startsWith('#EXTM3U\n') ||
    /#EXT-X-(?:KEY|SESSION-KEY|CONTENT-STEERING|DEFINE|INTERSTITIAL)/.test(text)
  )
    return false
  for (const line of text.split('\n')) {
    for (const value of playlistURIs(line)) {
      // rendition reports legitimately use ../low within the same lease
      try {
        if (/[%\\\s]/.test(value) || !scopedHLSURL(new URL(value, playlistURL).href, master, mode)) return false
      } catch {
        return false
      }
    }
  }
  return true
}

function playlistURIs(line: string) {
  if (!line.startsWith('#')) return [line.trim()].filter(Boolean)
  const values: string[] = []
  const pattern = /(?:^|[, :])(?:[A-Z-]*URI)="([^"]+)"/g
  let match
  while ((match = pattern.exec(line)) !== null) values.push(match[1])
  return values
}

// statuses after which the lease or the media is gone for good
export function terminalHLSStatus(status: number | undefined) {
  return [400, 401, 403, 404, 410, 413].includes(status as number)
}

// the media playlists a master names: the URLs native playback will load on its own
export function hlsMasterChildren(text: string, master: string, mode: HLSMode) {
  if (!validHLSPlaylist(text, master, master, mode)) throw new Error('Invalid HLS master')
  const children: string[] = []
  for (const line of text.split('\n')) {
    for (const value of playlistURIs(line)) {
      const url = new URL(value, master)
      if (!/(?:high|medium|low|audio)\/index\.m3u8$/.test(url.pathname) || url.search)
        throw new Error('Invalid HLS master child')
      if (!children.includes(url.href)) children.push(url.href)
    }
  }
  if (children.length < 2 || children.length > 4 || !children.some((value) => value.endsWith('/audio/index.m3u8')))
    throw new Error('Invalid HLS master children')
  return children
}

// drop the source and whatever the element buffered (URL or MSE); never for the WebRTC element
export function clearHLSVideo(video: HTMLVideoElement | null | undefined) {
  if (!video) return
  try {
    video.pause()
  } catch {}
  video.srcObject = null
  video.removeAttribute('src')
  try {
    video.load()
  } catch {}
}
