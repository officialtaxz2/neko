// the room bar's buttons: 40px, glass, rounded like the panels' inner corners
export const barButton =
  'relative inline-flex size-10 flex-none cursor-pointer items-center justify-center rounded-inner border border-glass-border bg-glass text-text-normal transition-all duration-300 ease-fluid hover:bg-glass-hover hover:text-pure disabled:cursor-default disabled:opacity-50 aria-expanded:border-mint/40 aria-expanded:text-mint'
// the phone dock's buttons: 44px for a finger
export const dockButton =
  'relative inline-flex size-11 flex-none cursor-pointer items-center justify-center rounded-control border border-glass-border bg-glass text-text-normal transition-all duration-300 ease-fluid hover:bg-glass-hover hover:text-pure disabled:cursor-default disabled:opacity-50 aria-expanded:border-mint/40 aria-expanded:text-mint'

// iOS ignores a page's volume changes; there only muting works, so no slider is offered
export const volumeAdjustable = (() => {
  if (typeof navigator === 'undefined') return true
  const ua = navigator.userAgent || ''
  return !(/iPad|iPhone|iPod/.test(ua) || (navigator.maxTouchPoints > 1 && /Macintosh/.test(ua)))
})()
