// the fork: the picture's element (features/video/Stage registers it), so the room bar and the More
// menu can take it to fullscreen now that the buttons on the picture are gone. Fullscreen tries what
// the browser has, as the Vue client did (utils/fullscreen.js): the standard call, then WebKit's
// prefixed one (Safari before 16.4, older iPads, TV browsers), then the video element alone (the
// iPhone's own player, for a viewer who only watches) and last the page filled with the picture and
// an exit button (an iPhone in control: its player would take the input away).
import { createStore } from 'zustand/vanilla'
import { selectControlling } from '@m1k1o/neko'
import { scoped, type NekoApp } from './app'

export const stageStore = scoped(() =>
  createStore(() => ({
    player: null as HTMLElement | null,
    // the page-filling fallback is on
    fallback: false,
  })),
)

type WebKitElement = HTMLElement & { webkitRequestFullscreen?: () => Promise<void> | void }
type WebKitVideo = HTMLVideoElement & { webkitEnterFullscreen?: () => void }
type WebKitDocument = Document & { webkitFullscreenElement?: Element | null }

// the element in fullscreen, also where only the prefixed API exists
export const fullscreenElement = (doc: Document = document) =>
  doc.fullscreenElement ?? (doc as WebKitDocument).webkitFullscreenElement ?? null

// the element's own fullscreen, standard or prefixed; false when the browser has neither or refuses
export async function elementFullscreen(el: HTMLElement | null): Promise<boolean> {
  if (!el) return false
  const w = el as WebKitElement
  try {
    if (typeof w.requestFullscreen === 'function') await w.requestFullscreen()
    else if (typeof w.webkitRequestFullscreen === 'function') await w.webkitRequestFullscreen()
    else return false
    return true
  } catch {
    return false
  }
}

export const setFallback = (neko: NekoApp, on: boolean) => stageStore(neko).setState({ fallback: on })

// Within the click that asks for it: the iPhone's player must be asked before anything is awaited,
// since iOS only allows it in the gesture itself.
export async function requestFullscreen(neko: NekoApp) {
  const stage = stageStore(neko)
  if (stage.getState().fallback) return setFallback(neko, false)
  const { player } = stage.getState()
  const video = neko.client.transport.element as WebKitVideo | null
  const watching = !selectControlling(neko.client.state)
  const elementApi =
    !!player && (typeof player.requestFullscreen === 'function' || !!(player as WebKitElement).webkitRequestFullscreen)
  if (!elementApi && watching && typeof video?.webkitEnterFullscreen === 'function') {
    try {
      video.webkitEnterFullscreen()
      return
    } catch {}
  }
  if (await elementFullscreen(player)) return
  setFallback(neko, true)
}
