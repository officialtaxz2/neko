export function makeid(length: number) {
  let result = ''
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const charactersLength = characters.length
  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength))
  }
  return result
}

export function lockKeyboard() {
  if (navigator && navigator.keyboard) {
    navigator.keyboard.lock()
  }
}

export function unlockKeyboard() {
  if (navigator && navigator.keyboard) {
    navigator.keyboard.unlock()
  }
}

export function elementRequestFullscreen(el: HTMLElement) {
  if (typeof el.requestFullscreen === 'function') {
    el.requestFullscreen()
    //@ts-ignore
  } else if (typeof el.webkitRequestFullscreen === 'function') {
    //@ts-ignore
    el.webkitRequestFullscreen()
    //@ts-ignore
  } else if (typeof el.webkitEnterFullscreen === 'function') {
    //@ts-ignore
    el.webkitEnterFullscreen()
    //@ts-ignore
  } else if (typeof el.mozRequestFullScreen === 'function') {
    //@ts-ignore
    el.mozRequestFullScreen()
    //@ts-ignore
  } else if (typeof el.msRequestFullScreen === 'function') {
    //@ts-ignore
    el.msRequestFullScreen()
  } else {
    return false
  }
  return true
}

export function isFullscreen(): boolean {
  return (
    document.fullscreenElement ||
    //@ts-ignore
    document.msFullscreenElement ||
    //@ts-ignore
    document.mozFullScreenElement ||
    //@ts-ignore
    document.webkitFullscreenElement
  )
}

export function onFullscreenChange(el: HTMLElement, fn: () => void) {
  // Use addEventListener for reliable cross-browser fullscreen detection.
  // The old property-check approach (el.onfullscreenchange === null) could
  // silently fail when the property was undefined instead of null.
  const events = ['fullscreenchange', 'webkitfullscreenchange', 'mozfullscreenchange', 'MSFullscreenChange']
  for (const event of events) {
    el.addEventListener(event, fn)
  }
}
