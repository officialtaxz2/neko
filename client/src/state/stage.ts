// the fork: the picture's element (features/video/Stage registers it), so the room bar and the More
// menu can take it to fullscreen now that the buttons on the picture are gone
import { createStore } from 'zustand/vanilla'
import { scoped, type NekoApp } from './app'

export const stageStore = scoped(() => createStore(() => ({ player: null as HTMLElement | null })))

// within the click that asks for it; iOS only allows the video element itself
export function requestFullscreen(neko: NekoApp) {
  const player = stageStore(neko).getState().player
  if (player?.requestFullscreen) player.requestFullscreen().catch(() => {})
  else (neko.client.transport.element as { webkitEnterFullscreen?: () => void } | null)?.webkitEnterFullscreen?.()
}
