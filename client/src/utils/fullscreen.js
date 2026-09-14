export async function elementRequestFullscreen(element) {
  if (!element) return false

  try {
    let result

    if (typeof element.requestFullscreen === 'function') {
      result = element.requestFullscreen()
    } else if (typeof element.webkitRequestFullscreen === 'function') {
      result = element.webkitRequestFullscreen()
    } else if (typeof element.webkitEnterFullscreen === 'function') {
      result = element.webkitEnterFullscreen()
    } else if (typeof element.mozRequestFullScreen === 'function') {
      result = element.mozRequestFullScreen()
    } else if (typeof element.msRequestFullScreen === 'function') {
      result = element.msRequestFullScreen()
    } else {
      return false
    }

    if (result && typeof result.then === 'function') {
      await result
    }
    return true
  } catch {
    return false
  }
}
