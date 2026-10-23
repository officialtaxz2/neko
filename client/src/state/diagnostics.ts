// The fork: a read-only handle for the console traces an operator pastes on an affected device
// (deploy/trace-media-browser.js, deploy/trace-hls-browser.js), in place of the Vue client's
// internals they used to read. It hands out states, the media element and the WebRTC peer (for
// getStats()); never a token, a URL or a way to act.
import { WebRTCTransport } from '@m1k1o/neko'
import type { NekoApp } from './app'
import { mediaStore } from './media'

declare global {
  interface Window {
    __nekoTrace?: Readonly<{
      backend: () => string
      connection: () => string
      video: () => HTMLVideoElement | null
      peer: () => RTCPeerConnection | null
      hls: () => { mode: string; status: string; detail: string } | null
    }>
  }
}

export function exposeDiagnostics(neko: NekoApp) {
  const { client } = neko
  const media = mediaStore(neko)
  const transport = client.transport
  const handle = Object.freeze({
    backend: () => media.getState().selection.backend,
    connection: () => client.state.connection.status,
    video: () => transport.element as HTMLVideoElement | null,
    peer: () => (transport instanceof WebRTCTransport ? transport.peer : null),
    hls: () => {
      const hls = media.getState().hls
      return hls && { mode: hls.mode, status: hls.status, detail: hls.detail }
    },
  })
  try {
    Object.defineProperty(window, '__nekoTrace', { value: handle, configurable: true })
  } catch {}
}
