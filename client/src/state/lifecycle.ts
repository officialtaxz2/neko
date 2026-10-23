// iOS WebKit (Safari, Brave) can keep a page that was navigated away from in the back/forward cache
// with its websocket and WebRTC peer still open: the left page stays a second connected session that
// keeps receiving media. When the page goes into that cache the connection is closed; when it comes
// back the same session connects again (no new login).
import type { NekoClient } from '@m1k1o/neko'

type PageTarget = Pick<EventTarget, 'addEventListener' | 'removeEventListener'>

export function bindPageLifecycle(client: NekoClient, target: PageTarget | undefined = globalThis): () => void {
  if (typeof target?.addEventListener !== 'function') return () => {}
  let resume = false
  const hide = (e: Event) => {
    if (!(e as PageTransitionEvent).persisted) return
    resume = client.state.connection.status !== 'disconnected'
    if (resume) client.disconnect()
  }
  const show = (e: Event) => {
    if (!(e as PageTransitionEvent).persisted || !resume) return
    resume = false
    if (client.state.authenticated) client.connect()
  }
  target.addEventListener('pagehide', hide)
  target.addEventListener('pageshow', show)
  return () => {
    target.removeEventListener('pagehide', hide)
    target.removeEventListener('pageshow', show)
  }
}
