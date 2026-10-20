/**
 * iOS WebKit (Safari and Brave) can keep a page it navigated away from in the
 * back/forward cache with its event WebSocket and WebRTC peer still open. The
 * left page then stays a second connected session that keeps receiving media.
 * Close everything when the page is cached and start fresh when it returns.
 */
export function bindPageLifecycle(target, { leave, restore }) {
  const hide = (event) => { if (event.persisted) leave() }
  const show = (event) => { if (event.persisted) restore() }
  target.addEventListener('pagehide', hide)
  target.addEventListener('pageshow', show)
  return () => {
    target.removeEventListener('pagehide', hide)
    target.removeEventListener('pageshow', show)
  }
}
