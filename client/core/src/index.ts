// @m1k1o/neko: everything needed to talk to a neko server, usable from any framework.
// The GUI in ../src imports only from here.
export {
  NekoClient,
  selectSession,
  selectControlling,
  selectIsAdmin,
  type NekoClientOptions,
  type Pos,
} from './client.ts'
export { Overlay } from './overlay.ts'
export type { TrackpadCursor } from './overlay.ts' // the fork
export { mount } from './mount.ts'
export { Emitter } from './emitter.ts'
export { NekoApi, ApiError } from './api.ts'
export type {
  StreamTransport,
  SessionInfo,
  InputChannel,
  TransportKind,
  TransportStatus,
  TransportState,
  TransportEvents,
} from './transport.ts'
export { WebRTCTransport } from './transport/webrtc.ts'
export {
  HLSTransport,
  hlsModes,
  type HLSMode,
  type HLSStatus,
  type HLSState,
  type HLSTransportEvents,
  type HLSTransportOptions,
} from './transport/hls.ts'
export { DataChannelInput } from './input/datachannel.ts'
export { samplePlayback } from './playback.ts' // the fork: the GUI's ?debug view
export { WebSocketInput } from './input/websocket.ts'
export * from './types.ts'
