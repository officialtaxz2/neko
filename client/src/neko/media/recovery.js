import { RECONNECT_DELAYS_MS, reconnectDelayForAttempt } from '../recovery.js'

export const MEDIA_RETRY_DELAYS_MS = RECONNECT_DELAYS_MS

export function isMediaRetryCloseCode(code) {
  return code === 4413 || code === 4500
}

export function shouldAwaitMediaCloseForEnd(reason) {
  return reason === 'backend_error'
}

export function hasVideoDecodeCapacity(outstanding, pending, cap) {
  return outstanding + pending < cap
}

export function mediaRetryDelayForAttempt(attempt) {
  return reconnectDelayForAttempt(attempt, MEDIA_RETRY_DELAYS_MS)
}
