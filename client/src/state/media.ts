// The fork: the stream's path. WebRTC is the default (and the only path that suits controlling);
// HLS and LL-HLS (docs/HLS_LL_HLS.md) are opt-ins for watching, chosen in the settings or for one
// visit with ?media=hls / ?media=ll-hls. A choice is stored and takes effect with a fresh page load,
// since the transport is fixed for the client's life. The server says which HLS modes it offers;
// only those can be chosen. Ported from the Vue client's media-selection.js (same storage key).
import { createStore } from 'zustand/vanilla'
import { HLSTransport, hlsModes, type HLSMode, type HLSState, type StreamTransport } from '@m1k1o/neko'
import workerPath from 'hls.js/dist/hls.worker.js?url'
import { scoped, type NekoApp } from './app'

export type MediaBackend = 'webrtc' | HLSMode
export const MEDIA_BACKEND_KEY = 'media_backend'

export interface MediaSelection {
  backend: MediaBackend
  // ?media= decides for this visit, over the stored choice
  overridden: boolean
  // ?media= named no exact HLS mode: WebRTC, the safe path
  invalid: boolean
}

export const isHLS = (value: unknown): value is HLSMode => value === 'hls' || value === 'll-hls'

// whatever is stored that is not an HLS mode is WebRTC (also the removed 'webcodecs-ws')
export const normalizeMedia = (value: unknown): MediaBackend => (isHLS(value) ? value : 'webrtc')

export function readMedia(storage: Pick<Storage, 'getItem'> | undefined = globalThis.localStorage): MediaBackend {
  try {
    return normalizeMedia(storage?.getItem(MEDIA_BACKEND_KEY))
  } catch {
    return 'webrtc'
  }
}

export function writeMedia(value: unknown, storage: Pick<Storage, 'setItem'> | undefined = globalThis.localStorage) {
  try {
    if (!storage) return false
    storage.setItem(MEDIA_BACKEND_KEY, normalizeMedia(value))
    return true
  } catch {
    return false
  }
}

// an exact ?media=hls|ll-hls overrides the stored choice without changing it; any other ?media=
// falls back to WebRTC instead of the stored experimental choice
export function resolveMedia(search: string, stored: unknown): MediaSelection {
  const media = new URLSearchParams(search).getAll('media')
  if (!media.length) return { backend: normalizeMedia(stored), overridden: false, invalid: false }
  if (media.length === 1 && isHLS(media[0])) return { backend: media[0], overridden: true, invalid: false }
  return { backend: 'webrtc', overridden: true, invalid: true }
}

// the page for a new choice: the same URL without ?media= (other parameters and a share link's
// fragment stay)
export function mediaNavigationURL(href: string) {
  const url = new URL(href)
  url.searchParams.delete('media')
  return url.toString()
}

// assigning the current URL is only a fragment navigation when it has one (a share link): reload then
export function navigateForMedia(loc: Pick<Location, 'href' | 'assign' | 'reload'> = location) {
  const next = mediaNavigationURL(loc.href)
  if (next === loc.href) loc.reload()
  else loc.assign(next)
}

// the transport for the selection; undefined is the client's default, WebRTC
export function mediaTransport(selection: MediaSelection, autoplay: () => boolean): StreamTransport | undefined {
  if (!isHLS(selection.backend)) return undefined
  return new HLSTransport({ mode: selection.backend, autoplay, workerPath })
}

// per instance: what was selected for this page, what the server offers, how HLS is doing
export const mediaStore = scoped(() =>
  createStore(() => ({
    selection: { backend: 'webrtc', overridden: false, invalid: false } as MediaSelection,
    // the HLS modes the server advertises to this session (none on plain HTTP)
    modes: [] as HLSMode[],
    hls: null as HLSState | null,
  })),
)

// start-up (createNekoApp): ask the server which HLS modes it offers once connected; follow the HLS
// transport's state
export function bindMedia(neko: NekoApp, selection: MediaSelection) {
  const { client } = neko
  const store = mediaStore(neko)
  store.setState({ selection })
  const transport = client.transport
  if (transport instanceof HLSTransport) {
    store.setState({ hls: transport.hls })
    transport.on('hls', (hls) => store.setState({ hls }))
  }
  client.events.on('message', (event, payload) => {
    if (event === 'media/hls/capabilities') store.setState({ modes: hlsModes(payload) })
  })
  client.events.on('connection.status', (status) => {
    if (status === 'disconnected') store.setState({ modes: [] })
    // asking allocates nothing on the server; it answers only for modes it has enabled, and only
    // over HTTPS can a lease work at all
    if (status !== 'connected' || new URL(location.href).protocol !== 'https:') return
    client.send('media/hls/capabilities/request', { version: 1, mode: 'hls' })
    client.send('media/hls/capabilities/request', { version: 1, mode: 'll-hls' })
  })
}

// the viewer picks a path: stored, then a fresh start on it. An HLS mode the server does not offer
// cannot be picked; picking what already runs changes nothing
export function chooseMedia(neko: NekoApp, backend: MediaBackend) {
  const { selection, modes } = mediaStore(neko).getState()
  if (isHLS(backend) && !modes.includes(backend)) return
  writeMedia(backend)
  if (backend === selection.backend && !selection.overridden) return
  reloadForMedia(neko)
}

// back to the stored choice: the page without ?media=
export const clearMediaOverride = (neko: NekoApp) =>
  mediaStore(neko).getState().selection.overridden && reloadForMedia(neko)

// leaving the page on purpose: the socket and the stream close now, so the old page does not
// linger as a second session
export function reloadForMedia(neko: NekoApp) {
  neko.client.disconnect()
  navigateForMedia()
}

// a browser that cannot do WebRTC: this visit over HLS (?media=hls), the stored choice untouched
export function tryHLS(loc: Pick<Location, 'href' | 'assign'> = location) {
  const url = new URL(loc.href)
  url.searchParams.set('media', 'hls')
  loc.assign(url.toString())
}

export function retryHLS(neko: NekoApp) {
  const transport = neko.client.transport
  if (transport instanceof HLSTransport) transport.retry()
}
