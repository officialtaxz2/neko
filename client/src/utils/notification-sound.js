// Optional notification audio must never interrupt room-event processing.
export function playNotificationSound() {
  try {
    if (typeof Audio !== 'function') return
    const playback = new Audio('chat.mp3').play()
    // Older media implementations can return undefined rather than a Promise.
    if (playback && typeof playback.catch === 'function') playback.catch(() => {})
  } catch (_) {
    // Unsupported media, synchronous errors and blocked autoplay are optional.
  }
}
