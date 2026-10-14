// Single touch heuristic for the client. "Force touch" in the settings makes
// it true on any device. The width fallback keeps the fork's long-standing
// behavior that narrow windows get the touch input path.
export function isTouchDevice(forceTouch: boolean): boolean {
  if (typeof window === 'undefined') return false
  if (forceTouch) return true
  return (
    'ontouchstart' in window ||
    navigator.maxTouchPoints > 0 ||
    /Mobi|Android|iPhone|iPad|iPod|Windows Phone/i.test(navigator.userAgent) ||
    (navigator.maxTouchPoints > 1 && /Macintosh|MacIntel/i.test(navigator.userAgent || navigator.platform)) ||
    (window.matchMedia && window.matchMedia('(pointer: coarse)').matches) ||
    (window.matchMedia && window.matchMedia('(any-pointer: coarse)').matches) ||
    window.innerWidth <= 1024
  )
}
