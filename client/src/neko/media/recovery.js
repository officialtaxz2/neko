import { RECONNECT_DELAYS_MS, reconnectDelayForAttempt } from '../recovery.js'

export const MEDIA_RETRY_DELAYS_MS = RECONNECT_DELAYS_MS
export const MEDIA_RETRY_STABILITY_MS = 30_000

export function isMediaRetryCloseCode(code) {
  return code === 4413 || code === 4500
}

export function shouldAwaitMediaCloseForEnd(reason) {
  return reason === 'backend_error'
}

export function shouldDropDecodedVideoOutput(outstanding, cap) {
  return outstanding >= cap
}

export function mediaRetryDelayForAttempt(attempt) {
  return reconnectDelayForAttempt(attempt, MEDIA_RETRY_DELAYS_MS)
}
